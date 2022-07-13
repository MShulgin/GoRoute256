package cache

import (
	"context"
	"errors"
	redis "github.com/go-redis/redis/v8"
)

type RedisCache[V any] struct {
	rdb          *redis.Client
	valueDecoder func(string) (*V, error)
	valueEncoder func(*V) (string, error)
}

func NewRedisCache[V any](server string,
	valueDecoder func(string) (*V, error),
	valueEncoder func(*V) (string, error)) (*RedisCache[V], error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     server,
		Password: "",
		DB:       0,
	})
	if err := rdb.Ping(context.TODO()).Err(); err != nil {
		return nil, err
	}
	c := RedisCache[V]{
		rdb:          rdb,
		valueDecoder: valueDecoder,
		valueEncoder: valueEncoder,
	}
	return &c, nil
}

func (r RedisCache[V]) Get(id string) (*V, error) {
	value, err := r.rdb.Get(context.TODO(), id).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, CacheMissError{}
		}
		return nil, err
	}

	return r.valueDecoder(value)
}

func (r RedisCache[V]) Set(id string, value *V) error {
	strValue, err := r.valueEncoder(value)
	if err != nil {
		return err
	}
	if err = r.rdb.Set(context.TODO(), id, strValue, 0).Err(); err != nil {
		return err
	}

	return nil
}

func (r RedisCache[V]) Invalidate(id string) error {
	if err := r.rdb.Del(context.TODO(), id).Err(); err != nil {
		return err
	}
	return nil
}
