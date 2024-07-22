// cmd/sftpdownloader/main.go
package main

import (
	"os"
	"sftp-client/pkg/config"
	"sftp-client/pkg/downloader"
	"time"

	"sftp-client/logger"

	"github.com/joho/godotenv"
)

// init é uma função especial em Go que é executada antes da função main.
// Ela é usada para inicialização de pacotes.
// ela acontece antes da inicializacao real da aplicação, NAO É um constructor
func init() {
	logger.InitLogger() // Inicializa o logger
	if err := godotenv.Load(); err != nil {
		logger.GetLogger().Err(err).Msg("Sem arquivo .env encontrado")
	}
}

// main igual java
func main() {
	start := time.Now()

	// Em Go, := é usado para declaração e atribuição de variáveis em uma única linha(igual kotlin/java quando nao passa o tipo pra inicializar a variavel)
	config := config.NewConfig(os.Getenv("HOST"), os.Getenv("PORT"), os.Getenv("SFTP_USER"), os.Getenv("SFTP_PASSWORD"))
	downloader := downloader.NewSFTPDownloader(config)

	// Em Go, o tratamento de erros é feito através de valores de retorno, não exceções.(saudades)
	if err := downloader.Connect(); err != nil {
		logger.GetLogger().Fatal().Err(err).Msg("Falha ao conectar ao servidor SFTP")
	}

	// defer agenda a chamada de downloader.Close() para ser executada quando a função main terminar.
	// Isso garante que os recursos (conexões SFTP e SSH) sejam sempre fechados, mesmo se ocorrer um erro(evita memory leak).
	// O defer é executado na ordem LIFO (Last In, First Out), então múltiplos defer são executados na ordem inversa de declaração.
	defer downloader.Close()

	if err := downloader.DownloadFiles("/upload", 10); err != nil {
		logger.GetLogger().Error().Err(err).Msg("Erro durante o download de arquivos")
	}

	logger.GetLogger().Info().Dur("executionTime", time.Since(start)).Msg("Tempo total de execução")
}
