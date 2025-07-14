package service

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"

	"go-desktop-app/api"
	"go-desktop-app/database"
	"go-desktop-app/ui"
)

var elog debug.Log
var webFiles embed.FS

// SetWebFiles define os arquivos web embarcados para o serviço
func SetWebFiles(files embed.FS) {
	webFiles = files
}

type myservice struct{}

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	
	// Inicia o serviço
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Inicia a aplicação
	go func() {
		startApplication()
	}()
	
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				elog.Info(1, "Parando serviço...")
				cancel()
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				elog.Error(1, fmt.Sprintf("Comando inesperado do serviço: %v", c.Cmd))
			}
		case <-ctx.Done():
			break loop
		}
	}
	
	changes <- svc.Status{State: svc.StopPending}
	return
}

func startApplication() {
	// Configura o log para arquivo quando rodando como serviço
	logFile, err := os.OpenFile("go-desktop-app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	log.Println("Iniciando Go Desktop App como serviço...")

	// Inicializa o banco de dados
	if err := database.InitDatabase(); err != nil {
		log.Printf("Erro ao inicializar banco de dados: %v", err)
		// Continua a execução mesmo com erro no banco
	}

	// Configura os arquivos web embarcados
	api.SetWebFiles(webFiles)

	// Inicia o servidor da API
	go func() {
		log.Println("Iniciando servidor API...")
		api.StartServer()
	}()

	// Configura o system tray (se disponível)
	log.Println("Configurando system tray...")
	ui.SetupTray()
}

func runService(name string, isDebug bool) {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("Iniciando serviço %s", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &myservice{})
	if err != nil {
		elog.Error(1, fmt.Sprintf("Erro ao executar serviço %s: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("Serviço %s parado", name))
}

// RunAsService executa a aplicação como serviço Windows
func RunAsService(serviceName string) {
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Falha ao determinar se está rodando como serviço: %v", err)
	}

	if isService {
		runService(serviceName, false)
		return
	}

	// Se não está rodando como serviço, roda em modo debug
	runService(serviceName, true)
}

// GetExecutablePath retorna o caminho completo do executável
func GetExecutablePath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(ex)
}
