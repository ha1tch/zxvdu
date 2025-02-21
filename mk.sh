#!/bin/bash
mkdir -p bin

echo "Building all binaries for zxvdu..."

BASENAME="zxvdu"
BINDIR="./bin"
SOURCES="main.go graphics.go buffers.go network.go commands.go handlers.go"


# Builds for some platforms are not yet supported 
# due to my momentary lack of a complete toolchain.
# If you know how to build this for Linux / RPi, etc. let me know!
# haitch@duck.com 
# https://oldbytes.social/@haitchfive 

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o $BINDIR/$BASENAME.win64.exe   $SOURCES
GOOS=windows GOARCH=386   go build -o $BINDIR/$BASENAME.win32.exe   $SOURCES

# Build for Linux
#GOOS=linux   GOARCH=amd64 go build -o $BINDIR/$BASENAME.linux64     $SOURCES
#GOOS=linux   GOARCH=386   go build -o $BINDIR/$BASENAME.linux32     $SOURCES

# Build for macOS (modern architectures)
GOOS=darwin  GOARCH=arm64 go build -o $BINDIR/$BASENAME.mac64.m1    $SOURCES
#GOOS=darwin  GOARCH=amd64 go build -o $BINDIR/$BASENAME.mac64.intel $SOURCES

# Build for Raspberry Pi
#GOOS=linux   GOARCH=arm   GOARM=6  go build -o $BINDIR/$BASENAME.rpi.arm6   $SOURCES  # Pi 1, Pi Zero
#GOOS=linux   GOARCH=arm   GOARM=7  go build -o $BINDIR/$BASENAME.rpi.arm7   $SOURCES  # Pi 2, Pi 3 (32-bit)
#GOOS=linux   GOARCH=arm64          go build -o $BINDIR/$BASENAME.rpi.arm64  $SOURCES  # Pi 3, Pi 4, Pi 5 (64-bit)

ls -l $BINDIR/$BASENAME*
