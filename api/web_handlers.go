package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LogEntry representa uma entrada de log
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
	Type      string `json:"type"`
}

// LogsResponse representa a resposta do endpoint de logs
type LogsResponse struct {
	Logs []LogEntry `json:"logs"`
}

// Sistema de logs em memória
var (
	logEntries []LogEntry
	logsMutex  sync.RWMutex
	logClients []chan LogEntry
	clientsMutex sync.RWMutex
)

// AddLogEntry adiciona uma nova entrada de log
func AddLogEntry(content string) {
	logsMutex.Lock()
	defer logsMutex.Unlock()

	entry := LogEntry{
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000"),
		Content:   content,
		Type:      determineLogType(content),
	}

	logEntries = append(logEntries, entry)

	// Manter apenas os últimos 1000 logs
	if len(logEntries) > 1000 {
		logEntries = logEntries[len(logEntries)-1000:]
	}

	// Notifica todos os clientes conectados
	notifyClients(entry)
}

// notifyClients envia o novo log para todos os clientes SSE conectados
func notifyClients(entry LogEntry) {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	fmt.Printf("DEBUG: Notificando %d clientes SSE sobre novo log\n", len(logClients))

	for i, client := range logClients {
		select {
		case client <- entry:
			fmt.Printf("Log enviado para cliente %d\n", i)
		default:
			fmt.Printf("Cliente %d não está respondendo, ignorando\n", i)
		}
	}
}

// determineLogType determina o tipo do log baseado no conteúdo
func determineLogType(content string) string {
	contentLower := strings.ToLower(content)
	if strings.Contains(contentLower, "erro") || strings.Contains(contentLower, "error") || strings.Contains(contentLower, "failed") {
		return "error"
	} else if strings.Contains(contentLower, "sucesso") || strings.Contains(contentLower, "success") || strings.Contains(contentLower, "200") {
		return "success"
	}
	return "info"
}

// WebHandler serve a interface web principal usando arquivos embarcados
func WebHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Se a URL é apenas "/", serve o index.html
	urlPath := r.URL.Path
	if urlPath == "/" {
		urlPath = "/index.html"
	}

	// Remove a barra inicial para construir o caminho do arquivo
	filePath := strings.TrimPrefix(urlPath, "/")
	// Usa sempre barras normais para arquivos embarcados
	fullPath := "web/" + filePath

	// Lê o arquivo dos arquivos embarcados
	data, err := webFiles.ReadFile(fullPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Define o Content-Type baseado na extensão do arquivo
	ext := filepath.Ext(fullPath)
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Serve o arquivo embarcado
	w.Write(data)
}

// LogsAPIHandler retorna os logs em formato JSON
func LogsAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	logsMutex.RLock()
	defer logsMutex.RUnlock()

	// Cria uma cópia dos logs para evitar problemas de concorrência
	logsCopy := make([]LogEntry, len(logEntries))
	copy(logsCopy, logEntries)

	response := LogsResponse{
		Logs: logsCopy,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ClearLogsHandler limpa todos os logs
func ClearLogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	logsMutex.Lock()
	defer logsMutex.Unlock()

	logEntries = []LogEntry{}

	response := MessageResponse{
		Mensagem: "Logs limpos com sucesso",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// LogsStreamHandler implementa Server-Sent Events para logs em tempo real
func LogsStreamHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DEBUG: Nova conexão SSE de %s\n", r.RemoteAddr)

	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Configura headers para SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	fmt.Printf("DEBUG: Headers SSE configurados\n")

	// Cria um canal para este cliente
	clientChan := make(chan LogEntry, 100)

	// Adiciona o cliente à lista
	clientsMutex.Lock()
	logClients = append(logClients, clientChan)
	clientCount := len(logClients)
	clientsMutex.Unlock()

	fmt.Printf("Novo cliente SSE conectado. Total de clientes: %d\n", clientCount)

	// Remove o cliente quando a conexão fechar
	defer func() {
		clientsMutex.Lock()
		for i, client := range logClients {
			if client == clientChan {
				logClients = append(logClients[:i], logClients[i+1:]...)
				break
			}
		}
		clientsMutex.Unlock()
		close(clientChan)
	}()

	// Envia logs existentes primeiro
	logsMutex.RLock()
	existingCount := len(logEntries)
	for _, entry := range logEntries {
		data, _ := json.Marshal(entry)
		fmt.Fprintf(w, "data: %s\n\n", data)
	}
	logsMutex.RUnlock()

	fmt.Printf("DEBUG: Enviados %d logs existentes para novo cliente SSE\n", existingCount)

	// Flush para enviar os dados imediatamente
	flusher, ok := w.(http.Flusher)
	if ok {
		flusher.Flush()
	}

	// Escuta por novos logs ou desconexão
	for {
		select {
		case entry := <-clientChan:
			// Novo log recebido, envia para o cliente
			data, _ := json.Marshal(entry)
			fmt.Printf("DEBUG: Enviando log via SSE: %s\n", string(data))
			fmt.Fprintf(w, "data: %s\n\n", data)
			if ok {
				flusher.Flush()
				fmt.Printf("DEBUG: Flush realizado\n")
			}
		case <-r.Context().Done():
			// Cliente desconectou
			fmt.Printf("DEBUG: Cliente SSE desconectou\n")
			return
		}
	}
}
