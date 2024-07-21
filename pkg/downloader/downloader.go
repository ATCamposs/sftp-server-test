// sftp-server-test/downloader.go

package downloader

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"sftp-client/pkg/config"

	"github.com/pkg/sftp"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

type SFTPDownloader struct {
	config *config.Config
	client *sftp.Client
	conn   *ssh.Client
}

func NewSFTPDownloader(config *config.Config) *SFTPDownloader {
	return &SFTPDownloader{
		config: config,
	}
}

func (d *SFTPDownloader) Connect() error {
	sshConfig := &ssh.ClientConfig{
		User: d.config.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(d.config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         d.config.Timeout,
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", d.config.Host, d.config.Port), sshConfig)
	if err != nil {
		return fmt.Errorf("falha ao conectar ao servidor SSH: %w", err)
	}
	d.conn = conn

	client, err := sftp.NewClient(conn)
	if err != nil {
		return fmt.Errorf("falha ao criar cliente SFTP: %w", err)
	}
	d.client = client

	return nil
}

func (d *SFTPDownloader) Close() {
	if d.client != nil {
		d.client.Close()
	}
	if d.conn != nil {
		d.conn.Close()
	}
}

func (d *SFTPDownloader) DownloadFiles(remoteDir string, maxFiles int) error {
	files, err := d.client.ReadDir(remoteDir)
	if err != nil {
		return fmt.Errorf("falha ao listar arquivos: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	if len(files) == 0 {
		log.Warn().Msg("Nenhum arquivo encontrado")
		return nil
	}

	workerCount := 5
	jobs := make(chan os.FileInfo, len(files))
	results := make(chan error, len(files))

	for w := 1; w <= workerCount; w++ {
		go d.worker(w, jobs, results)
	}

	for _, file := range files[:min(maxFiles, len(files))] {
		jobs <- file
	}
	close(jobs)

	for a := 1; a <= min(maxFiles, len(files)); a++ {
		if err := <-results; err != nil {
			log.Error().Err(err).Msg("Erro ao baixar arquivo")
		}
	}

	return nil
}

func (d *SFTPDownloader) worker(id int, jobs <-chan os.FileInfo, results chan<- error) {
	for file := range jobs {
		log.Info().
			Int("worker", id).
			Str("file", file.Name()).
			Msg("Iniciando download")

		err := d.downloadFile(file)
		results <- err
	}
}

func (d *SFTPDownloader) downloadFile(file os.FileInfo) error {
	remotePath := filepath.Join("/upload", file.Name())
	localPath := file.Name()

	remoteFile, err := d.client.Open(remotePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo remoto: %w", err)
	}
	defer remoteFile.Close()

	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo local: %w", err)
	}
	defer localFile.Close()

	if err := localFile.Truncate(file.Size()); err != nil {
		log.Warn().Err(err).Msg("Falha ao pré-alocar espaço no disco")
	}

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
