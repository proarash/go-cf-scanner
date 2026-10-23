#!/usr/bin/env sh
# Builds stripped cf-scanner binaries for every supported OS/architecture into out/
set -e
cd "$(dirname "$0")/.."
mkdir -p out

# android 386/amd64/arm need cgo + the Android NDK, so only arm64 is built
TARGETS="
windows/386 windows/amd64 windows/arm64
linux/386 linux/amd64 linux/arm linux/arm64 linux/loong64 linux/mips linux/mips64 linux/mips64le linux/mipsle linux/ppc64 linux/ppc64le linux/riscv64 linux/s390x
darwin/amd64 darwin/arm64
android/arm64
"

export CGO_ENABLED=0

for target in $TARGETS; do
    export GOOS="${target%/*}"
    export GOARCH="${target#*/}"

    os="$GOOS"; arch="$GOARCH"; ext=""
    [ "$os" = darwin ] && os=macos
    [ "$os" = windows ] && ext=".exe"
    [ "$arch" = amd64 ] && arch=x64
    [ "$arch" = 386 ] && arch=x86

    out="out/cf-scanner-$os-$arch$ext"
    echo "Building $out"
    go build -trimpath -buildvcs=false -ldflags="-s -w -buildid=" -o "$out" .
done
