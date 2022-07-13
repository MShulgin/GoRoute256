package cache

type Cache[V any] interface {
	Get(id string) (*V, error)
	Set(id string, value *V) error
	Invalidate(id string) error
}

type CacheMissError struct{}

func (e CacheMissError) Error() string {
	return "cache miss"
}
