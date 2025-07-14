package service

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	ServiceName        = "GoDesktopApp"
	ServiceDisplayName = "Go Desktop App Service"
	ServiceDescription = "Serviço da aplicação Go Desktop App para logs da API"
)

// InstallService instala o serviço no Windows
func InstallService() error {
	exepath, err := GetExecutablePath()
	if err != nil {
		return err
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(ServiceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("serviço %s já existe", ServiceName)
	}

	s, err = m.CreateService(ServiceName, exepath, mgr.Config{
		DisplayName:      ServiceDisplayName,
		Description:      ServiceDescription,
		StartType:        mgr.StartAutomatic,
		ServiceStartName: "",
	}, "service")
	if err != nil {
		return err
	}
	defer s.Close()

	err = eventlog.InstallAsEventCreate(ServiceName, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() falhou: %s", err)
	}

	return nil
}

// RemoveService remove o serviço do Windows
func RemoveService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(ServiceName)
	if err != nil {
		return fmt.Errorf("serviço %s não está instalado", ServiceName)
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		return err
	}

	err = eventlog.Remove(ServiceName)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() falhou: %s", err)
	}

	return nil
}

// StartService inicia o serviço
func StartService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(ServiceName)
	if err != nil {
		return fmt.Errorf("não foi possível acessar o serviço %s: %v", ServiceName, err)
	}
	defer s.Close()

	err = s.Start("service")
	if err != nil {
		return fmt.Errorf("não foi possível iniciar o serviço %s: %v", ServiceName, err)
	}

	return nil
}

// StopService para o serviço
func StopService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(ServiceName)
	if err != nil {
		return fmt.Errorf("não foi possível acessar o serviço %s: %v", ServiceName, err)
	}
	defer s.Close()

	status, err := s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("não foi possível enviar comando de parada para o serviço %s: %v", ServiceName, err)
	}

	timeout := time.Now().Add(10 * time.Second)
	for status.State != svc.Stopped {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout aguardando o serviço %s parar", ServiceName)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("não foi possível consultar o status do serviço %s: %v", ServiceName, err)
		}
	}

	return nil
}

// ServiceStatus retorna o status do serviço
func ServiceStatus() (string, error) {
	m, err := mgr.Connect()
	if err != nil {
		return "", err
	}
	defer m.Disconnect()

	s, err := m.OpenService(ServiceName)
	if err != nil {
		return "Não instalado", nil
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return "", err
	}

	switch status.State {
	case svc.Stopped:
		return "Parado", nil
	case svc.StartPending:
		return "Iniciando", nil
	case svc.StopPending:
		return "Parando", nil
	case svc.Running:
		return "Executando", nil
	case svc.ContinuePending:
		return "Continuando", nil
	case svc.PausePending:
		return "Pausando", nil
	case svc.Paused:
		return "Pausado", nil
	default:
		return fmt.Sprintf("Estado desconhecido: %d", status.State), nil
	}
}

// PrintServiceCommands imprime os comandos disponíveis para gerenciar o serviço
func PrintServiceCommands() {
	fmt.Println("Comandos disponíveis:")
	fmt.Println("  install   - Instala o serviço")
	fmt.Println("  remove    - Remove o serviço")
	fmt.Println("  start     - Inicia o serviço")
	fmt.Println("  stop      - Para o serviço")
	fmt.Println("  status    - Mostra o status do serviço")
	fmt.Println("  service   - Executa como serviço (uso interno)")
	fmt.Println("")
	fmt.Println("Exemplo de uso:")
	fmt.Printf("  %s install\n", os.Args[0])
	fmt.Printf("  %s start\n", os.Args[0])
	fmt.Printf("  %s status\n", os.Args[0])
}
