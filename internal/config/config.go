package config

type Config struct {
	Token         string `env:"TOKEN" `
	AddrRedis     string `env:"ADDR_REDIS" envDefault:"localhost:6379"`
	PasswordRedis string `env:"PASSWORD_REDIS" envDefault:""`
}
