package store

import (
	"context"
	"encoding/json"
	"fmt"
	"service/internal/model"
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

func (c *CachedStore) Create(ctx context.Context, title, content string) (model.Note, error) {
	note, err := c.next.Create(ctx, title, content)
	if err != nil {
		return model.Note{}, err
	}

	c.invalidate(ctx, note.ID)
	return note, nil

}

func (c *CachedStore) GetAll(ctx context.Context, limit, offset int) ([]model.Note, int, error) {
	return c.next.GetAll(ctx, limit, offset)
}

func (c *CachedStore) GetByID(ctx context.Context, id int) (model.Note, error) {
	cachedKey := fmt.Sprintf("note:%d", id)

	data, err := c.redis.Get(ctx, cachedKey).Bytes()

	if err == nil {
		var note model.Note
		if err := json.Unmarshal(data, &note); err == nil {
			return note, nil
		}
	}

	note, err := c.next.GetByID(ctx, id)

	if err != nil {
		return model.Note{}, err
	}

	if jsonData, err := json.Marshal(note); err == nil {
		c.redis.Set(ctx, cachedKey, jsonData, c.ttl)
	}

	return note, nil
}

func (c *CachedStore) invalidate(ctx context.Context, id int) {
	c.redis.Del(ctx, "notes:all", fmt.Sprintf("notes:%d", id))
}
