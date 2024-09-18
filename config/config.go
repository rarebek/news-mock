package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config -.
	Config struct {
		App     `yaml:"app"`
		HTTP    `yaml:"http"`
		Log     `yaml:"logger"`
		PG      `yaml:"postgres"`
		SMS     `yaml:"eskiz"`
		Redis   `yaml:"redis"`
		Casbin  `yaml:"casbin"`
		YouTube `yaml:"youtube"`
	}

	// App -.
	App struct {
		Name    string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Version string `env-required:"true" yaml:"version" env:"APP_VERSION"`
	}

	// HTTP -.
	HTTP struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"true" yaml:"log_level"   env:"LOG_LEVEL"`
	}

	// PG -.
	PG struct {
		PoolMax int    `env-required:"true" yaml:"pool_max" env:"PG_POOL_MAX"`
		URL     string `env-required:"true"                 env:"PG_URL"`
	}

	SMS struct {
		Token       string `yaml:"ESKIZ_TOKEN" env:"ESKIZ_TOKEN"`
		APIEndpoint string `yaml:"eskiz_api_endpoint" env:"ESKIZ_API_ENDPOINT"`
	}

	Redis struct {
		Host     string `env-required:"true" yaml:"redis_host" env:"REDIS_HOST"`
		Port     string `env-required:"true" yaml:"redis_port" env:"REDIS_PORT"`
		Password string `yaml:"redis_password" env:"REDIS_PASSWORD"`
		DB       int    `yaml:"redis_db" env:"REDIS_DB"`
	}

	Casbin struct {
		ConfigFilePath     string `env-required:"true" yaml:"config_file_path"`
		CSVFilePath        string `env-required:"true" yaml:"csv_file_path"`
		SigningKey         string `env-required:"true" yaml:"signing_key"`
		AccessTokenTimeOut int    `env-required:"true" yaml:"access_token_timeout"`
	}

	YouTube struct {
		ApiKey    string `yaml:"api_key"`
		ChannelID string `yaml:"channel_id"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("./config/config.yml", cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
