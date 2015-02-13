package main

import (
    "net/http"
    "strconv"
    "time"
    "fmt"
    "os"
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

func main() {
    log.Println("ospid:"+strconv.Itoa(os.Getpid()))
    http.HandleFunc("/", HelloServer)

    goreload.SingleHTTPService("0.0.0.0:58081", http.DefaultServeMux)
}
