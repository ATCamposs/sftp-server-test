// pkg/config/config.go
package config

import (
	"sftp-client/logger"
	"strconv"
	"time"
)

// Config é uma struct que armazena a configuração do cliente SFTP.
// Em Go, campos de struct que começam com letra maiúscula são públicos.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Timeout  time.Duration
}

// NewConfig cria e retorna uma nova instância de Config.
// Em Go, funções que começam com "New" são padrao pra construtores.
func NewConfig(host string, portStr string, user, password string) *Config {
	// Converte a string da porta para um inteiro.
	port, err := strconv.Atoi(portStr)
	if err != nil {
		logger.GetLogger().Err(err).Msg("Erro ao converter a porta")
		port = 22 // Usa a porta padrão SSH se houver um erro
	}

	// Retorna um ponteiro para uma nova instância de Config.
	return &Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Timeout:  30 * time.Second,
	}
}
