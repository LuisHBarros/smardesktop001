package api

import (
	"encoding/json"
	"net/http"

	"go-desktop-app/core"
)

// StatusResponse representa a resposta do endpoint de status
type StatusResponse struct {
	Status string `json:"status"`
}

// FileRequest representa a requisição para operações com arquivos
type FileRequest struct {
	NomeArquivo string `json:"nome_arquivo"`
}

// FileContentResponse representa a resposta com conteúdo do arquivo
type FileContentResponse struct {
	Conteudo string `json:"conteudo"`
}

// MessageResponse representa uma resposta com mensagem
type MessageResponse struct {
	Mensagem string `json:"mensagem"`
}

// ErrorResponse representa uma resposta de erro
type ErrorResponse struct {
	Erro string `json:"erro"`
}

// ExecuteRequest representa a requisição para executar processo
type ExecuteRequest struct {
	CaminhoExecutavel string `json:"caminho_executavel"`
}

// StatusHandler retorna o status da API
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	
	response := StatusResponse{Status: "API está funcionando"}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ReadFileHandler lê o conteúdo de um arquivo
func ReadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	
	var req FileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Erro: "JSON inválido"})
		return
	}
	
	content, err := core.ReadFileContent(req.NomeArquivo)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Erro: err.Error()})
		return
	}
	
	response := FileContentResponse{Conteudo: content}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MoveFileHandler move um arquivo para o diretório de arquivo
func MoveFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	
	var req FileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Erro: "JSON inválido"})
		return
	}
	
	destPath, err := core.MoveFile(req.NomeArquivo)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Erro: err.Error()})
		return
	}
	
	response := MessageResponse{Mensagem: "Arquivo movido com sucesso para " + destPath}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExecuteProcessHandler executa um processo externo
func ExecuteProcessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	
	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Erro: "JSON inválido"})
		return
	}
	
	if err := core.ExecuteProcess(req.CaminhoExecutavel); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Erro: err.Error()})
		return
	}
	
	response := MessageResponse{Mensagem: "Processo iniciado com sucesso"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
