package api

import (
	"embed"
	"log"
	"net/http"

	"go-desktop-app/config"
)

var webFiles embed.FS

// SetWebFiles define os arquivos web embarcados
func SetWebFiles(files embed.FS) {
	webFiles = files
}

// StartServer inicia o servidor HTTP da API
func StartServer() {
	// Cria o multiplexador de rotas
	mux := http.NewServeMux()

	// Registra as rotas da API
	mux.HandleFunc("/status", StatusHandler)
	mux.HandleFunc("/escreve_arquivo", ReadFileHandler)
	mux.HandleFunc("/move_arquivo", MoveFileHandler)
	mux.HandleFunc("/executar_terceiros", ExecuteProcessHandler)

	// Registra as rotas da interface web
	mux.HandleFunc("/api/logs", LogsAPIHandler)
	mux.HandleFunc("/api/logs/clear", ClearLogsHandler)
	mux.HandleFunc("/api/logs/stream", LogsStreamHandler)

	// Registra as rotas de licenciamento
	mux.HandleFunc("/api/license/status", LicenseStatusHandler)
	mux.HandleFunc("/api/license/setup", SetupLicenseHandler)
	mux.HandleFunc("/api/license/verify", VerifyLicenseHandler)
	mux.HandleFunc("/api/license/clear", ClearLicenseHandler)

	// Registra o handler para servir arquivos estáticos (deve ser o último)
	mux.HandleFunc("/", WebHandler)

	// Aplica os middlewares
	handler := LoggingMiddleware(CORSMiddleware(mux))

	log.Printf("Servidor API iniciado na porta %s", config.API_PORT)
	log.Printf("Interface web disponível em: http://localhost%s", config.API_PORT)

	// Inicia o servidor
	if err := http.ListenAndServe(config.API_PORT, handler); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
