package main

import (
	"fmt"
	"net"
)

// getLocalIPs 获取本机IP地址
func getLocalIPs() {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("获取网络接口失败: %v\n", err)
		return
	}

	for _, iface := range interfaces {
		// 跳过回环接口和未启用的接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// 获取接口地址
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// 只显示IPv4地址
			if ip != nil && ip.To4() != nil {
				fmt.Printf("  %s (%s)\n", ip.String(), iface.Name)
			}
		}
	}
}

// getFloat64 安全地获取float64值
func getFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0.0
	}
}

// getAccuracyDescription 获取精度描述
func getAccuracyDescription(accuracy int) string {
	switch accuracy {
	case 0:
		return "不可靠"
	case 1:
		return "低精度"
	case 2:
		return "中等精度"
	case 3:
		return "高精度"
	default:
		return "未知"
	}
}
