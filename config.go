package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type PostgresConfig struct {
	Host     string `json:"host"` // JSON tags
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:pasword`
	Name     string `json:"name"`
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host: "localhost",
		Port: 5432,
		User: "desmond",
		Name: "lenslocked_dev",
	}
}

func (c PostgresConfig) Dialect() string {
	return "postgres"
}

func (c PostgresConfig) ConnectionInfo() string {
	// Provide 2 potential connection info strings
	// based on whether a password is present
	if c.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Name)
	}

	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", c.Host, c.Port, c.Password, c.User, c.Name)
}

type Config struct {
	Port     int            `json:"port"`
	Env      string         `json:"env"`
	Pepper   string         `json:"pepper"`
	HMACKey  string         `json:"hmac_key"`
	Database PostgresConfig `json:"database"`
}

func DefaultConfig() Config {
	return Config{
		Port:     3000,
		Env:      "dev",
		Pepper:   "secret-random-string",
		HMACKey:  "secret-hmac-key",
		Database: DefaultPostgresConfig(),
	}
}

func LoadConfig(configReq bool) Config {
	f, err := os.Open(".config")
	if err != nil {
		if configReq {
			panic(err)
		}

		fmt.Println("Using the default config...")
		return DefaultConfig()
	}

	var c Config

	// We need a json decoder, which will read from the file
	// we opened when decoding
	decoder := json.NewDecoder(f)

	// Decode the file and place the results in c, the
	// Config variable we created for the results. The decoder
	// knows how to decode the data because of the struct tags
	// (eg `json:"port"`) we added to our Config and PostgresConfig
	// fields, much like GORM uses struct tags to know which
	// database column each field maps to.
	err = decoder.Decode(&c)
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully loaded .config")
	return c
}

func (c Config) IsProd() bool {
	return c.Env == "prod"
}
