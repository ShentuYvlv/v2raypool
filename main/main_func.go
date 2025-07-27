package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/iotames/miniutils"
	vp "github.com/iotames/v2raypool"
	"github.com/iotames/v2raypool/conf"
	"github.com/iotames/v2raypool/webserver"
)

func runServer() {
	logStart()
	cf := conf.GetConf()
	checkInitPorts(cf)

	// 设置信号处理，确保程序退出时正确清理
	setupSignalHandler()

	webPort := cf.WebServerPort
	if webPort == 0 {
		vp.RunServer()
		return
	}
	go vp.RunServer()
	time.Sleep(time.Second * 1)
	s := webserver.NewWebServer(webPort)
	err := miniutils.StartBrowserByUrl(fmt.Sprintf(`http://127.0.0.1:%d`, webPort))
	if err != nil {
		fmt.Println("StartBrowserByUrl error: " + err.Error())
	}
	s.ListenAndServe()
}

// setupSignalHandler 设置信号处理，确保程序退出时正确清理
func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n收到退出信号，正在清理资源...")

		// 停止所有代理节点
		pp := vp.GetProxyPool()
		if pp != nil {
			fmt.Println("正在停止所有代理节点...")
			err := pp.StopAll()
			if err != nil {
				fmt.Printf("停止代理节点时出错: %v\n", err)
			} else {
				fmt.Println("所有代理节点已停止")
			}
		}

		fmt.Println("清理完成，程序退出")
		os.Exit(0)
	}()
}

func checkInitPorts(cf conf.Conf) {
	appGrpcPort := cf.GrpcPort
	v2rayApiPort := cf.V2rayApiPort
	webPort := cf.WebServerPort
	sysPort1 := cf.GetHttpProxyPort()
	sysPort2 := sysPort1 - 1
	startPort := sysPort1 + 1
	if isPortBeUsed(webPort) {
		panic(fmt.Errorf("proxy pool web port(%d) may already be in use", webPort))
	}
	if isPortBeUsed(appGrpcPort) {
		panic(fmt.Errorf("proxy pool grpc control port(%d) may already be in use", appGrpcPort))
	}
	if isPortBeUsed(v2rayApiPort) {
		panic(fmt.Errorf("proxy pool of v2ray api port(%d) may already be in use", v2rayApiPort))
	}
	if isPortBeUsed(sysPort1) {
		panic(fmt.Errorf("http proxy port(%d) may already be in use", sysPort1))
	}
	if isPortBeUsed(sysPort2) {
		panic(fmt.Errorf("socks5 proxy port(%d) may already be in use", sysPort1))
	}
	if isPortBeUsed(startPort) {
		panic(fmt.Errorf("proxy pool start port(%d) may already be in use", startPort))
	}
}

func logStart() {
	ntime := time.Now()
	f, err := os.OpenFile("startat.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString(ntime.Format(time.RFC3339)) // "2006-01-02T15:04:05Z07:00"
	if err != nil {
		panic(err)
	}
	logsdir := filepath.Join(vconf.RuntimeDir, "logs")
	if !miniutils.IsPathExists(logsdir) {
		err = miniutils.Mkdir(logsdir)
		if err != nil {
			panic(err)
		}
	}
	lgpath := filepath.Join(logsdir, "start.log")
	f, err = os.OpenFile(lgpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		panic(err)
	}
	timestr := ntime.Format("[2006-01-02 15:04:05]")
	logmsg := fmt.Sprintf("\n%s: envFile(%s), MainGrpcPort(%d), V2rayApiPort(%d)", timestr, vconf.EnvFile, vconf.GrpcPort, vconf.V2rayApiPort)
	_, err = f.WriteString(logmsg)
	if err != nil {
		panic(err)
	}
}

// 检查端口是否被占用。被占用true, 未被占用 false
func isPortBeUsed(port int) bool {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true
	}
	defer l.Close()
	return false
}
