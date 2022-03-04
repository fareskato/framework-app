package cache

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
)

type RedisCache struct {
	Conn   *redis.Pool
	Prefix string
}

func (c *RedisCache) Has(k string) (bool, error) {
	// prepend the prefix to the key
	key := fmt.Sprintf("%s:%s", c.Prefix, k)
	// connect to redis: get the connection pool
	conn := c.Conn.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (c *RedisCache) Get(k string) (interface{}, error) {
	// prepend the prefix to the key
	key := fmt.Sprintf("%s:%s", c.Prefix, k)
	// connect to redis: get the connection pool
	conn := c.Conn.Get()
	defer conn.Close()
	// get from redis
	cacheEntry, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	// decode
	decodedData, err := decodeCache(string(cacheEntry))
	if err != nil {
		return nil, err
	}
	item := decodedData[key]
	return item, nil
}

func (c *RedisCache) Set(k string, val interface{}, expires ...int) error {
	// prepend the prefix to the key
	key := fmt.Sprintf("%s:%s", c.Prefix, k)
	// connect to redis: get the connection pool
	conn := c.Conn.Get()
	defer conn.Close()

	entry := EntryCache{}
	entry[key] = val

	// encode to store
	encodedData, err := encodeCache(entry)
	if err != nil {
		return err
	}
	// store in the cache
	if len(expires) > 0 {
		_, err := conn.Do("SETEX", key, expires[0], string(encodedData))
		if err != nil {
			return err
		}
	} else {
		_, err := conn.Do("SET", key, string(encodedData))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *RedisCache) Forget(k string) error {
	// prepend the prefix to the key
	key := fmt.Sprintf("%s:%s", c.Prefix, k)
	// connect to redis: get the connection pool
	conn := c.Conn.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	if err != nil {
		return err
	}
	return nil
}

func (c *RedisCache) EmptyByMatch(k string) error {
	// prepend the prefix to the key
	key := fmt.Sprintf("%s:%s", c.Prefix, k)
	// connect to redis: get the connection pool
	conn := c.Conn.Get()
	defer conn.Close()
	// fetch all matches
	keys, err := c.getKeys(key)
	if err != nil {
		return err
	}
	// empty all matched keys
	for _, x := range keys {
		_, err := conn.Do("DEL", x)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *RedisCache) Empty() error {
	// prepend the prefix to the key
	key := fmt.Sprintf("%s:", c.Prefix)
	// connect to redis: get the connection pool
	conn := c.Conn.Get()
	defer conn.Close()
	// fetch all
	keys, err := c.getKeys(key)
	if err != nil {
		return err
	}
	for _, x := range keys {
		_, err := conn.Do("DEL", x)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *RedisCache) getKeys(pattern string) ([]string, error) {
	// connect to redis: get the connection pool
	conn := c.Conn.Get()
	defer conn.Close()
	idx := 0
	keys := make([]string, 0)
	for {
		rep, err := redis.Values(conn.Do("SCAN", idx, "MATCH", fmt.Sprintf("%s*", pattern)))
		if err != nil {
			return keys, err
		}
		idx, _ = redis.Int(rep[0], nil)
		t, _ := redis.Strings(rep[1], nil)
		keys = append(keys, t...)

		if idx == 0 {
			break
		}
	}
	return keys, nil
}
