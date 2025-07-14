package ui

import (
	_ "embed"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/getlantern/systray"
)

//go:embed icon.ico
var embeddedIcon []byte

var (
	logWindow *LogWindow
)

// SetupTray configura o ícone da bandeja do sistema
func SetupTray() {
	log.Println("Iniciando system tray...")
	systray.Run(onReady, onExit)
}

// onReady é chamado quando o systray está pronto
func onReady() {
	log.Println("System tray inicializado com sucesso!")

	// Configura o ícone do sistema
	log.Println("Carregando ícone...")
	setTrayIcon()
	systray.SetTitle("Go App")

	// Atualiza tooltip com informações de licença
	updateTooltipWithLicense()

	log.Println("Ícone, título e tooltip definidos")

	// Cria os itens do menu
	mShowLogs := systray.AddMenuItem("📊 Abrir Logs", "Mostra a interface de logs")
	mLicense := systray.AddMenuItem(GetLicenseMenuText(), "Gerenciar licença do sistema")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Sair", "Encerra a aplicação")
	
	// Tenta inicializar a janela de logs nativa (pode falhar, mas não é crítico)
	InitGlobalLogWindow()
	logWindow = globalLogWindow

	// Inicia um goroutine para tentar reconfigurar o ícone periodicamente
	go refreshTrayIcon()
	
	// Loop para tratar eventos do menu
	go func() {
		for {
			select {
			case <-mShowLogs.ClickedCh:
				// Abre a interface web no Chrome como aplicativo
				openChromeApp()
				log.Println("Aplicativo Chrome aberto com sucesso")
			case <-mLicense.ClickedCh:
				// Abre a interface de licença no Chrome como aplicativo
				openLicenseApp()
				log.Println("Interface de licença aberta com sucesso")
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

// onExit é chamado quando o systray está sendo encerrado
func onExit() {
	log.Println("Encerrando aplicação...")
	if logWindow != nil {
		logWindow.Close()
	}
	os.Exit(0)
}

// setTrayIcon configura o ícone do system tray com fallbacks
func setTrayIcon() {
	var iconSet bool

	// Tenta usar o ícone embarcado primeiro
	if len(embeddedIcon) > 0 {
		log.Printf("Tentando usar ícone embarcado (%d bytes)", len(embeddedIcon))
		systray.SetIcon(embeddedIcon)
		iconSet = true
		log.Println("✅ Ícone embarcado configurado")
	}

	// Fallback 1: Tenta carregar icon.ico do diretório atual
	if !iconSet {
		if iconData, err := os.ReadFile("icon.ico"); err == nil {
			log.Printf("Usando icon.ico do diretório atual (%d bytes)", len(iconData))
			systray.SetIcon(iconData)
			iconSet = true
		} else {
			log.Printf("Não foi possível carregar icon.ico: %v", err)
		}
	}

	// Fallback 2: Tenta carregar do diretório ui/
	if !iconSet {
		if iconData, err := os.ReadFile("ui/icon.ico"); err == nil {
			log.Printf("Usando icon.ico do diretório ui/ (%d bytes)", len(iconData))
			systray.SetIcon(iconData)
			iconSet = true
		} else {
			log.Printf("Não foi possível carregar ui/icon.ico: %v", err)
		}
	}

	// Fallback 3: Cria um ícone simples programaticamente (PNG 16x16)
	if !iconSet {
		log.Println("Criando ícone padrão programaticamente")
		defaultIcon := createDefaultIcon()
		systray.SetIcon(defaultIcon)
		log.Printf("✅ Ícone padrão configurado (%d bytes)", len(defaultIcon))
	}

	// Força uma atualização do sistema tray
	systray.SetTitle("Go App")
	log.Println("🔄 Sistema tray atualizado")
}

// createDefaultIcon cria um ícone PNG simples de 16x16 pixels
func createDefaultIcon() []byte {
	// PNG mínimo de 16x16 pixels (azul simples)
	// Este é um PNG válido codificado em bytes
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x91, 0x68, 0x36, 0x00, 0x00, 0x00,
		0x3A, 0x49, 0x44, 0x41, 0x54, 0x28, 0x15, 0x63, 0x64, 0x54, 0x52, 0x64,
		0x00, 0x02, 0x46, 0x46, 0x06, 0x06, 0x86, 0x16, 0x26, 0x36, 0x0E, 0x1E,
		0x3E, 0x01, 0x31, 0x09, 0x19, 0x39, 0x05, 0x15, 0x35, 0x0D, 0x2D, 0x3D,
		0x03, 0x23, 0x13, 0x33, 0x0B, 0x2B, 0x3B, 0x07, 0x27, 0x17, 0x37, 0x0F,
		0x2F, 0x3F, 0x80, 0x90, 0x88, 0x98, 0x84, 0x94, 0x8C, 0x9C, 0x82, 0x92,
		0x8A, 0x9A, 0x86, 0x96, 0x8E, 0x9E, 0x81, 0x91, 0x89, 0x99, 0x00, 0x00,
		0x0C, 0x3F, 0x01, 0x37, 0x5D, 0x97, 0x1C, 0x50, 0x00, 0x00, 0x00, 0x00,
		0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
}

// refreshTrayIcon tenta reconfigurar o ícone periodicamente
func refreshTrayIcon() {
	// Aguarda 3 segundos antes da primeira tentativa
	time.Sleep(3 * time.Second)

	// Tenta reconfigurar o ícone algumas vezes
	for i := 0; i < 3; i++ {
		log.Printf("🔄 Tentativa %d de reconfiguração do ícone", i+1)
		setTrayIcon()
		updateTooltipWithLicense()
		time.Sleep(2 * time.Second)
	}

	log.Println("✅ Processo de reconfiguração do ícone concluído")
}

// updateTooltipWithLicense atualiza o tooltip do tray com informações de licença
func updateTooltipWithLicense() {
	status := GetLicenseStatusForTray()
	tooltip := "Go Desktop App\n" + status
	systray.SetTooltip(tooltip)
}

// openChromeApp abre a interface web no Chrome como aplicativo
func openChromeApp() {
	chromePath := "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
	url := "http://localhost:8080"

	// Parâmetros para simular um aplicativo desktop
	args := []string{
		"--app=" + url,
		"--window-size=1000,700",
		"--window-position=100,100",
		"--disable-web-security",
		"--disable-features=TranslateUI",
		"--no-first-run",
		"--no-default-browser-check",
	}

	cmd := exec.Command(chromePath, args...)

	if err := cmd.Start(); err != nil {
		log.Printf("Erro ao abrir Chrome: %v", err)
		log.Printf("Tentando abrir URL padrão: %s", url)
		// Fallback para comando padrão do Windows
		fallbackCmd := exec.Command("cmd", "/c", "start", url)
		fallbackCmd.Start()
	} else {
		log.Printf("Chrome app aberto em: %s", url)
	}
}

// openLicenseApp abre a interface de licença no Chrome como aplicativo
func openLicenseApp() {
	chromePath := "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
	url := "http://localhost:8080/license.html"

	// Parâmetros para simular um aplicativo desktop
	args := []string{
		"--app=" + url,
		"--window-size=800,600",
		"--window-position=200,200",
		"--disable-web-security",
		"--disable-features=TranslateUI",
		"--no-first-run",
		"--no-default-browser-check",
	}

	cmd := exec.Command(chromePath, args...)

	if err := cmd.Start(); err != nil {
		log.Printf("Erro ao abrir Chrome para licença: %v", err)
		log.Printf("Tentando abrir URL padrão: %s", url)
		// Fallback para comando padrão do Windows
		fallbackCmd := exec.Command("cmd", "/c", "start", url)
		fallbackCmd.Start()
	} else {
		log.Printf("Chrome app de licença aberto em: %s", url)
	}
}


