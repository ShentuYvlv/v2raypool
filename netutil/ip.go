package netutil

import (
	"fmt"
	"net"
	"time"
)

func IsPrivateIp(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16", // 链路本地地址
		"127.0.0.0/8",    // 本地环回地址
		"127.0.0.1/32",
		"::1/128",   // IPv6本地环回地址
		"fe80::/10", // IPv6链路本地地址
	}
	for _, r := range privateRanges {
		_, network, _ := net.ParseCIDR(r)
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// IsPortInUse 检测端口是否被占用
func IsPortInUse(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second*2)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// FindAvailablePort 从指定端口开始查找可用端口
func FindAvailablePort(startPort int) int {
	for port := startPort; port < startPort+1000; port++ {
		if !IsPortInUse(port) {
			return port
		}
	}
	return startPort // 如果找不到可用端口，返回原始端口
}
