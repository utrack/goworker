package goworker

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/pquerna/ffjson/ffjson"
)

// Client can be used to enqueue jobs to Resque.
type Client struct {
	pool   *redis.Pool
	prefix string
}

// NewClient initiates and returns the Client.
func NewClient(pool *redis.Pool, namespace string) *Client {
	return &Client{pool: pool, prefix: namespace}
}

// NewClientFromPrefs initiates and returns the Client
// using settings from PoolPrefs.
func NewClientFromPrefs(pool *redis.Pool, prefs PoolPrefs) *Client {
	return &Client{pool: pool, prefix: prefs.RedisNamespace}
}

// Enqueue adds job to the queue.
func (c *Client) Enqueue(queue string, data interface{}) error {
	buf, err := ffjson.Marshal(data)
	if err != nil {
		return err
	}
	rc := c.pool.Get()
	defer rc.Close()
	defer ffjson.Pool(buf)

	_, err = rc.Do("RPUSH", fmt.Sprintf("%vqueue:%v", c.prefix, queue), buf)
	return err

}
