package vasinstruments

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"vas2/general"

	"fyne.io/fyne/v2"
)

// MyTSIscanner söker efter AeroTrak (port 502) och DustTrak (port 3602)
func MyTSIscanner() error {
	myIPs := general.GetcurrentIP()
	if len(myIPs) == 0 {
		return errors.New("#1 MyTSIscanner: no current IP address exists")
	}

	var foundCount int32
	var wg sync.WaitGroup
	timeout := 300 * time.Millisecond

	// Semafor för att max ha 30 samtida nätverksanslutningar igång (skonar routern/OS)
	maxConcurrent := make(chan struct{}, 30)

	// Hållare för hittade adresser (skyddas med mutex för att undvika data race)
	var mu sync.Mutex
	var foundAeroTrak, foundDustTrak string

	for _, ipStr := range myIPs {
		lastDot := strings.LastIndex(ipStr, ".")
		if lastDot == -1 {
			continue
		}
		subnet := ipStr[:lastDot+1]

		log.Println("Searching IP subnet:", subnet+"x")

		for i := 1; i < 255; i++ {
			wg.Add(1)

			go func(hostNum int) {
				defer wg.Done()

				// Vänta på ledig plats i semaforen
				maxConcurrent <- struct{}{}
				defer func() { <-maxConcurrent }()

				ip := fmt.Sprintf("%s%d", subnet, hostNum)

				// 1. Kolla AeroTrak (Port 502)
				adr502 := fmt.Sprintf("%s:502", ip)
				conn, err := net.DialTimeout("tcp", adr502, timeout)
				if err == nil {
					_ = conn.Close()
					atomic.AddInt32(&foundCount, 1)

					mu.Lock()
					foundAeroTrak = adr502
					mu.Unlock()

					log.Println("Found AeroTrak at " + adr502)
					return
				}

				// 2. Kolla DustTrak (Port 3602)
				adr3602 := fmt.Sprintf("%s:3602", ip)
				conn, err = net.DialTimeout("tcp", adr3602, timeout)
				if err == nil {
					_ = conn.Close()
					atomic.AddInt32(&foundCount, 1)

					mu.Lock()
					foundDustTrak = adr3602
					mu.Unlock()

					log.Println("Found DustTrak at " + adr3602)
				}
			}(i)
		}
	}

	// Vänta tills alla scanningar är klara
	wg.Wait()

	// Hitta Fyne app säkert
	app := fyne.CurrentApp()
	// Spara till Preferences säkert på huvudtråden EFTER att alla gorutiner är klara
	if app != nil {
		if foundAeroTrak != "" {
			app.Preferences().SetString("AeroTrak", foundAeroTrak)
		}
		if foundDustTrak != "" {
			app.Preferences().SetString("DustTrak", foundDustTrak)
		}
	}

	if atomic.LoadInt32(&foundCount) == 0 {
		return errors.New("no TSI instruments found on network")
	}

	return nil
}
