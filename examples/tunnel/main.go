package main

import (
	"crypto/sha256"
	"flag"
	"io"
	"log"
	"net"
	"sync"

	"github.com/nyarime/nrup"
)

// nrup-tunnel: 两台服务器之间的加密隧道
//
// 用法:
//   服务端: nrup-tunnel -mode server -listen :4000 -forward 127.0.0.1:3306 -password secret
//   客户端: nrup-tunnel -mode client -server 1.2.3.4:4000 -listen :13306 -password secret
//
// 效果: 本地:13306 ← NRUP加密隧道 → 远端:3306
//
// 场景:
//   MySQL/Redis/SSH 跨机房加密访问
//   游戏服务器间低延迟可靠通信
//   内网服务暴露（替代 frp/ngrok）

func main() {
	mode := flag.String("mode", "", "server / client")
	listen := flag.String("listen", "", "监听地址")
	server := flag.String("server", "", "服务端地址 (client模式)")
	forward := flag.String("forward", "", "转发目标 (server模式)")
	password := flag.String("password", "", "密码")
	cipher := flag.String("cipher", "auto", "加密: auto/none")
	disguise := flag.String("disguise", "none", "伪装: anyconnect/quic/none")
	flag.Parse()

	if *password == "" {
		log.Fatal("需要 -password")
	}

	switch *mode {
	case "server":
		if *listen == "" || *forward == "" {
			log.Fatal("server模式需要 -listen 和 -forward")
		}
		runServer(*listen, *forward, *password, *cipher, *disguise)
	case "client":
		if *server == "" || *listen == "" {
			log.Fatal("client模式需要 -server 和 -listen")
		}
		runClient(*server, *listen, *password, *cipher, *disguise)
	default:
		log.Fatal("需要 -mode server 或 -mode client")
	}
}

func makeCfg(password, cipher, disguise string) *nrup.Config {
	cfg := nrup.DefaultConfig()
	h := sha256.Sum256([]byte("nrup-tunnel:" + password))
	cfg.PSK = h[:]
	if cipher == "none" {
		cfg.Cipher = nrup.CipherNone
	}
	cfg.Disguise = disguise
	return cfg
}

// runServer NRUP监听 → 转发到本地服务
func runServer(listenAddr, forwardAddr, password, cipher, disguise string) {
	cfg := makeCfg(password, cipher, disguise)
	listener, err := nrup.Listen(listenAddr, cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Tunnel Server %s → %s", listenAddr, forwardAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go func() {
			defer conn.Close()
			local, err := net.Dial("tcp", forwardAddr)
			if err != nil {
				log.Printf("连接 %s 失败: %v", forwardAddr, err)
				return
			}
			defer local.Close()
			relay(conn, local)
		}()
	}
}

// runClient 本地TCP监听 → NRUP连接到远端
func runClient(serverAddr, listenAddr, password, cipher, disguise string) {
	cfg := makeCfg(password, cipher, disguise)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Tunnel Client %s → %s", listenAddr, serverAddr)

	for {
		local, err := ln.Accept()
		if err != nil {
			continue
		}
		go func() {
			defer local.Close()

			remoteCfg := makeCfg(password, cipher, disguise)
			remoteCfg.StreamMode = true // TCP转发用流模式
			remote, err := nrup.Dial(serverAddr, remoteCfg)
			if err != nil {
				log.Printf("NRUP连接失败: %v", err)
				return
			}
			defer remote.Close()
			relay(local, remote)
		}()
	}
	_ = cfg
}

func relay(a, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); io.Copy(a, b) }()
	go func() { defer wg.Done(); io.Copy(b, a) }()
	wg.Wait()
}
