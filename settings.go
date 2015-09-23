package goworker

type PoolPrefs struct {
	// Queues is a list of Resque queues to scan.
	Queues []string
	// UseRegistered shows if registered jobs should be used as queues.
	UseRegistered bool
	// IsStrict shows if queue names are specified in strict format.
	IsStrict bool

	// SleepInterval is an interval between scans when no jobs are found.
	SleepInterval float64
	// Concurrency is a max number of concurrently executing jobs.
	Concurrency int

	// Redis is a Redis URI.
	Redis string
	// MaxConns is a max number of Redis connections.
	MaxConns int
	// RedisNamespace is a namespace of Redis keys.
	// Default: resque:
	RedisNamespace string

	// ExitOnComplete indicates if worker should quit after completing jobs.
	ExitOnComplete bool

	// UseNumber shows if json.Number should be used instead of float64.
	UseNumber bool
}
