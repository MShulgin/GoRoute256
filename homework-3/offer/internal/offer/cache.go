package offer

import (
	"encoding/json"
	"errors"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/cache"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/config"
	"strings"
)

func NewCacheFromConfig(conf config.CacheConfig) (cache.Cache[Offer], error) {
	switch strings.ToLower(conf.Type) {
	case "redis":
		return NewRedisCache(conf.Addr)
	case "memcache":
		return NewMemcache(conf.Addr)
	default:
		return nil, errors.New("unknown cache: " + conf.Type)
	}
}

func NewMemcache(memcacheServer string) (cache.Cache[Offer], error) {
	offerDecoder := func(d []byte) (*Offer, error) {
		var o Offer
		if err := json.Unmarshal(d, &o); err != nil {
			return nil, err
		}
		return &o, nil
	}
	offerEncoder := func(offer *Offer) ([]byte, error) {
		return json.Marshal(offer)
	}
	return cache.NewMemCache[Offer]([]string{memcacheServer}, offerDecoder, offerEncoder)
}

func NewRedisCache(server string) (cache.Cache[Offer], error) {
	offerDecoder := func(d string) (*Offer, error) {
		var o Offer
		if err := json.Unmarshal([]byte(d), &o); err != nil {
			return nil, err
		}
		return &o, nil
	}
	offerEncoder := func(offer *Offer) (string, error) {
		b, err := json.Marshal(offer)
		if err != nil {
			return "", nil
		}
		return string(b), nil
	}
	return cache.NewRedisCache[Offer](server, offerDecoder, offerEncoder)
}
