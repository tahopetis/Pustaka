package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type RedisDB struct {
	Client *redis.Client
}

func NewRedisDB(redisURL, password string, poolSize int) (*RedisDB, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse Redis URL: %w", err)
	}

	if password != "" {
		opt.Password = password
	}

	if poolSize > 0 {
		opt.PoolSize = poolSize
	}

	// Set additional options
	opt.MinIdleConns = 5
	opt.MaxRetries = 3
	opt.DialTimeout = 5 * time.Second
	opt.ReadTimeout = 3 * time.Second
	opt.WriteTimeout = 3 * time.Second
	opt.PoolTimeout = 4 * time.Second

	client := redis.NewClient(opt)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Redis: %w", err)
	}

	log.Info().
		Str("redis_url", redisURL).
		Int("pool_size", poolSize).
		Msg("Connected to Redis")

	return &RedisDB{Client: client}, nil
}

func (db *RedisDB) Close() error {
	if db.Client != nil {
		err := db.Client.Close()
		if err != nil {
			return fmt.Errorf("error closing Redis client: %w", err)
		}
		log.Info().Msg("Redis client closed")
	}
	return nil
}

func (db *RedisDB) Health(ctx context.Context) error {
	return db.Client.Ping(ctx).Err()
}

// Set stores a key-value pair with expiration
func (db *RedisDB) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return db.Client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (db *RedisDB) Get(ctx context.Context, key string) (string, error) {
	return db.Client.Get(ctx, key).Result()
}

// Del deletes one or more keys
func (db *RedisDB) Del(ctx context.Context, keys ...string) error {
	return db.Client.Del(ctx, keys...).Err()
}

// Exists checks if one or more keys exist
func (db *RedisDB) Exists(ctx context.Context, keys ...string) (int64, error) {
	return db.Client.Exists(ctx, keys...).Result()
}

// Expire sets a timeout on a key
func (db *RedisDB) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return db.Client.Expire(ctx, key, expiration).Err()
}

// HSet stores a hash field
func (db *RedisDB) HSet(ctx context.Context, key string, values ...interface{}) error {
	return db.Client.HSet(ctx, key, values...).Err()
}

// HGet retrieves a hash field
func (db *RedisDB) HGet(ctx context.Context, key, field string) (string, error) {
	return db.Client.HGet(ctx, key, field).Result()
}

// HGetAll retrieves all hash fields and values
func (db *RedisDB) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return db.Client.HGetAll(ctx, key).Result()
}

// HDel deletes one or more hash fields
func (db *RedisDB) HDel(ctx context.Context, key string, fields ...string) error {
	return db.Client.HDel(ctx, key, fields...).Err()
}

// ZAdd adds a member to a sorted set
func (db *RedisDB) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return db.Client.ZAdd(ctx, key, members...).Err()
}

// ZRange retrieves a range of members from a sorted set
func (db *RedisDB) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return db.Client.ZRange(ctx, key, start, stop).Result()
}

// ZRem removes one or more members from a sorted set
func (db *RedisDB) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return db.Client.ZRem(ctx, key, members...).Err()
}

// Incr increments a numeric value
func (db *RedisDB) Incr(ctx context.Context, key string) (int64, error) {
	return db.Client.Incr(ctx, key).Result()
}

// Decr decrements a numeric value
func (db *RedisDB) Decr(ctx context.Context, key string) (int64, error) {
	return db.Client.Decr(ctx, key).Result()
}

// FlushAll clears all keys from the database
func (db *RedisDB) FlushAll(ctx context.Context) error {
	return db.Client.FlushAll(ctx).Err()
}