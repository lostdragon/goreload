package main

import (
    "net/http"
    "strconv"
    "time"
    "fmt"
    "os"
    "net"
    "log"

    "github.com/lostdragon/goreload"
)

func HelloServer(w http.ResponseWriter, req *http.Request) {
    t := req.FormValue("time")
    i, err := strconv.Atoi(t)
    if err == nil {
        time.Sleep(time.Duration(i) * time.Second)
        fmt.Fprintf(w, "ospid:"+strconv.Itoa(os.Getpid())+" time")
    } else {
        fmt.Fprintf(w, "Hello, http!")
    }
}

func handler(conn net.Conn) {
    if c, ok := conn.(*goreload.Conn); ok {
        if tcpConn, ok := c.Conn.(*net.TCPConn); ok {
            tcpConn.SetDeadline(time.Now().Add(time.Second * 10))
            tcpConn.Write([]byte("Hello, socket!"))
        } else {
            log.Fatalf("c.Conn is %T not *net.TCPConn", c.Conn)
        }
    } else {
        log.Fatalf("conn is %T not *goreload.Conn", conn)
    }
    // you should close conn
    conn.Close()
}

func main() {
    log.Println("ospid:"+strconv.Itoa(os.Getpid()))
    http.HandleFunc("/", HelloServer)

    goreload.HTTPService(":58081", http.DefaultServeMux)
    goreload.SocketService(":59081", handler)
    goreload.Wait()
}
