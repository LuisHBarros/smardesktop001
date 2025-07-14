package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	"go-desktop-app/api"
	"go-desktop-app/database"
	"go-desktop-app/service"
	"go-desktop-app/ui"
)

//go:embed web/*
var webFiles embed.FS

// CustomLogWriter é um writer customizado que envia logs para a janela
type CustomLogWriter struct{}

func (w CustomLogWriter) Write(p []byte) (n int, err error) {
	message := string(p)

	// Envia para o log padrão
	os.Stderr.Write(p)

	// Envia para a janela de logs nativa
	ui.AddGlobalLog(message)

	// NÃO envia para a interface web aqui - apenas o middleware deve fazer isso
	// para evitar duplicação de logs

	return len(p), nil
}

func main() {
	// Verifica se foi passado algum argumento para gerenciar o serviço
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			err := service.InstallService()
			if err != nil {
				fmt.Printf("Erro ao instalar serviço: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Serviço instalado com sucesso!")
			return
		case "remove":
			err := service.RemoveService()
			if err != nil {
				fmt.Printf("Erro ao remover serviço: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Serviço removido com sucesso!")
			return
		case "start":
			err := service.StartService()
			if err != nil {
				fmt.Printf("Erro ao iniciar serviço: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Serviço iniciado com sucesso!")
			return
		case "stop":
			err := service.StopService()
			if err != nil {
				fmt.Printf("Erro ao parar serviço: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Serviço parado com sucesso!")
			return
		case "status":
			status, err := service.ServiceStatus()
			if err != nil {
				fmt.Printf("Erro ao consultar status do serviço: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Status do serviço: %s\n", status)
			return
		case "service":
			// Configura os arquivos web embarcados para o serviço
			service.SetWebFiles(webFiles)
			// Executa como serviço
			service.RunAsService(service.ServiceName)
			return
		default:
			service.PrintServiceCommands()
			return
		}
	}

	// Se não há argumentos, executa normalmente (modo interativo)
	runInteractiveMode()
}

func runInteractiveMode() {
	// Oculta o console no Windows
	hideConsole()

	// Configura o logger customizado
	log.SetOutput(CustomLogWriter{})
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Iniciando Go Desktop App...")

	// Inicializa o banco de dados
	if err := database.InitDatabase(); err != nil {
		log.Printf("Erro ao inicializar banco de dados: %v", err)
		// Continua a execução mesmo com erro no banco
	}

	// Configura os arquivos web embarcados
	api.SetWebFiles(webFiles)

	// Inicia o servidor da API
	log.Println("Iniciando servidor API...")
	go func() {
		api.StartServer()
	}()

	log.Println("Configurando system tray...")

	// Abre automaticamente o navegador após um tempo
	go func() {
		// Aguarda um pouco para o servidor iniciar
		time.Sleep(3 * time.Second)
		openBrowser("http://localhost:8080")
	}()

	// Configura e inicia o system tray (bloqueia a thread principal)
	ui.SetupTray()
}

// hideConsole oculta a janela do console no Windows
func hideConsole() {
	console := getConsoleWindow()
	if console != 0 {
		showWindow(console, 0) // SW_HIDE = 0
	}
}

// getConsoleWindow obtém o handle da janela do console
func getConsoleWindow() uintptr {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetConsoleWindow")
	ret, _, _ := proc.Call()
	return ret
}

// showWindow controla a visibilidade da janela
func showWindow(hwnd uintptr, cmdshow int) bool {
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("ShowWindow")
	ret, _, _ := proc.Call(hwnd, uintptr(cmdshow))
	return ret != 0
}

// openBrowser abre uma URL no navegador padrão
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("plataforma não suportada")
	}
	if err != nil {
		log.Printf("Erro ao abrir navegador: %v", err)
	}
}
