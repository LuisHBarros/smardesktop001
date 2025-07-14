# Go Desktop App

Aplicação desktop para Windows escrita em Go que opera em segundo plano com API RESTful e system tray.

## Funcionalidades

- **System Tray**: Ícone na bandeja do sistema com menu de contexto
- **API RESTful**: Servidor HTTP local na porta 8080
- **Operações de Arquivo**: Leitura e movimentação de arquivos
- **Execução de Processos**: Execução assíncrona de executáveis externos
- **Logging**: Sistema de logs das operações da API

## Estrutura do Projeto

```
/go-desktop-app
├── main.go                 # Ponto de entrada da aplicação
├── api/
│   ├── server.go           # Configuração do servidor HTTP
│   ├── handlers.go         # Handlers dos endpoints
│   └── middleware.go       # Middleware para CORS e logging
├── core/
│   ├── filesystem.go       # Operações de sistema de arquivos
│   └── executor.go         # Execução de processos externos
├── ui/
│   ├── tray.go             # Gerenciamento do system tray
│   └── logviewer.go        # Visualizador de logs
└── config/
    └── config.go           # Configurações e constantes
```

## Configuração

A aplicação utiliza os seguintes diretórios:
- **APP_DIR**: `C:\app\` - Diretório principal para operações de arquivo
- **ARCHIVE_DIR**: `C:\app\arquivo_morto\` - Diretório de destino para arquivos movidos
- **API_PORT**: `:8080` - Porta do servidor API

## Endpoints da API

### 1. Status da API
- **Endpoint**: `GET /status`
- **Descrição**: Verifica se a API está funcionando
- **Resposta**: `{"status": "API está funcionando"}`

### 2. Ler Arquivo
- **Endpoint**: `POST /escreve_arquivo`
- **Descrição**: Lê o conteúdo de um arquivo no diretório APP_DIR
- **Body**: `{"nome_arquivo": "exemplo.txt"}`
- **Resposta**: `{"conteudo": "conteúdo do arquivo"}`

### 3. Mover Arquivo
- **Endpoint**: `POST /move_arquivo`
- **Descrição**: Move um arquivo do APP_DIR para o ARCHIVE_DIR
- **Body**: `{"nome_arquivo": "exemplo.txt"}`
- **Resposta**: `{"mensagem": "Arquivo movido com sucesso para C:\\app\\arquivo_morto\\exemplo.txt"}`

### 4. Executar Processo
- **Endpoint**: `POST /executar_terceiros`
- **Descrição**: Executa um processo externo de forma assíncrona
- **Body**: `{"caminho_executavel": "C:\\caminho\\para\\programa.exe"}`
- **Resposta**: `{"mensagem": "Processo iniciado com sucesso"}`

## Como Usar

### 1. Compilação
```bash
go build -o go-desktop-app.exe
```

### 2. Execução
```bash
.\go-desktop-app.exe
```

### 3. System Tray
- A aplicação aparecerá na bandeja do sistema
- Clique com o botão direito no ícone para acessar o menu:
  - **Abrir Logs**: Exibe a janela de logs (se disponível)
  - **Sair**: Encerra a aplicação

### 4. Testando a API

#### Teste de Status
```powershell
Invoke-WebRequest -Uri "http://localhost:8080/status"
```

#### Teste de Leitura de Arquivo
```powershell
Invoke-WebRequest -Uri "http://localhost:8080/escreve_arquivo" -Method POST -Headers @{"Content-Type"="application/json"} -Body '{"nome_arquivo":"teste.txt"}'
```

#### Teste de Movimentação de Arquivo
```powershell
Invoke-WebRequest -Uri "http://localhost:8080/move_arquivo" -Method POST -Headers @{"Content-Type"="application/json"} -Body '{"nome_arquivo":"teste.txt"}'
```

#### Teste de Execução de Processo
```powershell
Invoke-WebRequest -Uri "http://localhost:8080/executar_terceiros" -Method POST -Headers @{"Content-Type"="application/json"} -Body '{"caminho_executavel":"C:\\Windows\\System32\\notepad.exe"}'
```

## Dependências

- `github.com/getlantern/systray` - Para o system tray
- `github.com/lxn/walk` - Para a interface gráfica (janela de logs)

## Notas

- A aplicação requer os diretórios `C:\app\` e `C:\app\arquivo_morto\` para funcionar corretamente
- Todas as operações da API são logadas no console e na janela de logs (quando disponível)
- O CORS está configurado para aceitar apenas requisições de localhost e 127.0.0.1
- A aplicação opera em segundo plano sem janela principal visível
