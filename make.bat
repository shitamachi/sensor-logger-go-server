@echo off
setlocal enabledelayedexpansion

REM 设置变量
set BINARY_NAME=sensor-logger-server
set VERSION=v1.0.0
set BUILD_TIME=%date% %time%

REM 获取Git信息
for /f "tokens=*" %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%i
for /f "tokens=*" %%i in ('git rev-parse --abbrev-ref HEAD 2^>nul') do set GIT_BRANCH=%%i
if "%GIT_COMMIT%"=="" set GIT_COMMIT=unknown
if "%GIT_BRANCH%"=="" set GIT_BRANCH=unknown

REM 构建标志
set "LDFLAGS=-ldflags "-X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME% -X main.GitCommit=%GIT_COMMIT% -X main.GitBranch=%GIT_BRANCH%""

if "%1"=="" goto help
if "%1"=="help" goto help
if "%1"=="build" goto build
if "%1"=="build-windows" goto build-windows
if "%1"=="build-linux" goto build-linux
if "%1"=="build-all" goto build-all
if "%1"=="test" goto test
if "%1"=="test-coverage" goto test-coverage
if "%1"=="clean" goto clean
if "%1"=="run" goto run
if "%1"=="run-dev" goto run-dev
if "%1"=="fmt" goto fmt
if "%1"=="vet" goto vet
if "%1"=="info" goto info
if "%1"=="docker-build" goto docker-build
if "%1"=="docker-run" goto docker-run
if "%1"=="compose-up" goto compose-up
if "%1"=="compose-down" goto compose-down

echo 未知命令: %1
goto help

:help
echo 可用的命令:
echo   help           显示帮助信息
echo   build          构建应用程序
echo   build-windows  构建Windows版本
echo   build-linux    构建Linux版本
echo   build-all      构建所有平台版本
echo   test           运行测试
echo   test-coverage  运行测试并生成覆盖率报告
echo   clean          清理构建文件
echo   run            运行应用程序
echo   run-dev        以开发模式运行应用程序
echo   fmt            格式化代码
echo   vet            运行go vet检查
echo   info           显示项目信息
echo   docker-build   构建Docker镜像
echo   docker-run     运行Docker容器
echo   compose-up     启动Docker Compose服务
echo   compose-down   停止Docker Compose服务
goto end

:build
echo 构建应用程序...
go build -ldflags "-X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME% -X main.GitCommit=%GIT_COMMIT% -X main.GitBranch=%GIT_BRANCH%" -o %BINARY_NAME%.exe -v .
if %errorlevel% neq 0 (
    echo 构建失败
    exit /b 1
)
echo 构建完成: %BINARY_NAME%.exe
goto end

:build-windows
echo 构建Windows版本...
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME% -X main.GitCommit=%GIT_COMMIT% -X main.GitBranch=%GIT_BRANCH%" -o %BINARY_NAME%.exe -v .
if %errorlevel% neq 0 (
    echo 构建失败
    exit /b 1
)
echo 构建完成: %BINARY_NAME%.exe
goto end

:build-linux
echo 构建Linux版本...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME% -X main.GitCommit=%GIT_COMMIT% -X main.GitBranch=%GIT_BRANCH%" -o %BINARY_NAME%_unix -v .
if %errorlevel% neq 0 (
    echo 构建失败
    exit /b 1
)
echo 构建完成: %BINARY_NAME%_unix
goto end

:build-all
echo 构建所有平台版本...
call %0 build-windows
call %0 build-linux
echo 所有平台构建完成
goto end

:test
echo 运行测试...
go test -v ./...
goto end

:test-coverage
echo 运行测试并生成覆盖率报告...
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
echo 覆盖率报告已生成: coverage.html
goto end

:clean
echo 清理构建文件...
del /q %BINARY_NAME%.exe 2>nul
del /q %BINARY_NAME%_unix 2>nul
del /q %BINARY_NAME%_darwin 2>nul
del /q coverage.out 2>nul
del /q coverage.html 2>nul
echo 清理完成
goto end

:run
echo 运行应用程序...
go run .
goto end

:run-dev
echo 以开发模式运行应用程序...
set ENVIRONMENT=dev
go run .
goto end

:fmt
echo 格式化代码...
go fmt ./...
goto end

:vet
echo 运行go vet检查...
go vet ./...
goto end

:info
echo 项目信息:
echo   名称: %BINARY_NAME%
echo   版本: %VERSION%
echo   构建时间: %BUILD_TIME%
echo   Git提交: %GIT_COMMIT%
echo   Git分支: %GIT_BRANCH%
goto end

:docker-build
echo 构建Docker镜像...
docker build -t sensor-logger-server:latest .
goto end

:docker-run
echo 运行Docker容器...
docker run -p 18000:18000 --name sensor-logger-server sensor-logger-server:latest
goto end

:compose-up
echo 启动Docker Compose服务...
docker-compose up -d
goto end

:compose-down
echo 停止Docker Compose服务...
docker-compose down
goto end

:end
endlocal 