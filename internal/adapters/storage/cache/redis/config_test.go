package redis_test

import "time"

type testConfig struct {
	connStr string
}

func (c *testConfig) HTTPPort() string              { return "8181" }
func (c *testConfig) HTTPTimeout() time.Duration    { return 5 * time.Second }
func (c *testConfig) TokenDuration() time.Duration  { return 15 * time.Minute }
func (c *testConfig) TokenSecret() string           { return "secret" }
func (c *testConfig) MetricsTimeout() time.Duration { return 5 * time.Second }
func (c *testConfig) MetricsPort() string           { return "8383" }
func (c *testConfig) AddSource() bool               { return false }
func (c *testConfig) LoggerLevel() string           { return "debug" }
func (c *testConfig) StorageConnStr() string {
	return "root:1124@tcp(localhost:3306)/taskmanager?parseTime=true&charset=utf8mb4"
}
func (c *testConfig) GracefullShutdownTimeout() time.Duration { return 5 * time.Second }
func (c *testConfig) LifeIdleConns() time.Duration            { return 5 * time.Minute }
func (c *testConfig) MaxOpenConns() int                       { return 5 }
func (c *testConfig) MaxIdleConns() int                       { return 2 }
func (c *testConfig) LifeConns() time.Duration                { return 5 * time.Minute }
func (c *testConfig) CacheHost() string                       { return "6379" }
func (c *testConfig) CachePort() string                       { return "localhost" }
func (c *testConfig) CacheTTL() time.Duration                 { return 5 * time.Minute }
func (c *testConfig) CachePass() string                       { return "" }
func (c *testConfig) CacheDB() int                            { return 0 }
func (c *testConfig) RateLimiterRate() float64                { return 100 }
func (c *testConfig) RateLimiterBurst() int                   { return 100 }
