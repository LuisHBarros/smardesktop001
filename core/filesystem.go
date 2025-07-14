package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	
	"go-desktop-app/config"
)

// ReadFileContent lê o conteúdo de um arquivo no diretório APP_DIR
func ReadFileContent(filename string) (string, error) {
	fullPath := filepath.Join(config.APP_DIR, filename)
	
	// Verifica se o arquivo existe
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("arquivo não encontrado")
	}
	
	// Lê o conteúdo do arquivo
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("erro ao ler arquivo: %v", err)
	}
	
	return string(content), nil
}

// MoveFile move um arquivo do APP_DIR para o ARCHIVE_DIR
func MoveFile(filename string) (string, error) {
	sourcePath := filepath.Join(config.APP_DIR, filename)
	destPath := filepath.Join(config.ARCHIVE_DIR, filename)
	
	// Verifica se o arquivo de origem existe
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return "", fmt.Errorf("arquivo não encontrado")
	}
	
	// Cria o diretório de destino se não existir
	if err := os.MkdirAll(config.ARCHIVE_DIR, 0755); err != nil {
		return "", fmt.Errorf("erro ao criar diretório de destino: %v", err)
	}
	
	// Move o arquivo
	if err := os.Rename(sourcePath, destPath); err != nil {
		return "", fmt.Errorf("erro ao mover arquivo: %v", err)
	}
	
	return destPath, nil
}
