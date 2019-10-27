package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

var timeout = 1000
var maxParallelNum = 10
var host string

func usage(cmd string) {
	fmt.Printf(`
%s [-timeout <int>] [-max <int>] host port [port]
	`, cmd)
}

func main() {
	flag.IntVar(&timeout, "timeout", timeout, "millisecond")
	flag.IntVar(&maxParallelNum, "max", maxParallelNum, "max parallel number")
	flag.Parse()
	if len(flag.Args()) < 2 {
		usage(os.Args[0])
		os.Exit(1)
	}
	var (
		host      string
		startPort int
		endPort   int
	)
	host = flag.Arg(0)
	var err error
	startPort, err = strconv.Atoi(flag.Arg(1))
	if err != nil {
		fmt.Println("port is not number")
		os.Exit(1)
	}
	if len(flag.Args()) > 2 {
		endPort, err = strconv.Atoi(flag.Arg(2))
		if err != nil {
			fmt.Println("port is not number")
			os.Exit(1)
		}
		if endPort < startPort {
			fmt.Println("port range error")
			os.Exit(1)
		}
	} else {
		endPort = startPort
	}
	if maxParallelNum <= 0 {
		maxParallelNum = 1
	}
	ports := make([]int, 0)
	var wg sync.WaitGroup
	ch := make(chan bool, maxParallelNum)
	var mutex sync.Mutex
	for port := startPort; port <= endPort; port++ {
		wg.Add(1)
		ch <- true
		go func(host string, port int, ch chan bool, wg *sync.WaitGroup) {
			defer func() {
				wg.Done()
				<-ch
			}()
			if isOpen(host, port, timeout) {
				mutex.Lock()
				ports = append(ports, port)
				mutex.Unlock()
			}
		}(host, port, ch, &wg)
	}
	wg.Wait()
	close(ch)
	if len(ports) == 0 {
		fmt.Println("no port is opened")
		return
	}
	sort.Ints(ports)
	for _, port := range ports {
		fmt.Printf("%d is opened\n", port)
	}
}

func isOpen(host string, port int, timeout int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Duration(timeout)*time.Millisecond)
	if err == nil {
		_ = conn.Close()
		return true
	}
	return false
}
