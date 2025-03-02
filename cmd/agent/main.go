package main

import (
    "flag"
    "github.com/Sinubio/mycalcservice/internal/agent"
)

func main() {
    serverURL := flag.String("server", "http://localhost:8080", "Server URL")
    flag.Parse()
    
    agent := agent.NewAgent(*serverURL)
    agent.Start()

    select {} 
}