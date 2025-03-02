package main

import (
    "os"
    "github.com/Sinubio/mycalcservice/internal/server"
)

func main() {
    addr := ":8080"
    if len(os.Args) > 1 {
        addr = os.Args[1]
    }
    server.NewServer(addr).Run()
}