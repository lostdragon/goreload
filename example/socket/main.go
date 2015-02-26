package main

import (
    "strconv"
    "os"
    "net"
    "log"

    "github.com/lostdragon/goreload"
)

func handler(conn net.Conn) {
    conn.Write([]byte("Hello, socket!"))
    conn.Close()
}

func main() {
    log.Println("ospid:"+strconv.Itoa(os.Getpid()))
    goreload.SingleTCPService("0.0.0.0:59081", handler)
}
