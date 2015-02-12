package goreload

import (
    "fmt"
    "log"
    "net"
    "os"
    "os/exec"
    "os/signal"
    "syscall"
    "sync"
    "net/http"
    "strconv"
    "reflect"
)

const (
    Graceful = "graceful"
)

// Test whether an error is equivalent to net.errClosing as returned by
// Accept during a graceful exit.
func IsErrClosing(err error) bool {
    if opErr, ok := err.(*net.OpError); ok {
        err = opErr.Err
    }
    return "use of closed network connection" == err.Error()
}

// Allows for us to notice when the connection is closed.
type conn struct {
    net.Conn
    wg      *sync.WaitGroup
    isclose bool
    lock    sync.Mutex
}

func (c conn) Close() error {
    log.Printf("close %s", c.RemoteAddr())
    c.lock.Lock()
    defer c.lock.Unlock()
    err := c.Conn.Close()
    if !c.isclose && err == nil {
        c.wg.Done()
        c.isclose = true
    }
    return err
}

type stoppableListener struct {
    net.Listener
    wg      sync.WaitGroup
}

// restart cmd
var cmd *exec.Cmd

// listener lock
var lock sync.Mutex

// listener wait group
var listenerWaitGroup sync.WaitGroup

// listener object
var listeners map[uintptr]net.Listener

// extra files file descriptor default entry is 3
var fd uintptr = 3

func init() {
    listeners = make(map[uintptr]net.Listener)
    path, err := exec.LookPath(os.Args[0])
    if nil != err {
        log.Fatalf("gracefulRestart: Failed to launch, error: %v", err)
    }
    cmd = exec.Command(path, os.Args[1:]...)
    cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%d", Graceful, 1))
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
}

func newStoppable(l net.Listener) (sl *stoppableListener) {
    lock.Lock()
    defer lock.Unlock()

    sl = &stoppableListener{Listener: l}
    listeners[fd] = l
    fd++
    return
}

func (sl *stoppableListener) Accept() (c net.Conn, err error) {
    c, err = sl.Listener.Accept()
    if err != nil {
        return
    }
    sl.wg.Add(1)
    // Wrap the returned connection, so that we can observe when
    // it is closed.
    c = conn{Conn: c, wg: &sl.wg}
    return
}

func (sl *stoppableListener) Close() error {
    log.Printf("close listener: %s", sl.Addr())
    return sl.Listener.Close()
}

// 等待信号
func Wait() {
    waitSignal()
    for _, listener := range (listeners) {
        listener.Close()
    }
    listenerWaitGroup.Wait()
    log.Println("close main process")
}

func waitSignal() error {
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGTERM, syscall.SIGHUP)
    for {
        sig := <-ch
        log.Println(sig.String())
        switch sig {

            case syscall.SIGTERM:
            return nil
            case syscall.SIGHUP:
            restart()
            return nil
        }
    }
    return nil // It'll never get here.
}

func restart() {
    for fd, listener := range (listeners) {
        // get listener fd
        cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%d", listener.Addr().String(), fd))
        v := reflect.ValueOf(listener).Elem().FieldByName("fd").Elem()
        originFD := uintptr(v.FieldByName("sysfd").Int())

        // entry i becomes file descriptor 3+i
        cmd.ExtraFiles = append(cmd.ExtraFiles, os.NewFile(
        originFD,
        listener.Addr().String(),
        ))
    }

    err := cmd.Start()
    if err != nil {
        log.Fatalf("gracefulRestart: Failed to launch, error: %v", err)
    }
}

func getInitListener(laddr string) (net.Listener, error) {
    var l net.Listener
    var err error
    listenerWaitGroup.Add(1)
    graceful := os.Getenv(Graceful)
    if graceful != "" {
        // get current file descriptor
        currFdStr := os.Getenv(laddr)
        currFd, err := strconv.Atoi(currFdStr)
        if err != nil {
            log.Printf("%s get fd fail: %v", laddr, err)
        }
        log.Printf("main: Listening to existing file descriptor %v.", currFd)
        f := os.NewFile(uintptr(currFd), "")
        // file listener dup fd
        l, err = net.FileListener(f)
        // close current file descriptor
        f.Close()
    } else {
        log.Printf("listen to %s.", laddr)
        l, err = net.Listen("tcp4", laddr)
    }
    return l, err
}

// socket service
func Serve(laddr string, handler func(net.Conn)) {
    l, err := getInitListener(laddr)
    if err != nil {
        log.Fatalf("start fail: %v", err)
    }
    theStoppable := newStoppable(l)
    serve(theStoppable, handler)
    log.Printf("%s wait all connection close...", laddr)
    theStoppable.wg.Wait()
    listenerWaitGroup.Done()
    log.Printf("close socket %s", laddr)
}

func serve(l net.Listener, handle func(net.Conn)) {
    defer l.Close()
    for {
        c, err := l.Accept()
        if nil != err {
            if IsErrClosing(err) {
                log.Println("error closing")
                return
            }
            log.Fatalln(err)
        }
        log.Println("handle client", c.RemoteAddr())
        handle(c)
    }
}

// HTTP service
func ListenAndServe(laddr string, handler http.Handler) {
    var err error
    var l net.Listener
    l, err = getInitListener(laddr)
    if err != nil {
        log.Fatalf("start fail: %v", err)
    }
    theStoppable := newStoppable(l)
    log.Printf("Serving on http://%s/", laddr)
    server := &http.Server{Handler: handler}
    err = server.Serve(theStoppable)
    if err != nil {
        log.Println("ListenAndServe: ", err)
    }
    log.Printf("%s wait all connection close...", laddr)
    theStoppable.wg.Wait()
    listenerWaitGroup.Done()
    log.Printf("close http %s", laddr)
}

// socket service
func SocketService(laddr string, handler func(net.Conn)) {
    go func() {
        Serve(laddr, handler)
    }()
}

// HTTP service
func HTTPService(laddr string, handler http.Handler) {
    go func() {
        ListenAndServe(laddr, handler)
    }()
}

// single HTTP service
func SingleHTTPService(laddr string, handler http.Handler) {
    HTTPService(laddr, handler)
    Wait()
}

// single socket service
func SingleSocketService(laddr string, handler func(net.Conn)) {
    SocketService(laddr, handler)
    Wait()
}