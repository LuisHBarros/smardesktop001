// License Management JavaScript

class LicenseManager {
    constructor() {
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.checkLicenseStatus();
    }

    setupEventListeners() {
        const form = document.getElementById('licenseForm');
        if (form) {
            form.addEventListener('submit', (e) => {
                e.preventDefault();
                this.setupLicense();
            });
        }
    }

    async checkLicenseStatus() {
        try {
            const response = await fetch('/api/license/status');
            const data = await response.json();
            
            this.updateStatusDisplay(data);
            this.updateMachineInfo(data.info);
        } catch (error) {
            console.error('Erro ao verificar status da licença:', error);
            this.showMessage('Erro ao verificar status da licença', 'error');
        }
    }

    updateStatusDisplay(data) {
        const statusElement = document.getElementById('licenseStatus');
        
        let statusHTML = '';
        
        if (data.has_license) {
            const statusClass = data.is_valid ? 'status-valid' : 'status-invalid';
            const statusIcon = data.is_valid ? '✅' : '❌';
            
            statusHTML = `
                <div class="license-status-item ${statusClass}">
                    <div class="status-icon">${statusIcon}</div>
                    <div class="status-details">
                        <div class="status-title">${data.message}</div>
                        <div class="status-subtitle">
                            ${data.is_valid ? 'Licença verificada e ativa' : 'Licença inativa ou expirada'}
                        </div>
                    </div>
                </div>
            `;
        } else {
            statusHTML = `
                <div class="license-status-item status-none">
                    <div class="status-icon">⚠️</div>
                    <div class="status-details">
                        <div class="status-title">Licença não configurada</div>
                        <div class="status-subtitle">Configure uma licença para usar o sistema</div>
                    </div>
                </div>
            `;
        }
        
        statusElement.innerHTML = statusHTML;
    }

    updateMachineInfo(info) {
        if (!info) {
            document.getElementById('deviceUUID').textContent = '--';
            document.getElementById('lastCheck').textContent = '--';
            document.getElementById('createdAt').textContent = '--';
            return;
        }

        document.getElementById('deviceUUID').textContent = info.device_uuid || '--';
        document.getElementById('lastCheck').textContent = this.formatDate(info.last_check) || '--';
        document.getElementById('createdAt').textContent = this.formatDate(info.created_at) || '--';
    }

    formatDate(dateString) {
        if (!dateString) return '--';
        
        try {
            const date = new Date(dateString);
            return date.toLocaleString('pt-BR');
        } catch (error) {
            return dateString;
        }
    }

    async setupLicense() {
        const token = document.getElementById('token').value.trim();
        const apiUrl = document.getElementById('apiUrl').value.trim();

        if (!token) {
            this.showMessage('Token é obrigatório', 'error');
            return;
        }

        try {
            this.showMessage('Configurando licença...', 'info');

            const response = await fetch('/api/license/setup', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    token: token,
                    api_url: apiUrl || 'http://localhost:8000'
                })
            });

            const data = await response.json();

            if (data.success) {
                this.showMessage('Licença configurada com sucesso!', 'success');
                document.getElementById('token').value = '';
                setTimeout(() => {
                    this.checkLicenseStatus();
                }, 1000);
            } else {
                this.showMessage(data.message || 'Erro ao configurar licença', 'error');
            }
        } catch (error) {
            console.error('Erro ao configurar licença:', error);
            this.showMessage('Erro ao configurar licença', 'error');
        }
    }

    async verifyLicense() {
        try {
            this.showMessage('Verificando licença...', 'info');

            const response = await fetch('/api/license/verify', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                }
            });

            const data = await response.json();

            if (data.valid) {
                this.showMessage('Licença verificada com sucesso!', 'success');
            } else {
                this.showMessage(data.message || 'Licença inválida', 'error');
            }

            setTimeout(() => {
                this.checkLicenseStatus();
            }, 1000);
        } catch (error) {
            console.error('Erro ao verificar licença:', error);
            this.showMessage('Erro ao verificar licença', 'error');
        }
    }

    async clearLicense() {
        if (!confirm('Tem certeza que deseja remover a licença? Esta ação não pode ser desfeita.')) {
            return;
        }

        try {
            this.showMessage('Removendo licença...', 'info');

            const response = await fetch('/api/license/clear', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                }
            });

            const data = await response.json();

            if (data.success) {
                this.showMessage('Licença removida com sucesso!', 'success');
                setTimeout(() => {
                    this.checkLicenseStatus();
                }, 1000);
            } else {
                this.showMessage(data.message || 'Erro ao remover licença', 'error');
            }
        } catch (error) {
            console.error('Erro ao remover licença:', error);
            this.showMessage('Erro ao remover licença', 'error');
        }
    }

    showMessage(message, type = 'info') {
        const container = document.getElementById('messageContainer');
        
        const messageElement = document.createElement('div');
        messageElement.className = `message message-${type}`;
        
        const icon = type === 'success' ? '✅' : type === 'error' ? '❌' : 'ℹ️';
        messageElement.innerHTML = `
            <span class="message-icon">${icon}</span>
            <span class="message-text">${message}</span>
            <button class="message-close" onclick="this.parentElement.remove()">×</button>
        `;
        
        container.appendChild(messageElement);
        
        // Remove automaticamente após 5 segundos
        setTimeout(() => {
            if (messageElement.parentElement) {
                messageElement.remove();
            }
        }, 5000);
    }
}

// Global functions for button clicks
function checkLicenseStatus() {
    if (window.licenseManager) {
        window.licenseManager.checkLicenseStatus();
    }
}

function verifyLicense() {
    if (window.licenseManager) {
        window.licenseManager.verifyLicense();
    }
}

function clearLicense() {
    if (window.licenseManager) {
        window.licenseManager.clearLicense();
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.licenseManager = new LicenseManager();
});
