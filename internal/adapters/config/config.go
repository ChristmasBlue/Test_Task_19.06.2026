package config

import (
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	*koanf.Koanf
}

func NewConfig(path string) *Config {
	k := koanf.New(".")
	err := k.Load(file.Provider(path), yaml.Parser())
	if err != nil {
		panic(err)
	}

	return &Config{
		Koanf: k,
	}
}

func (c *Config) HTTPPort() string {
	return c.Koanf.String("http.port")
}

func (c *Config) HTTPTimeout() time.Duration {
	return c.Duration("http.timeout")
}

func (c *Config) StorageConnStr() string {
	return c.Koanf.String("storage.connection_string")
}

func (c *Config) LoggerLevel() string {
	return c.Koanf.String("logger.level")
}

func (c *Config) AddSource() bool {
	return c.Bool("logger.add_source")
}

func (c *Config) MetricsPort() string {
	return c.Koanf.String("metrics.port")
}

func (c *Config) MetricsTimeout() time.Duration {
	return c.Duration("metrics.timeout")
}

func (c *Config) TokenSecret() string {
	return c.Koanf.String("token.secret")
}

func (c *Config) TokenDuration() time.Duration {
	return c.Duration("token.duration")
}

func (c *Config) GracefullShutdownTimeout() time.Duration {
	return c.Duration("gracefull_shutdown.timeout")
}

func (c *Config) LifeIdleConns() time.Duration {
	return c.Duration("storage.time_life_idle_conns")
}

func (c *Config) MaxOpenConns() int {
	return c.Koanf.Int("storage.max_conns")
}

func (c *Config) MaxIdleConns() int {
	return c.Koanf.Int("storage.max_idle_conns")
}

func (c *Config) LifeConns() time.Duration {
	return c.Duration("storage.time_life_conns")
}

func (c *Config) CacheHost() string {
	return c.Koanf.String("cache.host")
}

func (c *Config) CachePort() int {
	return c.Koanf.Int("cache.port")
}

func (c *Config) CacheTTL() time.Duration {
	return c.Duration("cache.ttl")
}

func (c *Config) CachePass() string {
	return c.Koanf.String("cache.password")
}

func (c *Config) CacheDB() int {
	return c.Koanf.Int("cache.db")
}

func (c *Config) RateLimiterRate() float64 {
	return c.Koanf.Float64("limiter.rate")
}

func (c *Config) RateLimiterBurst() int {
	return c.Koanf.Int("limiter.burst")
}
