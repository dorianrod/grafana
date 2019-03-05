package distcache

import (
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/grafana/grafana/pkg/setting"
)

type memcachedStorage struct {
	c *memcache.Client
}

func newMemcachedStorage(opts *setting.CacheOpts) *memcachedStorage {
	return &memcachedStorage{
		c: memcache.New(opts.ConnStr),
	}
}

func newItem(sid string, data []byte, expire int32) *memcache.Item {
	return &memcache.Item{
		Key:        sid,
		Value:      data,
		Expiration: expire,
	}
}

// Set sets value to given key in the cache.
func (s *memcachedStorage) Set(key string, val interface{}, expires time.Duration) error {
	item := &cachedItem{Val: val}
	bytes, err := encodeGob(item)
	if err != nil {
		return err
	}

	memcachedItem := newItem(key, bytes, int32(expires))
	return s.c.Set(memcachedItem)
}

// Get gets value by given key in the cache.
func (s *memcachedStorage) Get(key string) (interface{}, error) {
	memcachedItem, err := s.c.Get(key)
	if err != nil && err.Error() == "memcache: cache miss" {
		return nil, ErrCacheItemNotFound
	}

	if err != nil {
		return nil, err
	}

	item := &cachedItem{}

	err = decodeGob(memcachedItem.Value, item)
	if err != nil {
		return nil, err
	}

	return item.Val, nil
}

// Delete delete a key from the cache
func (s *memcachedStorage) Delete(key string) error {
	return s.c.Delete(key)
}
