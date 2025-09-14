package redis

type RedisConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
}
