package main

import (
	"github.com/qualiapps/observer/db"
	client "github.com/qualiapps/observer/models"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		conf       = make(chan client.Config)
		register   = make(chan client.Client)
		deregister = make(chan string)
		handler    = make(chan Response)
		listener   = make(chan *net.UDPConn)
		exit       = make(chan os.Signal, 1)
	)
	// Init DB
	es, err := db.GetClient()
	if err != nil {
		log.Fatalf("Can't init DB: %#v\n", err)
	}

	connString := HostPort{Net, ":"}
	signal.Notify(exit, os.Interrupt)
	signal.Notify(exit, syscall.SIGTERM)

	go ServHttp(conf, register, deregister)
	go ServCoap(listener, handler, connString)

	l := <-listener
	RegisterDevices(l, <-conf)

	for {
		select {
		case device := <-register:
			Register(l, device)
		case keyID := <-deregister:
			id, client := GetRegClientByKey(keyID)
			if client != nil {
				for _, req := range client.Req {
					Deregister(l, client.RemoteAddr, &req)
				}
				RegRes = append(RegRes[:id], RegRes[id+1:]...)
			}
		case response := <-handler:
			go ProcessResponse(l, es, response)
		case <-exit:
			go func() {
				DeregisterDevices(l, RegRes)
				os.Exit(0)
			}()

		}
	}
}
