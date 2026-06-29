package config

import "time"

type Config interface {
	HTTPPort() string
	HTTPTimeout() time.Duration
	TokenDuration() time.Duration
	TokenSecret() string
	MetricsTimeout() time.Duration
	MetricsPort() string
	AddSource() bool
	LoggerLevel() string
	StorageConnStr() string
	GracefullShutdownTimeout() time.Duration
	LifeIdleConns() time.Duration
	MaxOpenConns() int
	MaxIdleConns() int
	LifeConns() time.Duration
}
