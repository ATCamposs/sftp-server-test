package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func main() {
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
	}

	// Conectar ao servidor SSH
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sftpHost, sftpPort), config)
	if err != nil {
		log.Fatalf("Falha ao conectar ao servidor SSH: %v", err)
	}
	defer conn.Close()

	// Criar cliente SFTP
	client, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatalf("Falha ao criar cliente SFTP: %v", err)
	}
	defer client.Close()

	// Obter e imprimir o diretório atual
	pwd, err := client.Getwd()
	if err != nil {
		log.Printf("Erro ao obter diretório atual: %v", err)
	} else {
		fmt.Printf("Diretório atual: %s\n", pwd)
	}

	// Listar arquivos no diretório remoto
	files, err := client.ReadDir("/upload")
	if err != nil {
		log.Fatalf("Falha ao listar arquivos: %v", err)
	}

	// Ordenar arquivos por data de modificação (mais recente primeiro)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	// Imprimir lista de arquivos
	fmt.Println("Arquivos disponíveis:")
	for _, file := range files {
		fmt.Printf("%s - %s\n", file.Name(), file.ModTime().Format(time.RFC3339))
	}

	if len(files) > 0 {
		// Pegar o arquivo mais recente
		latestFile := files[0]
		fmt.Printf("\nBaixando o arquivo mais recente: %s\n", latestFile.Name())

		// Abrir arquivo remoto
		fullPath := filepath.Join("/upload", latestFile.Name())
		fmt.Printf("Tentando abrir arquivo: %s\n", fullPath)
		remoteFile, err := client.Open(fullPath)
		if err != nil {
			log.Fatalf("Falha ao abrir arquivo remoto: %v", err)
		}
		defer remoteFile.Close()

		// Criar arquivo local
		localFile, err := os.Create(latestFile.Name())
		if err != nil {
			log.Fatalf("Falha ao criar arquivo local: %v", err)
		}
		defer localFile.Close()

		// Copiar conteúdo do arquivo remoto para o local
		_, err = io.Copy(localFile, remoteFile)
		if err != nil {
			log.Fatalf("Falha ao copiar arquivo: %v", err)
		}

		fmt.Printf("Arquivo %s baixado com sucesso.\n", latestFile.Name())
	} else {
		fmt.Println("Nenhum arquivo encontrado.")
	}
}
