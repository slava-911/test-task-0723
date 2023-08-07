package config

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App struct {
		IsDebug        bool   `yaml:"is-debug" env:"IS_DEBUG" env-default:"false"`
		IsDevelopment  bool   `yaml:"is-development" env:"IS_DEV" env-default:"false"`
		Id             string `yaml:"id" env:"ID" env-default:"0"`
		Name           string `yaml:"name" env:"NAME" env-default:"app"`
		LogLevel       string `yaml:"log-level" env:"LOG_LEVEL" env-default:"trace"`
		MigrationsPath string `yaml:"migrations-path" env:"MIGRATIONS_PATH" env-default:"file:///migrations"`
		AdminUser      struct {
			Email    string `yaml:"email" env:"ADMIN_EMAIL" env-default:"admin"`
			Password string `yaml:"password" env:"ADMIN_PWD" env-default:"admin"`
		} `yaml:"admin"`
	} `yaml:"app"`
	HTTP struct {
		IP           string        `yaml:"ip" env:"HTTP-IP"`
		Port         string        `yaml:"port" env:"HTTP-PORT"`
		ReadTimeout  time.Duration `yaml:"read_timeout" env:"HTTP-READ-TIMEOUT"`
		WriteTimeout time.Duration `yaml:"write_timeout" env:"HTTP-WHITE-TIMEOUT"`
	} `yaml:"http"`
	JWT struct {
		Secret string `yaml:"secret" env:"JWT_SECRET"` //env-required:"true"`
	} `yaml:"jwt"`
	DB struct {
		Type              string        `yaml:"type" env:"DB_TYPE" env-default:"postgresql"`
		DSN               string        `yaml:"dsn" env:"POSTGRES_DSN"`
		Username          string        `yaml:"username" env:"POSTGRES_USER"`
		Password          string        `yaml:"password" env:"POSTGRES_PASSWORD"`
		Host              string        `yaml:"host" env:"POSTGRES_SERVER"`
		Port              string        `yaml:"port" env:"POSTGRES_PORT"`
		Name              string        `yaml:"name" env:"POSTGRES_DB"`
		MaxAttempts       int           `yaml:"max_attempts" env:"DB_MAXATTEMPTS"`
		ConnectionTimeout time.Duration `yaml:"connection_timeout" env:"DB_CONNTIMEOUT"`
	} `yaml:"database"`
}

const (
	EnvConfigPathName  = "CONFIG-PATH"
	FlagConfigPathName = "config"
)

var configPath string
var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		flag.StringVar(&configPath, FlagConfigPathName, "configs/config.local.yaml", "this is app config file")
		flag.Parse()

		log.Print("config initialization")
		if configPath == "" {
			configPath = os.Getenv(EnvConfigPathName)
		}
		if configPath == "" {
			log.Fatal("config path is required")
		}

		//if err := godotenv.Load("../.env"); err != nil {
		//	log.Fatalf("error loading env variables: %v", err)
		//}

		instance = &Config{}
		if err := cleanenv.ReadConfig(configPath, instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			log.Println(help)
			log.Fatal(err)
		}

	})
	return instance
}
