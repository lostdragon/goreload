#!/bin/sh

set -e -x

cd "$(dirname "$0")"


for NAME in "socket"
do
    cd "example/$NAME"
    go build
    ./$NAME &
    PID="$!"
    i=1;
    while [[ i -le 5 ]] ;
    do
        OLDPID="$PID"
        sleep 1
        kill -HUP "$PID"
        sleep 2
        PID="$(pgrep "$NAME")"
        i=$((i+1));
    done
    [ "$(nc "127.0.0.1" "59081")" = "Hello, socket!" ]
    kill -TERM "$PID"
    sleep 2
    [ -z "$(pgrep "$NAME")" ]
    cd "$OLDPWD"
done

for NAME in "httpandsocket"
do
    cd "example/$NAME"
    go build
    ./$NAME &
    PID="$!"
    i=1;
    while [[ i -le 5 ]] ;
    do
        echo "========= ${i} ================"
        OLDPID="$PID"
        sleep 2
        [ "$(nc "127.0.0.1" "59081")" = "Hello, socket!" ]
        [ "$(curl "http://127.0.0.1:58081")" = "Hello, http!" ]
        kill -HUP "$PID"
        sleep 2
        PID="$(pgrep "$NAME")"
        i=$((i+1));
    done;
    [ "$(nc "127.0.0.1" "59081")" = "Hello, socket!" ]
    [ "$(curl "http://127.0.0.1:58081")" = "Hello, http!" ]
    kill -TERM "$PID"
    sleep 3
    [ -z "$(pgrep "$NAME")" ]
    cd "$OLDPWD"
done

for NAME in "http"
do
    cd "example/$NAME"
    go build
    ./$NAME &
    PID="$!"
    i=1;
    while [[ i -le 5 ]] ;
    do
        echo "========= ${i} ================"
        OLDPID="$PID"
        sleep 2
        [ "$(curl "http://127.0.0.1:58081")" = "Hello, http!" ]
        kill -HUP "$PID"
        sleep 2
        PID="$(pgrep "$NAME")"
        i=$((i+1));
    done;
    [ "$(curl "http://127.0.0.1:58081")" = "Hello, http!" ]
    kill -TERM "$PID"
    sleep 3
    [ -z "$(pgrep "$NAME")" ]
    cd "$OLDPWD"
done


for NAME in "multihttp"
do
    cd "example/$NAME"
    go build
    ./$NAME &
    PID="$!"
    i=1;
    while [[ i -le 5 ]] ;
    do
        echo "========= ${i} ================"
        OLDPID="$PID"
        sleep 2
        [ "$(curl "http://127.0.0.1:58081")" = "Hello, http!" ]
        [ "$(curl "http://127.0.0.1:58082")" = "Hello, http!" ]
        kill -HUP "$PID"
        sleep 2
        PID="$(pgrep "$NAME")"
        i=$((i+1));
    done;
    [ "$(curl "http://127.0.0.1:58081")" = "Hello, http!" ]
    [ "$(curl "http://127.0.0.1:58082")" = "Hello, http!" ]
    kill -TERM "$PID"
    sleep 3
    [ -z "$(pgrep "$NAME")" ]
    cd "$OLDPWD"
done