package config

import "fmt"

type Config struct {
	PostgresUser     string `env:"POSTGRES_USER" envDefault:"bannerify"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" envDefault:"bannerify"`
	PostgresDB       string `env:"POSTGRES_DB" envDefault:"bannerify"`
	PostgresPort     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	ServicePort      int    `env:"SERVICE_PORT" envDefault:"8080"`
	ServiceHost      string `env:"SERVICE_HOST" envDefault:"0.0.0.0"`
	MigrationsPath   string `env:"MIGRATIONS_PATH" envDefault:"migrations"`
	LogFilePath      string `env:"LOG_FILE_PATH" envDefault:"logfile.log"`
	JWTKey           string `env:"JWT_KEY" envDefault:"notreallysecret"`
	CachePort        int    `env:"REDIS_PORT" envDefault:"6379"`
}

func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@postgres:%d/%s?sslmode=disable",
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresPort,
		c.PostgresDB,
	)
}
