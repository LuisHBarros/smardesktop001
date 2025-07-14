package ui

import (
	"fmt"
	"log"

	"go-desktop-app/database"
)

// GetLicenseStatusForTray obtém o status da licença para exibição no tray
func GetLicenseStatusForTray() string {
	// Verifica se há licença configurada
	if !database.HasLicenseInfo() {
		return "❌ Licença não configurada"
	}

	// Recupera as informações da licença
	info, err := database.GetLicenseInfo()
	if err != nil {
		log.Printf("Erro ao recuperar informações de licença: %v", err)
		return "❌ Erro ao verificar licença"
	}

	if info == nil {
		return "❌ Licença não encontrada"
	}

	// Verifica o status
	if info.IsActive {
		return fmt.Sprintf("✅ Licença ativa\nUUID: %s\nÚltima verificação: %s", 
			info.DeviceUUID[:8]+"...", 
			formatLastCheck(info.LastCheck))
	} else {
		return fmt.Sprintf("⚠️ Licença inativa\nUUID: %s", info.DeviceUUID[:8]+"...")
	}
}

// formatLastCheck formata o timestamp da última verificação
func formatLastCheck(lastCheck string) string {
	if lastCheck == "" {
		return "Nunca"
	}
	
	// Simplifica o formato para exibição no tray
	if len(lastCheck) > 16 {
		return lastCheck[:16] // Mostra apenas data e hora
	}
	
	return lastCheck
}

// UpdateTrayTooltipWithLicense atualiza o tooltip do tray com informações de licença
func UpdateTrayTooltipWithLicense() {
	status := GetLicenseStatusForTray()
	
	// Cria um tooltip mais informativo
	tooltip := fmt.Sprintf("Go Desktop App\n%s", status)
	
	// Esta função seria chamada para atualizar o tooltip do systray
	// Por enquanto, apenas loga a informação
	log.Printf("Tooltip atualizado: %s", tooltip)
}

// GetLicenseMenuText retorna o texto para o menu do tray baseado no status da licença
func GetLicenseMenuText() string {
	if !database.HasLicenseInfo() {
		return "⚙️ Configurar Licença"
	}

	info, err := database.GetLicenseInfo()
	if err != nil || info == nil {
		return "⚙️ Configurar Licença"
	}

	if info.IsActive {
		return "✅ Licença Ativa"
	} else {
		return "⚠️ Licença Inativa"
	}
}

// ShouldShowLicenseWarning verifica se deve mostrar aviso de licença
func ShouldShowLicenseWarning() bool {
	if !database.HasLicenseInfo() {
		return true
	}

	info, err := database.GetLicenseInfo()
	if err != nil || info == nil {
		return true
	}

	return !info.IsActive
}
