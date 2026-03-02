@echo off
REM Check for protoc
where protoc >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo Error: protoc not found in PATH. Please install Protocol Buffers.
    echo Visit https://github.com/protocolbuffers/protobuf/releases
    pause
    exit /b 1
)

echo Generating gRPC code from proto/sync.proto...
mkdir internal\sync\proto 2>nul

protoc --go_out=. --go_opt=paths=source_relative ^
    --go-grpc_out=. --go-grpc_opt=paths=source_relative ^
    proto/sync.proto

if %ERRORLEVEL% eq 0 (
    echo Generation successful!
) else (
    echo Generation failed.
)
pause
