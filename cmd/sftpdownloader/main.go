package main

import (
	"os"
	"sftp-client/pkg/config"
	"sftp-client/pkg/downloader"
	"time"

	"sftp-client/logger"

	"github.com/joho/godotenv"
)

func init() {
	logger.InitLogger() // Inicializa o logger
	if err := godotenv.Load(); err != nil {
		logger.GetLogger().Err(err).Msg("Sem arquivo .env encontrado")
	}
}

func main() {
	start := time.Now()

	config := config.NewConfig(os.Getenv("HOST"), os.Getenv("PORT"), os.Getenv("SFTP_USER"), os.Getenv("SFTP_PASSWORD"))
	downloader := downloader.NewSFTPDownloader(config)

	if err := downloader.Connect(); err != nil {
		logger.GetLogger().Fatal().Err(err).Msg("Falha ao conectar ao servidor SFTP")
	}
	defer downloader.Close()

	if err := downloader.DownloadFiles("/upload", 10); err != nil {
		logger.GetLogger().Error().Err(err).Msg("Erro durante o download de arquivos")
	}

	logger.GetLogger().Info().Dur("executionTime", time.Since(start)).Msg("Tempo total de execução")
}
