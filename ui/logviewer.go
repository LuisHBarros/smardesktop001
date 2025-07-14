package ui

import (
	"log"
	"strings"
	"sync"

	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
)

// LogWindow representa a janela de visualização de logs
type LogWindow struct {
	window   *walk.MainWindow
	textEdit *walk.TextEdit
	logs     []string
	mutex    sync.RWMutex
	visible  bool
}

// NewLogWindow cria uma nova janela de logs
func NewLogWindow() *LogWindow {
	lw := &LogWindow{
		logs: make([]string, 0),
	}

	// Tenta criar a janela, mas não falha se houver erro
	if err := lw.createWindow(); err != nil {
		log.Printf("Aviso: Janela de logs nativa não disponível: %v", err)
		log.Printf("Os logs continuarão sendo exibidos no console e na interface web")
		// Retorna a estrutura mesmo com erro para manter compatibilidade
		return lw
	}

	return lw
}

// createWindow cria a interface da janela
func (lw *LogWindow) createWindow() error {
	err := declarative.MainWindow{
		AssignTo: &lw.window,
		Title:    "Go Desktop App - Logs",
		Size:     declarative.Size{Width: 800, Height: 600},
		Layout:   declarative.VBox{},
		Visible:  false, // Inicia oculta
		Children: []declarative.Widget{
			declarative.TextEdit{
				AssignTo: &lw.textEdit,
				ReadOnly: true,
				VScroll:  true,
			},
		},
	}.Create()

	if err != nil {
		return err
	}

	// Configura o evento de fechamento
	if lw.window != nil {
		lw.window.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
			// Apenas oculta a janela em vez de fechar
			*canceled = true
			lw.Hide()
		})
	}

	return nil
}

// Show exibe a janela de logs
func (lw *LogWindow) Show() {
	if lw.window != nil {
		lw.visible = true
		lw.window.SetVisible(true)
		lw.window.BringToTop()
		lw.updateDisplay()
	} else {
		// Se a janela nativa não está disponível, abre a interface web
		log.Println("Janela nativa não disponível. Abrindo interface web em http://localhost:8080")
		// Aqui você poderia adicionar código para abrir o navegador automaticamente
		// exec.Command("cmd", "/c", "start", "http://localhost:8080").Start()
	}
}

// Hide oculta a janela de logs
func (lw *LogWindow) Hide() {
	if lw.window != nil {
		lw.visible = false
		lw.window.SetVisible(false)
	}
}

// Close fecha a janela de logs
func (lw *LogWindow) Close() {
	if lw.window != nil {
		lw.window.Close()
	}
}

// AddLog adiciona uma nova entrada de log
func (lw *LogWindow) AddLog(message string) {
	lw.mutex.Lock()
	defer lw.mutex.Unlock()
	
	lw.logs = append(lw.logs, message)
	
	// Mantém apenas os últimos 1000 logs para evitar uso excessivo de memória
	if len(lw.logs) > 1000 {
		lw.logs = lw.logs[len(lw.logs)-1000:]
	}
	
	// Atualiza a exibição se a janela estiver visível
	if lw.visible && lw.textEdit != nil {
		lw.updateDisplay()
	}
}

// updateDisplay atualiza o conteúdo exibido na janela
func (lw *LogWindow) updateDisplay() {
	lw.mutex.RLock()
	defer lw.mutex.RUnlock()
	
	if lw.textEdit != nil {
		content := strings.Join(lw.logs, "\n")
		lw.textEdit.SetText(content)
		
		// Rola para o final
		lw.textEdit.SendMessage(0x0115, 7, 0) // WM_VSCROLL, SB_BOTTOM
	}
}

// GetInstance retorna a instância global da janela de logs
var globalLogWindow *LogWindow

// InitGlobalLogWindow inicializa a janela de logs global
func InitGlobalLogWindow() {
	globalLogWindow = NewLogWindow()
	// Se falhar, globalLogWindow será nil, mas a aplicação continuará funcionando
}

// AddGlobalLog adiciona um log à janela global
func AddGlobalLog(message string) {
	if globalLogWindow != nil {
		globalLogWindow.AddLog(message)
	}
}
