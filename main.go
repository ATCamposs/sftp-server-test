package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)
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
		log.WithFields(logrus.Fields{
			"host": sftpHost,
			"port": sftpPort,
		}).Error("Falha ao conectar ao servidor SSH")
		log.Fatal(err)
	}
	log.WithFields(logrus.Fields{
		"host": sftpHost,
		"port": sftpPort,
	}).Info("Conexão SSH estabelecida com sucesso")
	defer conn.Close()

	// Criar cliente SFTP
	client, err := sftp.NewClient(conn)
	if err != nil {
		log.Error("Falha ao criar cliente SFTP após conexão SSH bem-sucedida")
		log.Fatal(err)
	}
	log.Info("Cliente SFTP criado com sucesso")
	defer client.Close()

	// Obter e imprimir o diretório atual
	pwd, err := client.Getwd()
	if err != nil {
		log.WithError(err).Warn("Erro ao obter diretório atual")
	} else {
		log.WithField("directory", pwd).Info("Diretório atual")
	}

	// Verificar permissões do diretório remoto
	info, err := client.Stat("/upload")
	if err != nil {
		log.WithError(err).Error("Erro ao verificar informações do diretório /upload")
		log.Fatal("Falha ao verificar diretório")
	}
	log.WithField("permissions", info.Mode()).Debug("Permissões do diretório /upload")

	// Listar arquivos no diretório remoto
	files, err := client.ReadDir("/upload")
	if err != nil {
		log.WithError(err).Error("Erro ao tentar ler o diretório /upload")
		log.Fatal("Falha ao listar arquivos")
	}
	log.WithField("fileCount", len(files)).Info("Arquivos lidos do diretório /upload")

	// Ordenar arquivos por data de modificação (mais recente primeiro)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	// Imprimir lista de arquivos
	log.Info("Arquivos disponíveis:")
	for _, file := range files {
		log.WithFields(logrus.Fields{
			"fileName": file.Name(),
			"modTime":  file.ModTime().Format(time.RFC3339),
		}).Debug("Arquivo encontrado")
	}

	if len(files) > 0 {
		// Pegar o arquivo mais recente
		latestFile := files[0]
		log.WithField("fileName", latestFile.Name()).Info("Baixando o arquivo mais recente")

		// Verificar espaço em disco
		var stat syscall.Statfs_t
		wd, _ := os.Getwd()
		syscall.Statfs(wd, &stat)
		freeSpace := stat.Bavail * uint64(stat.Bsize)
		if uint64(latestFile.Size()) > freeSpace {
			log.Error("Espaço insuficiente no disco para baixar o arquivo")
			return
		}
		log.WithField("freeSpace", freeSpace).Debug("Espaço livre no disco")

		// Abrir arquivo remoto
		fullPath := filepath.Join("/upload", latestFile.Name())
		log.WithField("filePath", fullPath).Debug("Tentando abrir arquivo remoto")
		remoteFile, err := client.Open(fullPath)
		if err != nil {
			log.WithError(err).Error("Erro ao tentar abrir o arquivo remoto")
			return
		}
		log.WithField("filePath", fullPath).Info("Arquivo remoto aberto com sucesso")
		defer remoteFile.Close()

		// Criar arquivo local
		localFile, err := os.Create(latestFile.Name())
		if err != nil {
			log.WithError(err).Error("Erro ao tentar criar o arquivo local")
			return
		}
		log.WithField("fileName", latestFile.Name()).Info("Arquivo local criado com sucesso")
		defer localFile.Close()

		// Copiar conteúdo do arquivo remoto para o local
		bytesWritten, err := io.Copy(localFile, remoteFile)
		if err != nil {
			log.WithFields(logrus.Fields{
				"remoteFile": fullPath,
				"localFile":  latestFile.Name(),
			}).Error("Erro durante a cópia do arquivo")
			return
		}
		log.WithFields(logrus.Fields{
			"fileName":     latestFile.Name(),
			"bytesWritten": bytesWritten,
		}).Info("Arquivo copiado com sucesso")

	} else {
		log.Warn("Nenhum arquivo encontrado")
	}

	log.WithField("executionTime", time.Since(start)).Info("Tempo total de execução")
}
