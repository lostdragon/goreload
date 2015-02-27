#goreload

## Zero-downtime restarts in Go

The `goreload` support zero-downtime restarts in go applications that provide HTTP and/or TCP service.

## Installation

    go get github.com/lostdragon/goreload

## Usage
Send `HUP` to a process using `goreload` and it will restart without downtime.

    kill -HUP ${pid}

Send `QUIT` to a process using `goreload` and it will graceful shutdown.

    kill -QUIT ${pid}


| Signal            | Function              	|
| ------------------|---------------------------|
| TERM, INT    	    | Quick shutdown		    |
| QUIT              | Graceful shutdown     	|
| KILL              | Halts a stubborn process  |
| HUP               | Graceful restart        	|


## Refer

[https://github.com/rcrowley/goagain](https://github.com/rcrowley/goagain)

[http://grisha.org/blog/2014/06/03/graceful-restart-in-golang/](http://grisha.org/blog/2014/06/03/graceful-restart-in-golang/)

[https://github.com/mindreframer/golang-stuff/blob/master/github.com/astaxie/beego/reload.go](https://github.com/mindreframer/golang-stuff/blob/master/github.com/astaxie/beego/reload.go)

