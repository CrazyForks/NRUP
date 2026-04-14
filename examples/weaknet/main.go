package main

import (
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/nyarime/nrup"
)

func main() {
	scenarios := []struct {
		name  string
		loss  string
		delay string
	}{
		{"正常网络", "0%", "0ms"},
		{"30%丢包", "30%", "50ms"},
		{"50%丢包", "50%", "50ms"},
		{"70%丢包", "70%", "50ms"},
	}

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  NRUP v1.4.2 弱网Benchmark (3次平均)")
	fmt.Printf("  %-12s | %8s | %8s | %6s | %6s\n",
		"场景", "送达率", "延迟", "FEC效", "Parity")
	fmt.Println("───────────────────────────────────────────────────────────────")

	for _, sc := range scenarios {
		exec.Command("tc", "qdisc", "del", "dev", "lo", "root").Run()
		time.Sleep(100 * time.Millisecond)
		if sc.loss != "0%" {
			exec.Command("tc", "qdisc", "add", "dev", "lo", "root", "netem",
				"loss", sc.loss, "delay", sc.delay).Run()
		}
		time.Sleep(200 * time.Millisecond)

		var totalSent, totalRecv int
		var totalFECEff float64
		var totalParity int
		runs := 3

		for r := 0; r < runs; r++ {
			sent, recv, stats := runTest(20)
			totalSent += sent
			totalRecv += recv
			totalFECEff += stats.FECEffectiveness
			totalParity += stats.CurrentParity
		}

		rate := float64(totalRecv) / float64(max(totalSent, 1)) * 100
		fmt.Printf("  %-12s | %3d/%-3d %3.0f%% | %6s | %5.1f%% | %5d\n",
			sc.name, totalRecv/runs, totalSent/runs, rate,
			"~", totalFECEff/float64(runs)*100, totalParity/runs)
	}

	exec.Command("tc", "qdisc", "del", "dev", "lo", "root").Run()
	fmt.Println("═══════════════════════════════════════════════════════════════")
}

func runTest(count int) (sent, recv int, stats nrup.ConnStats) {
	cfg := nrup.DefaultConfig()
	listener, err := nrup.Listen(":0", cfg)
	if err != nil { return }
	defer listener.Close()

	var mu sync.Mutex
	var serverRecv int
	go func() {
		conn, err := listener.Accept()
		if err != nil { return }
		defer conn.Close()
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if err != nil || n == 0 { return }
			mu.Lock()
			serverRecv++
			mu.Unlock()
		}
	}()

	conn, err := nrup.Dial(listener.Addr().String(), cfg)
	if err != nil { return }

	for i := 0; i < count; i++ {
		conn.Write([]byte(fmt.Sprintf("pkt-%04d-padding-data-for-fec", i)))
		time.Sleep(50 * time.Millisecond)
	}

	// 等待FEC恢复
	time.Sleep(3 * time.Second)

	stats = conn.Stats()
	conn.Close()

	sent = count
	mu.Lock()
	recv = serverRecv
	mu.Unlock()
	return
}

func max(a, b int) int { if a > b { return a }; return b }
