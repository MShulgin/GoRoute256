package cache

import (
	"errors"
	"github.com/bradfitz/gomemcache/memcache"
)

type MemCache[V any] struct {
	client       *memcache.Client
	valueDecoder func([]byte) (*V, error)
	valueEncoder func(*V) ([]byte, error)
}

func NewMemCache[V any](servers []string,
	valueDecoder func([]byte) (*V, error),
	valueEncoder func(*V) ([]byte, error)) (*MemCache[V], error) {

	client := memcache.New(servers...)
	err := client.Ping()
	if err != nil {
		return nil, err
	}
	mc := MemCache[V]{
		client:       client,
		valueDecoder: valueDecoder,
		valueEncoder: valueEncoder,
	}
	return &mc, nil
}

func (m *MemCache[V]) Get(id string) (*V, error) {
	item, err := m.client.Get(id)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return nil, CacheMissError{}
		}
		return nil, err
	}

	value, err := m.valueDecoder(item.Value)

	if err != nil {
		return nil, err
	}
	return value, nil
}

func (m *MemCache[V]) Set(id string, value *V) error {
	cacheValue, err := m.valueEncoder(value)
	if err != nil {
		return err
	}
	cacheItem := memcache.Item{Key: id, Value: cacheValue}
	if err := m.client.Set(&cacheItem); err != nil {
		return err
	}
	return nil
}

func (m *MemCache[V]) Invalidate(id string) error {
	if err := m.client.Delete(id); err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return nil
		} else {
			return err
		}
	}
	return nil
}
