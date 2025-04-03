# install.ps1
param (
    [string]$RepoOwner = "Eutika",
    [string]$RepoName = "eu-missions-cli",
    [string]$OutputDir = "$env:USERPROFILE\.local\bin"
)

# Function to detect architecture
function Get-Architecture {
    if ([Environment]::Is64BitOperatingSystem) {
        return "x86_64"
    } else {
        return "i386"
    }
}

# Function to get the latest release
function Get-LatestRelease {
    $headers = @{
        "Accept" = "application/vnd.github+json"
        "X-GitHub-Api-Version" = "2022-11-28"
    }

    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest" -Headers $headers
        return $response
    } catch {
        Write-Error "Failed to fetch latest release: $_"
        exit 1
    }
}

# Function to download and extract archive
function Install-MissionsCLI {
  

    # Detect platform
    $os = "Windows"
    $arch = Get-Architecture

    # Create output directory
    if (-not (Test-Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir | Out-Null
    }

    # Get latest release
    Write-Host "Fetching latest release information..."
    $release = Get-LatestRelease

    # Find matching asset
    $asset = $release.assets | 
        Where-Object { 
            $_.name -match "Windows_$arch" -and 
            $_.name.EndsWith(".zip") 
        } | 
        Select-Object -First 1

    if (-not $asset) {
        Write-Error "No matching binary found for Windows $arch"
        exit 1
    }

    # Download asset
    $tempFile = Join-Path $env:TEMP $asset.name
    Write-Host "Downloading $($asset.name)..."
    
    $headers = @{
        "Accept" = "application/octet-stream"
        "X-GitHub-Api-Version" = "2022-11-28"
    }

    Invoke-WebRequest -Uri $asset.browser_download_url -Headers $headers -OutFile $tempFile

    # Create temporary extraction directory
    $tempDir = Join-Path $env:TEMP "missions-cli-install"
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    # Extract zip
    Expand-Archive -Path $tempFile -DestinationPath $tempDir -Force

    # Find the binary
    $binaryPath = Get-ChildItem -Path $tempDir -Recurse -Include "$RepoName.exe" | Select-Object -First 1

    if (-not $binaryPath) {
        Write-Error "Binary not found in archive"
        exit 1
    }

    # Copy to output directory
    $outputBinaryPath = Join-Path $OutputDir "missions.exe"
    Copy-Item -Path $binaryPath -Destination $outputBinaryPath -Force

    # Clean up
    Remove-Item -Path $tempFile -Force
    Remove-Item -Path $tempDir -Recurse -Force

    # Update PATH if needed
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if (-not $currentPath.Contains($OutputDir)) {
        $newPath = "$OutputDir;$currentPath"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Write-Host "Added $OutputDir to PATH. Please restart your terminal."
    }

    Write-Host "Successfully installed 'missions' CLI to $outputBinaryPath"
}

# Run the installation
Install-MissionsCLI
