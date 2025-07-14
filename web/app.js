class LogsApp {
    constructor() {
        console.log('=== CONSTRUTOR LogsApp INICIADO ===');

        this.logs = [];
        this.totalLogs = 0;
        this.errorCount = 0;
        this.successCount = 0;
        this.lastLogTime = null;

        console.log('=== OBTENDO ELEMENTOS DOM ===');
        this.logsContainer = document.getElementById('logsContainer');
        this.totalLogsElement = document.getElementById('totalLogs');
        this.lastLogTimeElement = document.getElementById('lastLogTime');
        this.errorCountElement = document.getElementById('errorCount');
        this.successCountElement = document.getElementById('successCount');

        console.log('Elementos encontrados:', {
            logsContainer: !!this.logsContainer,
            totalLogsElement: !!this.totalLogsElement,
            lastLogTimeElement: !!this.lastLogTimeElement,
            errorCountElement: !!this.errorCountElement,
            successCountElement: !!this.successCountElement
        });

        console.log('=== INICIANDO APLICAÇÃO ===');
        this.init();

        console.log('=== CONSTRUTOR LogsApp FINALIZADO ===');
    }
    
    init() {
        this.loadInitialLogs();
        this.startSmartPolling();
    }

    setupSSE() {
        // Configura Server-Sent Events para logs em tempo real
        if (typeof EventSource !== 'undefined') {
            console.log('DEBUG: Iniciando conexão SSE...');

            try {
                this.eventSource = new EventSource('/api/logs/stream');
                console.log('DEBUG: EventSource criado, estado:', this.eventSource.readyState);

                this.eventSource.onopen = (event) => {
                    console.log('DEBUG: SSE conectado com sucesso!', event);
                    console.log('DEBUG: Estado da conexão:', this.eventSource.readyState);
                };

                this.eventSource.onmessage = (event) => {
                    console.log('DEBUG: Mensagem SSE recebida:', event);
                    console.log('DEBUG: Dados da mensagem:', event.data);
                    try {
                        const logData = JSON.parse(event.data);
                        console.log('DEBUG: Log parseado:', logData);
                        this.addLogFromSSE(logData);
                    } catch (error) {
                        console.error('ERRO: Falha ao processar log SSE:', error);
                        console.error('ERRO: Dados que causaram erro:', event.data);
                    }
                };

                this.eventSource.onerror = (event) => {
                    console.error('DEBUG: Erro na conexão SSE:', event);
                    console.log('DEBUG: Estado da conexão SSE:', this.eventSource.readyState);
                    console.log('DEBUG: Tipo de erro:', event.type);

                    // Reconecta após 5 segundos se houver erro
                    setTimeout(() => {
                        if (this.eventSource.readyState === EventSource.CLOSED) {
                            console.log('DEBUG: Reconectando SSE...');
                            this.setupSSE();
                        }
                    }, 5000);
                };

                console.log('DEBUG: SSE configurado para logs em tempo real');

            } catch (error) {
                console.error('DEBUG: Erro ao criar EventSource:', error);
                this.startPollingFallback();
            }
        } else {
            console.warn('SSE não suportado, usando fallback');
            this.startPollingFallback();
        }
    }

    startSmartPolling() {
        console.log('DEBUG: Iniciando polling inteligente...');
        this.lastLogCount = this.logs.length;

        // Polling mais frequente para detectar novos logs rapidamente
        this.pollingInterval = setInterval(() => {
            this.checkForNewLogs();
        }, 2000); // Verifica a cada 2 segundos
    }

    async checkForNewLogs() {
        try {
            const response = await fetch('/api/logs');
            if (response.ok) {
                const data = await response.json();
                if (data.logs && data.logs.length > this.lastLogCount) {
                    console.log('DEBUG: Novos logs detectados!', data.logs.length, 'vs', this.lastLogCount);

                    // Adiciona apenas os novos logs
                    const newLogs = data.logs.slice(this.lastLogCount);
                    newLogs.forEach(log => this.addLogFromAPI(log));

                    this.lastLogCount = data.logs.length;
                }
            }
        } catch (error) {
            console.error('Erro ao verificar novos logs:', error);
        }
    }

    startPollingFallback() {
        // Fallback para navegadores que não suportam SSE
        setInterval(() => {
            this.fetchLogs();
        }, 5000); // Polling menos frequente como fallback
    }
    
    async loadInitialLogs() {
        // Carrega logs iniciais apenas uma vez
        console.log('DEBUG: Carregando logs iniciais...');
        try {
            const response = await fetch('/api/logs');
            if (response.ok) {
                const data = await response.json();
                if (data.logs && data.logs.length > 0) {
                    // Limpa logs existentes e adiciona os novos
                    this.logs = [];
                    this.logsContainer.innerHTML = '';
                    this.totalLogs = 0;
                    this.errorCount = 0;
                    this.successCount = 0;

                    data.logs.forEach(log => this.addLogFromAPI(log));
                    console.log('DEBUG: Logs iniciais carregados:', this.logs.length);
                }
            }
        } catch (error) {
            console.error('Erro ao carregar logs iniciais:', error);
        }
    }

    async fetchLogs() {
        // Mantido apenas para fallback
        await this.loadInitialLogs();
    }

    addLogFromAPI(logData) {
        const logEntry = {
            timestamp: logData.timestamp,
            content: logData.content,
            type: logData.type,
            id: Date.now() + Math.random()
        };

        this.logs.push(logEntry);
        this.updateStats(logEntry.type);
        this.renderLog(logEntry);
        this.scrollToBottom();
    }

    addLogFromSSE(logData) {
        console.log('DEBUG: Adicionando log via SSE:', logData);

        // Adiciona log recebido via SSE (tempo real)
        const logEntry = {
            timestamp: logData.timestamp,
            content: logData.content,
            type: logData.type,
            id: Date.now() + Math.random()
        };

        this.logs.push(logEntry);
        this.updateStats(logEntry.type);
        this.renderLog(logEntry);
        this.scrollToBottom();

        console.log('DEBUG: Log adicionado à interface. Total de logs:', this.logs.length);

        // Mantém apenas os últimos 1000 logs no frontend
        if (this.logs.length > 1000) {
            this.logs = this.logs.slice(-1000);
            this.renderAllLogs();
        }
    }
    
    addLog(logData) {
        const timestamp = logData.timestamp || new Date().toISOString();
        const content = logData.content || logData.message || logData;
        const type = this.determineLogType(content);
        
        const logEntry = {
            timestamp,
            content,
            type,
            id: Date.now() + Math.random()
        };
        
        this.logs.push(logEntry);
        this.updateStats(type);
        this.renderLog(logEntry);
        this.scrollToBottom();
        
        // Manter apenas os últimos 1000 logs
        if (this.logs.length > 1000) {
            this.logs = this.logs.slice(-1000);
            this.renderAllLogs();
        }
    }
    
    determineLogType(content) {
        const contentLower = content.toLowerCase();
        if (contentLower.includes('erro') || contentLower.includes('error') || contentLower.includes('failed')) {
            return 'error';
        } else if (contentLower.includes('sucesso') || contentLower.includes('success') || contentLower.includes('200')) {
            return 'success';
        }
        return 'info';
    }
    
    updateStats(type) {
        this.totalLogs++;
        this.lastLogTime = new Date();
        
        if (type === 'error') {
            this.errorCount++;
        } else if (type === 'success') {
            this.successCount++;
        }
        
        this.updateStatsDisplay();
    }
    
    updateStatsDisplay() {
        this.totalLogsElement.textContent = this.totalLogs;
        this.errorCountElement.textContent = this.errorCount;
        this.successCountElement.textContent = this.successCount;
        
        if (this.lastLogTime) {
            const timeStr = this.lastLogTime.toLocaleTimeString('pt-BR', {
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit'
            });
            this.lastLogTimeElement.textContent = timeStr;
        }
    }
    
    renderLog(logEntry) {
        console.log('DEBUG: Renderizando log:', logEntry);

        const logElement = document.createElement('div');
        logElement.className = `log-entry ${logEntry.type}`;
        logElement.innerHTML = `
            <div class="log-timestamp">[${this.formatTimestamp(logEntry.timestamp)}]</div>
            <div class="log-content">${this.escapeHtml(logEntry.content)}</div>
        `;

        this.logsContainer.appendChild(logElement);
        console.log('DEBUG: Log adicionado ao DOM. Total de elementos:', this.logsContainer.children.length);
    }
    
    renderAllLogs() {
        this.logsContainer.innerHTML = '';
        this.logs.forEach(log => this.renderLog(log));
        this.scrollToBottom();
    }
    
    formatTimestamp(timestamp) {
        const date = new Date(timestamp);
        return date.toLocaleString('pt-BR', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });
    }
    
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
    
    scrollToBottom() {
        this.logsContainer.scrollTop = this.logsContainer.scrollHeight;
    }
    
    async clearLogs() {
        try {
            const response = await fetch('/api/logs/clear', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                }
            });

            if (response.ok) {
                this.logs = [];
                this.totalLogs = 0;
                this.errorCount = 0;
                this.successCount = 0;
                this.lastLogTime = null;

                this.logsContainer.innerHTML = '';
                this.updateStatsDisplay();
            }
        } catch (error) {
            console.error('Erro ao limpar logs:', error);
        }
    }

    // Método para parar o polling quando necessário
    disconnect() {
        if (this.pollingInterval) {
            clearInterval(this.pollingInterval);
            console.log('Polling parado');
        }
        if (this.eventSource) {
            this.eventSource.close();
            console.log('Conexão SSE fechada');
        }
    }
    
    // Método para adicionar logs de exemplo (para demonstração)
    addSampleLogs() {
        const sampleLogs = [
            { content: "[2025-07-02T18:20:07.6132] GET /status - IP: ::1", type: "info" },
            { content: "[2025-07-02T18:20:07.6132] GET /status - Status: 200", type: "success" },
            { content: "[2025-07-02T18:21:55.9712] POST /escreve_arquivo - IP: ::1", type: "info" }
        ];
        
        sampleLogs.forEach((log, index) => {
            setTimeout(() => {
                this.addLog(log);
            }, index * 500);
        });
    }
}

// Função global para limpar logs
async function clearLogs() {
    if (window.logsApp) {
        await window.logsApp.clearLogs();
    }
}

// Detecta se está rodando como app do Chrome
function isRunningAsApp() {
    return window.navigator.standalone ||
           window.matchMedia('(display-mode: standalone)').matches ||
           document.referrer.includes('android-app://') ||
           window.location.search.includes('app=true');
}

// Logs básicos para verificar se o JavaScript está carregando
console.log('=== JAVASCRIPT CARREGADO ===');
console.log('Timestamp:', new Date().toISOString());

// Inicializar a aplicação quando o DOM estiver carregado
document.addEventListener('DOMContentLoaded', () => {
    console.log('=== DOM CARREGADO ===');

    try {
        console.log('=== CRIANDO LogsApp ===');
        window.logsApp = new LogsApp();
        console.log('=== LogsApp CRIADO COM SUCESSO ===');

        // Se está rodando como app, adiciona classe especial
        if (isRunningAsApp()) {
            document.body.classList.add('app-mode');
            console.log('Rodando como aplicativo desktop');
        }

        // Previne menu de contexto para simular app nativo
        document.addEventListener('contextmenu', (e) => {
            if (isRunningAsApp()) {
                e.preventDefault();
            }
        });

        // Previne seleção de texto para simular app nativo
        document.addEventListener('selectstart', (e) => {
            if (isRunningAsApp() && !e.target.matches('input, textarea')) {
                e.preventDefault();
            }
        });

    } catch (error) {
        console.error('=== ERRO AO INICIALIZAR ===', error);
    }
});

// Fechar conexão SSE quando a página for fechada
window.addEventListener('beforeunload', () => {
    if (window.logsApp) {
        window.logsApp.disconnect();
    }
});
