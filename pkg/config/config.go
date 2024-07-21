package config

import (
	"sftp-client/logger"
	"strconv"
	"time"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Timeout  time.Duration
}

func NewConfig(host string, portStr string, user, password string) *Config {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		logger.GetLogger().Err(err).Msg("Erro ao converter a porta")
		port = 22
	}

	return &Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Timeout:  30 * time.Second,
	}
}
