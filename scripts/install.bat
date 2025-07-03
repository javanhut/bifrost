@echo off
REM Bifrost Installation Script for Windows

setlocal enabledelayedexpansion

REM Configuration
set BIFROST_VERSION=0.1.0
set INSTALL_DIR=C:\Program Files\Bifrost
set BIFROST_HOME=%USERPROFILE%\.carrion

echo Bifrost Package Manager Installer for Windows
echo =============================================
echo.

REM Check if running as administrator
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: This script must be run as Administrator
    echo Right-click and select "Run as administrator"
    pause
    exit /b 1
)

echo System: Windows
echo Install directory: %INSTALL_DIR%
echo.

REM Check if Go is installed
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Go is not installed
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

echo Found Go installation
go version
echo.

REM Create temporary directory
set TEMP_DIR=%TEMP%\bifrost_install_%RANDOM%
mkdir "%TEMP_DIR%"
cd /d "%TEMP_DIR%"

REM Try to download binary
echo Downloading Bifrost %BIFROST_VERSION%...
set DOWNLOAD_URL=https://github.com/javanhut/bifrost/releases/download/v%BIFROST_VERSION%/bifrost-%BIFROST_VERSION%-windows-amd64.zip

powershell -Command "& {try { Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile 'bifrost.zip' -ErrorAction Stop } catch { exit 1 }}"
if %errorlevel% neq 0 (
    echo Binary not available for download. Building from source...
    goto build_from_source
)

REM Extract binary
echo Extracting binary...
powershell -Command "& {Expand-Archive -Path 'bifrost.zip' -DestinationPath '.' -Force}"
if exist "bifrost.exe" (
    goto install_binary
) else (
    echo Extraction failed. Building from source...
    goto build_from_source
)

:build_from_source
echo Building Bifrost from source...
git clone https://github.com/javanhut/bifrost.git
if %errorlevel% neq 0 (
    echo Error: Failed to clone repository
    pause
    exit /b 1
)

cd bifrost
go build -o bifrost.exe ./cmd/bifrost
if %errorlevel% neq 0 (
    echo Error: Build failed
    pause
    exit /b 1
)

move bifrost.exe "%TEMP_DIR%\bifrost.exe"
cd /d "%TEMP_DIR%"

:install_binary
echo Installing Bifrost...

REM Create install directory
if not exist "%INSTALL_DIR%" (
    mkdir "%INSTALL_DIR%"
)

REM Copy binary
copy /y bifrost.exe "%INSTALL_DIR%\bifrost.exe"
if %errorlevel% neq 0 (
    echo Error: Failed to copy binary
    pause
    exit /b 1
)

REM Add to PATH
echo Adding Bifrost to PATH...
setx PATH "%PATH%;%INSTALL_DIR%" /M >nul 2>&1
if %errorlevel% neq 0 (
    echo Warning: Could not add to system PATH. Please add manually:
    echo %INSTALL_DIR%
)

REM Setup directories
echo Setting up Bifrost directories...
if not exist "%BIFROST_HOME%" (
    mkdir "%BIFROST_HOME%"
)
if not exist "%BIFROST_HOME%\packages" (
    mkdir "%BIFROST_HOME%\packages"
)
if not exist "%BIFROST_HOME%\cache" (
    mkdir "%BIFROST_HOME%\cache"
)
if not exist "%BIFROST_HOME%\registry" (
    mkdir "%BIFROST_HOME%\registry"
)

echo Created Bifrost home directory: %BIFROST_HOME%

REM Test installation
echo.
echo Testing installation...
"%INSTALL_DIR%\bifrost.exe" version
if %errorlevel% equ 0 (
    echo.
    echo Installation complete!
    echo.
    echo To get started:
    echo   bifrost init       # Create a new Carrion package
    echo   bifrost --help     # Show available commands
    echo.
    echo Note: You may need to restart your command prompt for PATH changes to take effect.
) else (
    echo Error: Installation test failed
    pause
    exit /b 1
)

REM Cleanup
cd /d %TEMP%
rmdir /s /q "%TEMP_DIR%"

echo.
echo Press any key to exit...
pause >nul