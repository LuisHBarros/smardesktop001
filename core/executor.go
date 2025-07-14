package core

import (
	"fmt"
	"os"
	"os/exec"
)

// ExecuteProcess executa um processo externo de forma assíncrona
func ExecuteProcess(executablePath string) error {
	// Verifica se o arquivo executável existe
	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		return fmt.Errorf("executável não encontrado: %s", executablePath)
	}
	
	// Cria o comando
	cmd := exec.Command(executablePath)
	
	// Inicia o processo de forma assíncrona (não bloqueia)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("erro ao iniciar processo: %v", err)
	}
	
	return nil
}
