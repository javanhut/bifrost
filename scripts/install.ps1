# Bifrost Installation Script for Windows PowerShell

param(
    [string]$Version = "0.1.0",
    [string]$InstallDir = "$env:LOCALAPPDATA\Bifrost",
    [switch]$Global,
    [switch]$Force
)

# Configuration
$BifrostVersion = $Version
$BifrostHome = "$env:USERPROFILE\.carrion"

if ($Global) {
    $InstallDir = "$env:ProgramFiles\Bifrost"
}

# Functions
function Write-ColoredText {
    param(
        [string]$Text,
        [string]$Color = "White"
    )
    Write-Host $Text -ForegroundColor $Color
}

function Write-Error {
    param([string]$Message)
    Write-ColoredText "Error: $Message" "Red"
}

function Write-Success {
    param([string]$Message)
    Write-ColoredText $Message "Green"
}

function Write-Info {
    param([string]$Message)
    Write-ColoredText $Message "Yellow"
}

function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = [Security.Principal.WindowsPrincipal]$currentUser
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Test-GoInstalled {
    try {
        $null = Get-Command go -ErrorAction Stop
        $goVersion = & go version
        Write-Info "Found Go: $goVersion"
        return $true
    }
    catch {
        Write-Error "Go is not installed. Please install Go from https://golang.org/dl/"
        return $false
    }
}

function Download-Binary {
    param(
        [string]$Url,
        [string]$OutputPath
    )
    
    try {
        Write-Info "Downloading from: $Url"
        Invoke-WebRequest -Uri $Url -OutFile $OutputPath -ErrorAction Stop
        return $true
    }
    catch {
        Write-Info "Download failed: $($_.Exception.Message)"
        return $false
    }
}

function Build-FromSource {
    $tempDir = [System.IO.Path]::GetTempPath()
    $buildDir = Join-Path $tempDir "bifrost-build-$(Get-Random)"
    
    try {
        Write-Info "Creating build directory: $buildDir"
        New-Item -ItemType Directory -Path $buildDir -Force | Out-Null
        Set-Location $buildDir
        
        Write-Info "Cloning Bifrost repository..."
        & git clone https://github.com/javanhut/bifrost.git
        if ($LASTEXITCODE -ne 0) {
            throw "Git clone failed"
        }
        
        Set-Location "bifrost"
        
        Write-Info "Building Bifrost..."
        & go build -ldflags="-s -w" -o "bifrost.exe" "./cmd/bifrost"
        if ($LASTEXITCODE -ne 0) {
            throw "Build failed"
        }
        
        $binaryPath = Join-Path (Get-Location) "bifrost.exe"
        if (-not (Test-Path $binaryPath)) {
            throw "Binary not found after build"
        }
        
        return $binaryPath
    }
    catch {
        Write-Error "Build from source failed: $($_.Exception.Message)"
        return $null
    }
    finally {
        if (Test-Path $buildDir) {
            Remove-Item -Path $buildDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

function Install-Binary {
    param(
        [string]$BinaryPath,
        [string]$DestinationDir
    )
    
    try {
        Write-Info "Installing Bifrost to: $DestinationDir"
        
        # Create destination directory
        if (-not (Test-Path $DestinationDir)) {
            New-Item -ItemType Directory -Path $DestinationDir -Force | Out-Null
        }
        
        # Copy binary
        $destinationPath = Join-Path $DestinationDir "bifrost.exe"
        Copy-Item -Path $BinaryPath -Destination $destinationPath -Force
        
        # Verify installation
        if (Test-Path $destinationPath) {
            Write-Success "Binary installed successfully!"
            return $true
        }
        else {
            Write-Error "Installation failed - binary not found"
            return $false
        }
    }
    catch {
        Write-Error "Installation failed: $($_.Exception.Message)"
        return $false
    }
}

function Add-ToPath {
    param(
        [string]$Directory,
        [switch]$Global
    )
    
    try {
        $target = if ($Global) { "Machine" } else { "User" }
        
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", $target)
        
        if ($currentPath -notlike "*$Directory*") {
            $newPath = "$currentPath;$Directory"
            [Environment]::SetEnvironmentVariable("PATH", $newPath, $target)
            Write-Success "Added to PATH ($target scope)"
            return $true
        }
        else {
            Write-Info "Directory already in PATH"
            return $true
        }
    }
    catch {
        Write-Error "Failed to add to PATH: $($_.Exception.Message)"
        return $false
    }
}

function Setup-Directories {
    param(
        [string]$HomeDir
    )
    
    try {
        Write-Info "Setting up Bifrost directories..."
        
        $dirs = @(
            $HomeDir,
            "$HomeDir\packages",
            "$HomeDir\cache",
            "$HomeDir\registry"
        )
        
        foreach ($dir in $dirs) {
            if (-not (Test-Path $dir)) {
                New-Item -ItemType Directory -Path $dir -Force | Out-Null
            }
        }
        
        Write-Success "Created Bifrost home directory: $HomeDir"
        return $true
    }
    catch {
        Write-Error "Failed to setup directories: $($_.Exception.Message)"
        return $false
    }
}

function Test-Installation {
    param(
        [string]$BinaryPath
    )
    
    try {
        Write-Info "Testing installation..."
        $output = & $BinaryPath version 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Installation test passed!"
            Write-Host $output
            return $true
        }
        else {
            Write-Error "Installation test failed"
            return $false
        }
    }
    catch {
        Write-Error "Installation test failed: $($_.Exception.Message)"
        return $false
    }
}

# Main installation logic
Write-Host "Bifrost Package Manager Installer for Windows" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""

# Check administrator rights if global install
if ($Global -and -not (Test-Administrator)) {
    Write-Error "Global installation requires administrator privileges"
    Write-Info "Please run PowerShell as Administrator or use -Global:`$false for user installation"
    exit 1
}

# Check Go installation
if (-not (Test-GoInstalled)) {
    exit 1
}

Write-Info "System: Windows"
Write-Info "Install directory: $InstallDir"
Write-Info "Global installation: $Global"
Write-Host ""

# Create temporary directory
$tempDir = [System.IO.Path]::GetTempPath()
$workDir = Join-Path $tempDir "bifrost-install-$(Get-Random)"
New-Item -ItemType Directory -Path $workDir -Force | Out-Null

try {
    Set-Location $workDir
    
    # Try to download binary first
    $downloadUrl = "https://github.com/javanhut/bifrost/releases/download/v$BifrostVersion/bifrost-$BifrostVersion-windows-amd64.zip"
    $zipPath = Join-Path $workDir "bifrost.zip"
    
    $binaryPath = $null
    
    if (Download-Binary -Url $downloadUrl -OutputPath $zipPath) {
        Write-Info "Extracting binary..."
        Expand-Archive -Path $zipPath -DestinationPath $workDir -Force
        $binaryPath = Join-Path $workDir "bifrost.exe"
        
        if (-not (Test-Path $binaryPath)) {
            Write-Info "Binary not found in archive. Building from source..."
            $binaryPath = Build-FromSource
        }
    }
    else {
        Write-Info "Binary download failed. Building from source..."
        $binaryPath = Build-FromSource
    }
    
    if (-not $binaryPath -or -not (Test-Path $binaryPath)) {
        Write-Error "Failed to obtain Bifrost binary"
        exit 1
    }
    
    # Install binary
    if (-not (Install-Binary -BinaryPath $binaryPath -DestinationDir $InstallDir)) {
        exit 1
    }
    
    # Add to PATH
    if (-not (Add-ToPath -Directory $InstallDir -Global:$Global)) {
        Write-Info "Please add the following directory to your PATH manually:"
        Write-Info $InstallDir
    }
    
    # Setup directories
    if (-not (Setup-Directories -HomeDir $BifrostHome)) {
        exit 1
    }
    
    # Test installation
    $installedBinary = Join-Path $InstallDir "bifrost.exe"
    if (Test-Installation -BinaryPath $installedBinary) {
        Write-Host ""
        Write-Success "Installation complete!"
        Write-Host ""
        Write-Host "To get started:"
        Write-Host "  bifrost init       # Create a new Carrion package"
        Write-Host "  bifrost --help     # Show available commands"
        Write-Host ""
        Write-Info "Note: You may need to restart your PowerShell session for PATH changes to take effect."
    }
    else {
        exit 1
    }
}
finally {
    # Cleanup
    if (Test-Path $workDir) {
        Remove-Item -Path $workDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}