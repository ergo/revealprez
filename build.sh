#!/bin/sh
GOOS=windows   GOARCH=amd64 go build -o builds/revealprez_windows_amd64.exe
GOOS=linux     GOARCH=amd64 go build -o builds/revealprez_linux_amd64
GOOS=darwin    GOARCH=amd64 go build -o builds/revealprez_darwin_amd64
