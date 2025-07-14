@echo off
setlocal

:: Verifica se está rodando como administrador
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo.
    echo ERRO: Este script precisa ser executado como Administrador!
    echo Clique com o botao direito e selecione "Executar como administrador"
    echo.
    pause
    exit /b 1
)

set APP_NAME=go-desktop-app.exe
set SERVICE_NAME=GoDesktopApp

echo.
echo ========================================
echo   Gerenciador de Servico - Go Desktop App
echo ========================================
echo.

if "%1"=="" goto menu

:: Executa comando direto se passado como parametro
goto execute_command

:menu
echo Escolha uma opcao:
echo.
echo 1. Instalar servico
echo 2. Desinstalar servico
echo 3. Iniciar servico
echo 4. Parar servico
echo 5. Status do servico
echo 6. Executar em modo interativo (com system tray)
echo 7. Sair
echo.
set /p choice="Digite sua opcao (1-7): "

if "%choice%"=="1" set command=install
if "%choice%"=="2" set command=remove
if "%choice%"=="3" set command=start
if "%choice%"=="4" set command=stop
if "%choice%"=="5" set command=status
if "%choice%"=="6" set command=interactive
if "%choice%"=="7" goto end

:execute_command
if "%1" neq "" set command=%1

if "%command%"=="install" goto install
if "%command%"=="remove" goto remove
if "%command%"=="start" goto start
if "%command%"=="stop" goto stop
if "%command%"=="status" goto status
if "%command%"=="interactive" goto interactive

echo Opcao invalida!
goto menu

:install
echo.
echo Instalando servico...
%APP_NAME% install
if %errorLevel% equ 0 (
    echo Servico instalado com sucesso!
    echo.
    echo Deseja iniciar o servico agora? (S/N)
    set /p start_now=""
    if /i "%start_now%"=="S" (
        echo Iniciando servico...
        %APP_NAME% start
    )
) else (
    echo Erro ao instalar servico!
)
goto end_or_menu

:remove
echo.
echo Parando servico (se estiver rodando)...
%APP_NAME% stop >nul 2>&1
echo Removendo servico...
%APP_NAME% remove
if %errorLevel% equ 0 (
    echo Servico removido com sucesso!
) else (
    echo Erro ao remover servico!
)
goto end_or_menu

:start
echo.
echo Iniciando servico...
%APP_NAME% start
goto end_or_menu

:stop
echo.
echo Parando servico...
%APP_NAME% stop
goto end_or_menu

:status
echo.
echo Consultando status do servico...
%APP_NAME% status
goto end_or_menu

:interactive
echo.
echo Executando em modo interativo...
echo Pressione Ctrl+C para parar ou feche pelo system tray
echo.
%APP_NAME%
goto end_or_menu

:end_or_menu
if "%1" neq "" goto end
echo.
echo Pressione qualquer tecla para voltar ao menu...
pause >nul
goto menu

:end
echo.
pause
