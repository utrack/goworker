package goworker

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/cihub/seelog"
	"github.com/youtube/vitess/go/pools"
	"golang.org/x/net/context"
)

var (
	logger seelog.LoggerInterface
	pool   *pools.ResourcePool
	ctx    context.Context
)

// Init initializes the goworker process. This will be
// called by the Work function, but may be used by programs
// that wish to access goworker functions and configuration
// without actually processing jobs.
func Init(set *PoolPrefs) error {
	var err error
	logger, err = seelog.LoggerFromWriterWithMinLevel(os.Stdout, seelog.InfoLvl)
	if err != nil {
		return err
	}

	setDefaults(set)
	ctx = context.Background()

	pool = newRedisPool(set.Redis, set.MaxConns, set.MaxConns, time.Minute)

	return nil
}

// setDefaults fills in the blanks in PoolPrefs
// with default values.
func setDefaults(sets *PoolPrefs) {
	// Set defaults
	if sets.Concurrency == 0 {
		sets.Concurrency = 25
	}
	if sets.MaxConns == 0 {
		sets.MaxConns = 2
	}
	if sets.Redis == "" {
		sets.Redis = "redis://localhost:6379/1"
	}
	if sets.RedisNamespace == "" {
		sets.RedisNamespace = "resque:"
	}
	if sets.SleepInterval == 0 {
		sets.SleepInterval = time.Second * 5
	}

	// Add registered queues to the list.
	if sets.UseRegistered {
		sets.Queues = make([]string, 0, len(workers))
		for key, _ := range workers {
			sets.Queues = append(sets.Queues, key)
		}
	}
}

// GetConn returns a connection from the goworker Redis
// connection pool. When using the pool, check in
// connections as quickly as possible, because holding a
// connection will cause concurrent worker functions to lock
// while they wait for an available connection. Expect this
// API to change drastically.
func GetConn() (*RedisConn, error) {
	resource, err := pool.Get(ctx)

	if err != nil {
		return nil, err
	}
	return resource.(*RedisConn), nil
}

// PutConn puts a connection back into the connection pool.
// Run this as soon as you finish using a connection that
// you got from GetConn. Expect this API to change
// drastically.
func PutConn(conn *RedisConn) {
	pool.Put(conn)
}

// Close cleans up resources initialized by goworker. This
// will be called by Work when cleaning up. However, if you
// are using the Init function to access goworker functions
// and configuration without processing jobs by calling
// Work, you should run this function when cleaning up. For
// example,
//
//	if err := goworker.Init(); err != nil {
//		fmt.Println("Error:", err)
//	}
//	defer goworker.Close()
func Close() {
	pool.Close()
}

// Work starts the goworker process. Check for errors in
// the return value. Work will take over the Go executable
// and will run until a QUIT, INT, or TERM signal is
// received, or until the queues are empty if the
// -exit-on-complete flag is set.
func Work(set PoolPrefs) error {
	err := Init(&set)
	if err != nil {
		return err
	}
	defer Close()

	quit := signals()

	poller, err := newPoller(set)
	if err != nil {
		return err
	}
	jobs := poller.poll(time.Duration(set.SleepInterval), quit)

	var monitor sync.WaitGroup

	for id := 0; id < set.Concurrency; id++ {
		worker, err := newWorker(strconv.Itoa(id), set)
		if err != nil {
			return err
		}
		worker.work(jobs, &monitor)
	}

	monitor.Wait()

	return nil
}
