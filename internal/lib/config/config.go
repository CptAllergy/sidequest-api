package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr     string
	Server   Server
	Database Database
	Auth     Auth
}

type Server struct {
	AllowedOrigins    []string
	WriteTimeout      time.Duration
	ReadTimeout       time.Duration
	IdleTimeout       time.Duration
	MiddlewareTimeout time.Duration
}

type Database struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
	SslMode  string
	Settings Settings
}

type Settings struct {
	MaxConnections string
	MaxLifetime    string
	MaxIdleTime    string
}

type Auth struct {
	ZitadelUrl          string
	ZitadelClientId     string
	ZitadelInsecure     bool
	ZitadelInsecurePort string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	server, err := loadServer()
	if err != nil {
		return Config{}, err
	}

	auth, err := loadAuth()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Addr:   ":" + os.Getenv("SERVER_PORT"),
		Server: server,
		Database: Database{
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Name:     os.Getenv("DB_NAME"),
			SslMode:  os.Getenv("DB_SSLMODE"),
			Settings: Settings{
				MaxConnections: os.Getenv("DB_POOL_MAX_CONNECTIONS"),
				MaxLifetime:    os.Getenv("DB_MAX_CONN_LIFETIME"),
				MaxIdleTime:    os.Getenv("DB_MAX_IDLE_TIME"),
			},
		},
		Auth: auth,
	}, nil
}

func loadDuration(name string) (time.Duration, error) {
	value := os.Getenv(name)
	d, err := time.ParseDuration(value)

	if err != nil {
		return 0, fmt.Errorf("%s=%q: %w", name, value, err)
	}
	return d, nil
}

func loadBoolean(name string) (bool, error) {
	value := os.Getenv(name)
	b, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s=%q: %w", name, value, err)
	}
	return b, nil
}

func loadServer() (Server, error) {
	readTimeout, err := loadDuration("SERVER_READ_TIMEOUT")
	if err != nil {
		return Server{}, err
	}
	writeTimeout, err := loadDuration("SERVER_WRITE_TIMEOUT")
	if err != nil {
		return Server{}, err

	}
	idleTimeout, err := loadDuration("SERVER_IDLE_TIMEOUT")
	if err != nil {
		return Server{}, err
	}
	middlewareTimeout, err := loadDuration("SERVER_MIDDLEWARE_TIMEOUT")
	if err != nil {
		return Server{}, err
	}

	return Server{
		AllowedOrigins:    strings.Split(os.Getenv("SERVER_ALLOWED_ORIGINS"), ","),
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MiddlewareTimeout: middlewareTimeout,
	}, nil
}

func loadAuth() (Auth, error) {
	zitadelInsecure, err := loadBoolean("ZITADEL_INSECURE")
	if err != nil {
		return Auth{}, err
	}

	return Auth{
		ZitadelUrl:          os.Getenv("ZITADEL_URL"),
		ZitadelClientId:     os.Getenv("ZITADEL_CLIENT_ID"),
		ZitadelInsecure:     zitadelInsecure,
		ZitadelInsecurePort: os.Getenv("ZITADEL_INSECURE_PORT"),
	}, nil
}

func PrintStartupBanner() {
	ascii := `
   _____ _     __                           __
  / ___/(_)___/ /__  ____ ___  _____  _____/ /_
  \__ \/ / __  / _ \/ __ '/ / / / _ \/ ___/ __/
 ___/ / / /_/ /  __/ /_/ / /_/ /  __(__  ) /_
/____/_/\__,_/\___/\__, /\__,_/\___/____/\__/
                     /_/

`

	fmt.Print(ascii)
}
