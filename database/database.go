package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // Pure Go SQLite driver (no CGO required)
	// _ "github.com/mattn/go-sqlite3" // CGO-based driver (commented out)
)

// LicenseInfo representa as informações de licença armazenadas
type LicenseInfo struct {
	ID         int    `json:"id"`
	Token      string `json:"token"`
	DeviceUUID string `json:"device_uuid"`
	IsActive   bool   `json:"is_active"`
	CreatedAt  string `json:"created_at"`
	LastCheck  string `json:"last_check"`
}

var db *sql.DB

// InitDatabase inicializa a conexão com o banco de dados
func InitDatabase() error {
	// Cria o diretório de dados se não existir
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório de dados: %v", err)
	}

	// Caminho do banco de dados
	dbPath := filepath.Join(dataDir, "license.db")

	// Abre a conexão com o banco
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("erro ao abrir banco de dados: %v", err)
	}

	// Testa a conexão
	if err := db.Ping(); err != nil {
		return fmt.Errorf("erro ao conectar com banco de dados: %v", err)
	}

	// Cria as tabelas se não existirem
	if err := createTables(); err != nil {
		return fmt.Errorf("erro ao criar tabelas: %v", err)
	}

	log.Println("Banco de dados inicializado com sucesso")
	return nil
}

// createTables cria as tabelas necessárias
func createTables() error {
	// Tabela principal para informações de licença
	licenseTableQuery := `
	CREATE TABLE IF NOT EXISTS license_info (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		token TEXT NOT NULL UNIQUE,
		device_uuid TEXT NOT NULL UNIQUE,
		is_active BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_check DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := db.Exec(licenseTableQuery)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela license_info: %v", err)
	}

	// Índices para melhor performance
	indexQueries := []string{
		`CREATE INDEX IF NOT EXISTS idx_license_token ON license_info(token);`,
		`CREATE INDEX IF NOT EXISTS idx_license_uuid ON license_info(device_uuid);`,
		`CREATE INDEX IF NOT EXISTS idx_license_active ON license_info(is_active);`,
	}

	for _, indexQuery := range indexQueries {
		_, err := db.Exec(indexQuery)
		if err != nil {
			log.Printf("Aviso: erro ao criar índice: %v", err)
		}
	}

	log.Println("Tabelas e índices criados com sucesso")
	return nil
}

// CloseDatabase fecha a conexão com o banco de dados
func CloseDatabase() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// HasLicenseInfo verifica se há informações de licença armazenadas
func HasLicenseInfo() bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM license_info").Scan(&count)
	if err != nil {
		log.Printf("Erro ao verificar licença: %v", err)
		return false
	}
	return count > 0
}

// GetLicenseInfo recupera as informações de licença
func GetLicenseInfo() (*LicenseInfo, error) {
	var info LicenseInfo
	query := `
		SELECT id, token, device_uuid, is_active, created_at, last_check 
		FROM license_info 
		ORDER BY id DESC 
		LIMIT 1
	`

	err := db.QueryRow(query).Scan(
		&info.ID,
		&info.Token,
		&info.DeviceUUID,
		&info.IsActive,
		&info.CreatedAt,
		&info.LastCheck,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("erro ao recuperar informações de licença: %v", err)
	}

	return &info, nil
}

// SaveLicenseInfo salva as informações de licença
func SaveLicenseInfo(token, deviceUUID string) error {
	// Valida os parâmetros
	if token == "" {
		return fmt.Errorf("token não pode estar vazio")
	}
	if deviceUUID == "" {
		return fmt.Errorf("device UUID não pode estar vazio")
	}

	// Remove informações antigas
	if err := ClearLicenseInfo(); err != nil {
		log.Printf("Aviso: erro ao limpar licença antiga: %v", err)
	}

	// Insere nova informação com timestamp atual
	query := `
		INSERT INTO license_info (token, device_uuid, is_active, created_at, last_check, updated_at)
		VALUES (?, ?, TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := db.Exec(query, token, deviceUUID)
	if err != nil {
		return fmt.Errorf("erro ao salvar informações de licença: %v", err)
	}

	// Verifica se a inserção foi bem-sucedida
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Aviso: não foi possível verificar linhas afetadas: %v", err)
	} else if rowsAffected == 0 {
		return fmt.Errorf("nenhuma linha foi inserida")
	}

	log.Printf("Informações de licença salvas com sucesso para dispositivo %s (token: %s...)",
		deviceUUID, token[:min(8, len(token))])
	return nil
}

// min retorna o menor valor entre dois inteiros
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// UpdateLastCheck atualiza o timestamp da última verificação
func UpdateLastCheck() error {
	query := `
		UPDATE license_info
		SET last_check = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = (SELECT MAX(id) FROM license_info)
	`

	result, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("erro ao atualizar última verificação: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Aviso: não foi possível verificar linhas afetadas: %v", err)
	} else if rowsAffected == 0 {
		return fmt.Errorf("nenhuma licença encontrada para atualizar")
	}

	return nil
}

// UpdateActiveStatus atualiza o status ativo da licença
func UpdateActiveStatus(isActive bool) error {
	query := `
		UPDATE license_info
		SET is_active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = (SELECT MAX(id) FROM license_info)
	`

	result, err := db.Exec(query, isActive)
	if err != nil {
		return fmt.Errorf("erro ao atualizar status ativo: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Aviso: não foi possível verificar linhas afetadas: %v", err)
	} else if rowsAffected == 0 {
		return fmt.Errorf("nenhuma licença encontrada para atualizar")
	}

	return nil
}

// ClearLicenseInfo remove todas as informações de licença
func ClearLicenseInfo() error {
	query := "DELETE FROM license_info"
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("erro ao limpar informações de licença: %v", err)
	}

	log.Println("Informações de licença removidas")
	return nil
}

// GetDatabaseStats retorna estatísticas do banco de dados
func GetDatabaseStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Conta total de registros de licença
	var licenseCount int
	err := db.QueryRow("SELECT COUNT(*) FROM license_info").Scan(&licenseCount)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar licenças: %v", err)
	}
	stats["license_count"] = licenseCount

	// Verifica se há licença ativa
	var activeCount int
	err = db.QueryRow("SELECT COUNT(*) FROM license_info WHERE is_active = TRUE").Scan(&activeCount)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar licenças ativas: %v", err)
	}
	stats["active_licenses"] = activeCount

	// Última verificação
	if licenseCount > 0 {
		var lastCheck string
		err = db.QueryRow("SELECT last_check FROM license_info ORDER BY id DESC LIMIT 1").Scan(&lastCheck)
		if err == nil {
			stats["last_check"] = lastCheck
		}
	}

	return stats, nil
}

// GetLicenseByToken recupera informações de licença pelo token
func GetLicenseByToken(token string) (*LicenseInfo, error) {
	var info LicenseInfo
	query := `
		SELECT id, token, device_uuid, is_active, created_at, last_check
		FROM license_info
		WHERE token = ?
		LIMIT 1
	`

	err := db.QueryRow(query, token).Scan(
		&info.ID,
		&info.Token,
		&info.DeviceUUID,
		&info.IsActive,
		&info.CreatedAt,
		&info.LastCheck,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("erro ao recuperar licença por token: %v", err)
	}

	return &info, nil
}

// GetLicenseByUUID recupera informações de licença pelo UUID do dispositivo
func GetLicenseByUUID(deviceUUID string) (*LicenseInfo, error) {
	var info LicenseInfo
	query := `
		SELECT id, token, device_uuid, is_active, created_at, last_check
		FROM license_info
		WHERE device_uuid = ?
		LIMIT 1
	`

	err := db.QueryRow(query, deviceUUID).Scan(
		&info.ID,
		&info.Token,
		&info.DeviceUUID,
		&info.IsActive,
		&info.CreatedAt,
		&info.LastCheck,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("erro ao recuperar licença por UUID: %v", err)
	}

	return &info, nil
}

// ValidateToken verifica se um token existe e está ativo
func ValidateToken(token string) (bool, error) {
	var isActive bool
	query := `
		SELECT is_active
		FROM license_info
		WHERE token = ?
		LIMIT 1
	`

	err := db.QueryRow(query, token).Scan(&isActive)
	if err == sql.ErrNoRows {
		return false, nil // Token não encontrado
	}

	if err != nil {
		return false, fmt.Errorf("erro ao validar token: %v", err)
	}

	return isActive, nil
}

// ValidateUUID verifica se um UUID existe e está ativo
func ValidateUUID(deviceUUID string) (bool, error) {
	var isActive bool
	query := `
		SELECT is_active
		FROM license_info
		WHERE device_uuid = ?
		LIMIT 1
	`

	err := db.QueryRow(query, deviceUUID).Scan(&isActive)
	if err == sql.ErrNoRows {
		return false, nil // UUID não encontrado
	}

	if err != nil {
		return false, fmt.Errorf("erro ao validar UUID: %v", err)
	}

	return isActive, nil
}

// GetAllLicenses retorna todas as licenças (para administração)
func GetAllLicenses() ([]LicenseInfo, error) {
	query := `
		SELECT id, token, device_uuid, is_active, created_at, last_check
		FROM license_info
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erro ao recuperar todas as licenças: %v", err)
	}
	defer rows.Close()

	var licenses []LicenseInfo
	for rows.Next() {
		var info LicenseInfo
		err := rows.Scan(
			&info.ID,
			&info.Token,
			&info.DeviceUUID,
			&info.IsActive,
			&info.CreatedAt,
			&info.LastCheck,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao escanear linha de licença: %v", err)
		}
		licenses = append(licenses, info)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar sobre licenças: %v", err)
	}

	return licenses, nil
}

// TestDatabaseConnection testa a conexão e funcionalidade básica do banco
func TestDatabaseConnection() error {
	if db == nil {
		return fmt.Errorf("banco de dados não inicializado")
	}

	// Testa a conexão
	if err := db.Ping(); err != nil {
		return fmt.Errorf("erro ao conectar com banco: %v", err)
	}

	// Testa uma query simples
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM license_info").Scan(&count)
	if err != nil {
		return fmt.Errorf("erro ao executar query de teste: %v", err)
	}

	log.Printf("Banco de dados funcionando corretamente. Total de licenças: %d", count)
	return nil
}

// CreateTestLicense cria uma licença de teste (apenas para desenvolvimento)
func CreateTestLicense() error {
	testToken := "test-token-12345678901234567890123456789012345678901234567890123456"
	testUUID := "550e8400-e29b-41d4-a716-446655440000"

	// Verifica se já existe
	existing, err := GetLicenseByToken(testToken)
	if err != nil {
		return fmt.Errorf("erro ao verificar licença existente: %v", err)
	}

	if existing != nil {
		log.Println("Licença de teste já existe")
		return nil
	}

	// Cria a licença de teste
	err = SaveLicenseInfo(testToken, testUUID)
	if err != nil {
		return fmt.Errorf("erro ao criar licença de teste: %v", err)
	}

	log.Printf("Licença de teste criada com sucesso (Token: %s..., UUID: %s)",
		testToken[:16], testUUID)
	return nil
}
