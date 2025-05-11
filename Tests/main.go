package main

import (
	"net"
	"net/http"
	"runtime"

	"fmt"
)

// ----------------------------------------------------------------
func main() {
	fmt.Println("Go-Sync")
	fmt.Printf("Operating System : %s\n", runtime.GOOS)
	xip := fmt.Sprintf("%s", GetOutboundIP())
	port := "8080"

	fmt.Println("Server running....")
	fmt.Println("Listening on " + xip + ":" + port)
	if err := http.ListenAndServe(xip+":"+port, nil); err != nil {
		panic(err)
	}
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
