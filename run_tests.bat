@echo off
echo ================================
echo 运行传感器日志服务器测试套件
echo ================================
echo.

echo 1. 运行基本测试...
go test -v
if %errorlevel% neq 0 (
    echo 基本测试失败!
    exit /b 1
)
echo.

echo 2. 运行测试覆盖率分析...
go test -cover
echo.

echo 3. 运行基准测试...
go test -bench=. -benchmem
echo.

echo 4. 运行竞态检测...
go test -race
echo.

echo ================================
echo 所有测试完成!
echo ================================
pause 