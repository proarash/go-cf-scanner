# Builds stripped cf-scanner binaries for every supported OS/architecture into out/
$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")
New-Item -ItemType Directory -Force out | Out-Null

# android 386/amd64/arm need cgo + the Android NDK, so only arm64 is built
$targets = @(
    "windows/amd64", "windows/386", "windows/arm64",
    "linux/amd64", "linux/386", "linux/arm64", "linux/arm",
    "darwin/amd64", "darwin/arm64",
    "android/arm64"
)

$env:CGO_ENABLED = "0"

foreach ($target in $targets) {
    $env:GOOS, $env:GOARCH = $target.Split("/")

    $os = if ($env:GOOS -eq "darwin") { "macos" } else { $env:GOOS }
    $arch = switch ($env:GOARCH) { "amd64" { "x64" } "386" { "x86" } default { $env:GOARCH } }
    $ext = if ($env:GOOS -eq "windows") { ".exe" } else { "" }

    $out = "out/cf-scanner-$os-$arch$ext"
    Write-Host "Building $out"
    go build -trimpath -buildvcs=false -ldflags="-s -w -buildid=" -o $out .
    if ($LASTEXITCODE -ne 0) { throw "build failed for $target" }
}
