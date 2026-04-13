package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"

	"github.com/Grisha1Kadetov/slotbooking/internal/pkg/log"
)

var cfg *Config

type Config struct {
	Port       string        `env:"PORT" env-default:"8080"`
	JWTSecret  string        `env:"JWT_SECRET" env-default:"secret"`
	PregenDay  int           `env:"PREGEN_DAY" env-default:"7"`
	SlotDur    time.Duration `env:"SLOT_DUR" env-default:"30m"`
	PostgresDB PostgresDB
}

type PostgresDB struct {
	User     string `env:"POSTGRES_USER"     env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DB       string `env:"POSTGRES_DB"       env-required:"true"`
	Port     string `env:"POSTGRES_PORT"     env-required:"true"`
	Host     string `env:"POSTGRES_HOST"     env-required:"true"`
}

func (p PostgresDB) GetDSN() string {
	return "postgres://" + p.User + ":" + p.Password + "@" + p.Host + ":" + p.Port + "/" + p.DB
}

func GetConfig(l log.Logger) Config {
	if cfg != nil {
		return *cfg
	}
	err := godotenv.Load()
	if err != nil {
		l.Warn("can't load .env file")
	}
	cfg = &Config{}
	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		panic(err)
	}
	return *cfg
}
