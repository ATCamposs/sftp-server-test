package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/pkg/sftp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

//Porque usar Zerolog
//https://betterstack-community.github.io/go-logging-benchmarks/
func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func main() {
	start := time.Now()

	// Configurações de conexão SFTP
	sftpHost := "localhost"
	sftpPort := 2222
	sftpUser := "foo"
	sftpPass := "pass"

	// Configurar cliente SSH
	config := &ssh.ClientConfig{
		User: sftpUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sftpPass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Conectar ao servidor SSH
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sftpHost, sftpPort), config)
	if err != nil {
		log.Fatal().
			Str("host", sftpHost).
			Int("port", sftpPort).
			Err(err).
			Msg("Falha ao conectar ao servidor SSH")
	}
	defer conn.Close()

	// Criar cliente SFTP
	client, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatal().Err(err).Msg("Falha ao criar cliente SFTP")
	}
	defer client.Close()

	// Listar arquivos no diretório remoto
	files, err := client.ReadDir("/upload")
	if err != nil {
		log.Fatal().Err(err).Msg("Falha ao listar arquivos")
	}

	// Ordenar arquivos por data de modificação (mais recente primeiro)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	if len(files) == 0 {
		log.Warn().Msg("Nenhum arquivo encontrado")
		return
	}

	// Criar um pool de workers
	workerCount := 5
	jobs := make(chan os.FileInfo, len(files))
	results := make(chan error, len(files))

	// Iniciar workers
	for w := 1; w <= workerCount; w++ {
		go worker(w, client, jobs, results)
	}

	// Enviar jobs para os workers
	for _, file := range files[:min(10, len(files))] { // Limitar a 10 arquivos
		jobs <- file
	}
	close(jobs)

	// Coletar resultados
	for a := 1; a <= min(10, len(files)); a++ {
		if err := <-results; err != nil {
			log.Error().Err(err).Msg("Erro ao baixar arquivo")
		}
	}

	log.Info().Dur("executionTime", time.Since(start)).Msg("Tempo total de execução")
}

func worker(id int, client *sftp.Client, jobs <-chan os.FileInfo, results chan<- error) {
	for file := range jobs {
		log.Info().
			Int("worker", id).
			Str("file", file.Name()).
			Msg("Iniciando download")

		err := downloadFile(client, file)
		results <- err
	}
}

func downloadFile(client *sftp.Client, file os.FileInfo) error {
	remotePath := filepath.Join("/upload", file.Name())
	localPath := file.Name()

	// Abrir arquivo remoto
	remoteFile, err := client.Open(remotePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo remoto: %w", err)
	}
	defer remoteFile.Close()

	// Criar arquivo local
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo local: %w", err)
	}
	defer localFile.Close()

	// Pré-alocar espaço no disco
	if err := localFile.Truncate(file.Size()); err != nil {
		log.Warn().Err(err).Msg("Falha ao pré-alocar espaço no disco")
	}

	// Copiar conteúdo com buffer
	buffer := make([]byte, 32*1024)
	bytesWritten, err := io.CopyBuffer(localFile, remoteFile, buffer)
	if err != nil {
		return fmt.Errorf("erro durante a cópia do arquivo: %w", err)
	}

	log.Info().
		Str("file", file.Name()).
		Int64("bytesWritten", bytesWritten).
		Msg("Arquivo baixado com sucesso")
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
