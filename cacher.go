package main

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/vmihailenco/msgpack"
)

// Cacher - represents a cacher
type Cacher struct {
	redis *redis.Client
}

// NewCacher - initialize a new cacher
func NewCacher(redisaddr string) (*Cacher, error) {
	opts, err := redis.ParseURL(redisaddr)
	if err != nil {
		return nil, err
	}

	c := new(Cacher)
	c.redis = redis.NewClient(opts)

	if _, err := c.redis.Ping().Result(); err != nil {
		return nil, err
	}

	return c, nil
}

// Put - put a new item into cache
func (c *Cacher) Put(k string, v interface{}, ttl int64, tags []string) {
	k = "sqler:cache:value:" + k
	data, _ := msgpack.Marshal(v)

	c.redis.Set(k, string(data), time.Duration(ttl)*time.Second)

	for _, tag := range tags {
		tag = "sqler:cache:tag:" + tag
		c.redis.SAdd(tag, k)
	}
}

// Get - fetch the data of the specified key
func (c *Cacher) Get(k string) interface{} {
	k = "sqler:cache:value:" + k
	if c.redis.Exists(k).Val() < 1 {
		return nil
	}

	encodedVal := c.redis.Get(k).Val()

	var data interface{}

	msgpack.Unmarshal([]byte(encodedVal), &data)

	return data
}

// ClearTagged - clear cached data tagged with the specified tag
func (c *Cacher) ClearTagged(tag string) {
	tag = "sqler:cache:tag:" + tag

	for _, k := range c.redis.SMembers(tag).Val() {
		c.redis.Del(k)
	}
}
