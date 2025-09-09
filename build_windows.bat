@echo off

REM go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo
go generate

go build -ldflags="-s -w -H=windowsgui"