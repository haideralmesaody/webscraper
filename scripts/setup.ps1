# Create directories if they don't exist
$directories = @(
    "cmd",
    "internal/scraper",
    "internal/utils",
    "configs",
    "scripts",
    "docker",
    "test/testdata",
    "docs",
    "output",
    "logs",
    "temp_builds"
)

foreach ($dir in $directories) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force
        Write-Host "Created directory: $dir"
    }
}

# Move files to their new locations
$moves = @{
    "main.go" = "cmd/main.go"
    "scraper/scraper.go" = "internal/scraper/scraper.go"
    "utils/utils.go" = "internal/utils/utils.go"
    "utils/logger.go" = "internal/utils/logger.go"
    "utils/config.go" = "internal/utils/config.go"
    "Dockerfile" = "docker/Dockerfile"
}

foreach ($source in $moves.Keys) {
    $destination = $moves[$source]
    if (Test-Path $source) {
        Move-Item -Path $source -Destination $destination -Force
        Write-Host "Moved: $source -> $destination"
    }
} 