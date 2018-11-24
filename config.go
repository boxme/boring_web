package main

import "fmt"

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
	Port int
	Env  string
}

func DefaultConfig() Config {
	return Config{
		Port: 3000,
		Env:  "dev",
	}
}

func (c Config) IsProd() bool {
	return c.Env == "prod"
}
