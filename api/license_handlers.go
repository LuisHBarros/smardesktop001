package api

import (
	"encoding/json"
	"net/http"

	"go-desktop-app/database"
	"go-desktop-app/license"
)

// LicenseStatusResponse representa a resposta do status da licença
type LicenseStatusResponse struct {
	HasLicense bool                   `json:"has_license"`
	IsValid    bool                   `json:"is_valid"`
	Info       *database.LicenseInfo  `json:"info,omitempty"`
	Message    string                 `json:"message"`
}

// SetupLicenseRequest representa a requisição para configurar licença
type SetupLicenseRequest struct {
	Token   string `json:"token"`
	APIUrl  string `json:"api_url"`
}

// SetupLicenseResponse representa a resposta da configuração de licença
type SetupLicenseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// VerifyLicenseResponse representa a resposta da verificação de licença
type VerifyLicenseResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// LicenseStatusHandler retorna o status atual da licença
func LicenseStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Verifica se há informações de licença
	hasLicense := database.HasLicenseInfo()
	
	response := LicenseStatusResponse{
		HasLicense: hasLicense,
		IsValid:    false,
		Message:    "Licença não configurada",
	}

	if hasLicense {
		// Recupera as informações de licença
		info, err := database.GetLicenseInfo()
		if err != nil {
			response.Message = "Erro ao recuperar informações de licença"
		} else {
			response.Info = info
			response.IsValid = info.IsActive
			if info.IsActive {
				response.Message = "Licença ativa"
			} else {
				response.Message = "Licença inativa"
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SetupLicenseHandler configura uma nova licença
func SetupLicenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req SetupLicenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SetupLicenseResponse{
			Success: false,
			Message: "JSON inválido",
		})
		return
	}

	// Valida os campos obrigatórios
	if req.Token == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SetupLicenseResponse{
			Success: false,
			Message: "Token é obrigatório",
		})
		return
	}

	// Define URL padrão se não fornecida
	apiURL := req.APIUrl
	if apiURL == "" {
		apiURL = "http://localhost:8000" // URL padrão da API de licenciamento
	}

	// Cria o cliente de licenciamento
	client := license.NewLicenseClient(apiURL)

	// Configura a licença
	err := client.SetupLicense(req.Token)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SetupLicenseResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Resposta de sucesso
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SetupLicenseResponse{
		Success: true,
		Message: "Licença configurada com sucesso",
	})
}

// VerifyLicenseHandler verifica a licença atual
func VerifyLicenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Verifica se há informações de licença
	if !database.HasLicenseInfo() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(VerifyLicenseResponse{
			Valid:   false,
			Message: "Licença não configurada",
		})
		return
	}

	// Cria o cliente de licenciamento (usando URL padrão)
	client := license.NewLicenseClient("http://localhost:8000")

	// Verifica a licença
	valid, err := client.CheckLicense()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(VerifyLicenseResponse{
			Valid:   false,
			Message: err.Error(),
		})
		return
	}

	// Resposta de sucesso
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(VerifyLicenseResponse{
		Valid:   valid,
		Message: "Licença verificada com sucesso",
	})
}

// ClearLicenseHandler remove as informações de licença
func ClearLicenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Remove as informações de licença
	err := database.ClearLicenseInfo()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SetupLicenseResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Resposta de sucesso
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SetupLicenseResponse{
		Success: true,
		Message: "Licença removida com sucesso",
	})
}
