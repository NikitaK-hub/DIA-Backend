package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	servicePrefix = "DIA_Backend"
	jwtPrefix     = "jwt-blacklist"
)

type Config struct {
	Host        string
	Password    string
	Port        int
	User        string
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

type Client struct {
	cfg    *Config
	client *redis.Client
}

func New(cfg *Config) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:    cfg.Password,
		DB:          0,
		DialTimeout: cfg.DialTimeout,
		ReadTimeout: cfg.ReadTimeout,
	})

	return &Client{
		cfg:    cfg,
		client: rdb,
	}
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) WriteJWTToBlacklist(ctx context.Context, token string, expiresAt time.Time) error {
	key := fmt.Sprintf("%s.%s:%s", servicePrefix, jwtPrefix, token)
	expiration := time.Until(expiresAt)

	if expiration <= 0 {
		return fmt.Errorf("token already expired")
	}

	err := c.client.Set(ctx, key, "blacklisted", expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to write JWT to blacklist: %v", err)
	}

	return nil
}

func (c *Client) CheckJWTInBlacklist(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("%s.%s:%s", servicePrefix, jwtPrefix, token)

	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check JWT in blacklist: %v", err)
	}

	return exists > 0, nil
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx).Result()
	return err
}
