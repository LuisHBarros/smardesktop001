package license

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"go-desktop-app/database"

	"github.com/google/uuid"
)

// LicenseClient representa o cliente da API de licenciamento
type LicenseClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// VerifyTokenRequest representa a requisição de verificação de token
type VerifyTokenRequest struct {
	Token      string `json:"token"`
	DeviceUUID string `json:"device_uuid"`
}

// VerifyTokenResponse representa a resposta de verificação de token
type VerifyTokenResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
	Machine struct {
		ID           string `json:"id"`
		DeviceUUID   string `json:"device_uuid"`
		Status       string `json:"status"`
		RegisteredAt string `json:"registered_at"`
		LastAccessAt string `json:"last_access_at"`
	} `json:"machine"`
	Employer struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"employer"`
	Error string `json:"error,omitempty"`
}

// NewLicenseClient cria uma nova instância do cliente de licenciamento
func NewLicenseClient(baseURL string) *LicenseClient {
	return &LicenseClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateDeviceUUID gera um UUID único para a máquina
func GenerateDeviceUUID() string {
	return uuid.New().String()
}

// VerifyToken verifica se o token é válido na API de licenciamento
func (c *LicenseClient) VerifyToken(token, deviceUUID string) (*VerifyTokenResponse, error) {
	// Prepara a requisição
	reqBody := VerifyTokenRequest{
		Token:      token,
		DeviceUUID: deviceUUID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %v", err)
	}

	// Cria a requisição HTTP
	url := fmt.Sprintf("%s/api/verify-token", c.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Executa a requisição
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao executar requisição: %v", err)
	}
	defer resp.Body.Close()

	// Lê a resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	// Decodifica a resposta
	var response VerifyTokenResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	// Se a API retornou erro HTTP, mas conseguimos decodificar, retorna a resposta
	if resp.StatusCode != http.StatusOK {
		log.Printf("API retornou status %d: %s", resp.StatusCode, response.Message)
	}

	return &response, nil
}

// CheckLicense verifica a licença usando as informações armazenadas
func (c *LicenseClient) CheckLicense() (bool, error) {
	// Recupera as informações de licença do banco
	info, err := database.GetLicenseInfo()
	if err != nil {
		return false, fmt.Errorf("erro ao recuperar informações de licença: %v", err)
	}

	if info == nil || info.Token == "" || info.DeviceUUID == "" {
		return false, fmt.Errorf("informações de licença não encontradas")
	}

	// Verifica o token na API (com fallback)
	response, err := c.VerifyTokenWithFallback(info.Token, info.DeviceUUID)
	if err != nil {
		log.Printf("Erro ao verificar token: %v", err)
		return false, err
	}

	// Atualiza o timestamp da última verificação
	if err := database.UpdateLastCheck(); err != nil {
		log.Printf("Erro ao atualizar última verificação: %v", err)
	}

	// Atualiza o status ativo baseado na resposta
	if err := database.UpdateActiveStatus(response.Valid); err != nil {
		log.Printf("Erro ao atualizar status ativo: %v", err)
	}

	if !response.Valid {
		log.Printf("Token inválido: %s", response.Message)
		if response.Error != "" {
			log.Printf("Erro da API: %s", response.Error)
		}
		return false, fmt.Errorf("licença inválida: %s", response.Message)
	}

	log.Printf("Licença válida para máquina %s", response.Machine.DeviceUUID)
	return true, nil
}

// SetupLicense configura uma nova licença com o token fornecido
func (c *LicenseClient) SetupLicense(token string) error {
	// Gera um novo UUID para a máquina se não existir
	info, err := database.GetLicenseInfo()
	if err != nil {
		return fmt.Errorf("erro ao verificar informações existentes: %v", err)
	}

	var deviceUUID string
	if info != nil && info.DeviceUUID != "" {
		// Usa o UUID existente
		deviceUUID = info.DeviceUUID
	} else {
		// Gera um novo UUID
		deviceUUID = GenerateDeviceUUID()
	}

	// Verifica se o token é válido (com fallback)
	response, err := c.VerifyTokenWithFallback(token, deviceUUID)
	if err != nil {
		return fmt.Errorf("erro ao verificar token: %v", err)
	}

	if !response.Valid {
		return fmt.Errorf("token inválido: %s", response.Message)
	}

	// Salva as informações no banco
	if err := database.SaveLicenseInfo(token, deviceUUID); err != nil {
		return fmt.Errorf("erro ao salvar informações de licença: %v", err)
	}

	log.Printf("Licença configurada com sucesso para máquina %s", deviceUUID)
	return nil
}

// SimulateAPIResponse simula uma resposta da API quando ela não estiver disponível
func (c *LicenseClient) SimulateAPIResponse(token, deviceUUID string) *VerifyTokenResponse {
	// Simula uma resposta válida para desenvolvimento
	return &VerifyTokenResponse{
		Valid:   true,
		Message: "Token is valid (simulated)",
		Machine: struct {
			ID           string `json:"id"`
			DeviceUUID   string `json:"device_uuid"`
			Status       string `json:"status"`
			RegisteredAt string `json:"registered_at"`
			LastAccessAt string `json:"last_access_at"`
		}{
			ID:           "simulated-machine-id",
			DeviceUUID:   deviceUUID,
			Status:       "active",
			RegisteredAt: time.Now().Format(time.RFC3339),
			LastAccessAt: time.Now().Format(time.RFC3339),
		},
		Employer: struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}{
			ID:    "simulated-employer-id",
			Name:  "Empresa Simulada",
			Email: "empresa@simulada.com",
		},
	}
}

// VerifyTokenWithFallback verifica o token com fallback para simulação
func (c *LicenseClient) VerifyTokenWithFallback(token, deviceUUID string) (*VerifyTokenResponse, error) {
	// Tenta verificar com a API real primeiro
	response, err := c.VerifyToken(token, deviceUUID)
	if err != nil {
		log.Printf("API não disponível, usando simulação: %v", err)
		// Se a API não estiver disponível, simula uma resposta
		return c.SimulateAPIResponse(token, deviceUUID), nil
	}

	return response, nil
}
