@echo off
echo 测试传感器日志服务器配置功能
echo ================================

echo.
echo 1. 测试默认配置...
go test -v -run TestLoadConfig

echo.
echo 2. 测试环境变量配置...
go test -v -run TestLoadFromEnv

echo.
echo 3. 测试配置验证...
go test -v -run TestValidateConfig

echo.
echo 4. 测试.env文件加载...
go test -v -run TestLoadEnvFile

echo.
echo 5. 测试服务器地址生成...
go test -v -run TestGetServerAddr

echo.
echo 配置功能测试完成！
pause 