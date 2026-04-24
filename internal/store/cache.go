package store

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type CachedStore struct {
	next  NoteStorer
	redis *redis.Client
	ttl   time.Duration
}

func NewCachedStore(next NoteStorer, rdb *redis.Client) *CachedStore {
	return &CachedStore{
		next:  next,
		redis: rdb,
		ttl:   1 * time.Minute,
	}
}
