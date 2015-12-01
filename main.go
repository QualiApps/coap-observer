package main

import (
	"github.com/qualiapps/observer/db"
	client "github.com/qualiapps/observer/models"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		conf       = make(chan client.Config) // bootstrap channel
		register   = make(chan client.Client) // register event
		deregister = make(chan string)        // delete event
		handler    = make(chan Response)      // response event
		listener   = make(chan *net.UDPConn)  // entry point, to listen notifications
		exit       = make(chan os.Signal, 1)  // terminate
	)
	// Init DB
	es := db.GetClient()

	connString := HostPort{Net, ":"}
	signal.Notify(exit, os.Interrupt)
	signal.Notify(exit, syscall.SIGTERM)

	go ServHttp(conf, register, deregister)
	go ServCoap(listener, handler, connString)

	l := <-listener
	// Bootstrap
	RegisterDevices(l, <-conf)

	for {
		select {
		// register new device
		case device := <-register:
			Register(l, device)
		// remove device
		case keyID := <-deregister:
			id, client := GetRegClientByKey(keyID)
			if client != nil {
				for _, req := range client.Req {
					Deregister(l, client.RemoteAddr, &req)
				}
				RegRes = append(RegRes[:id], RegRes[id+1:]...)
			}
		// on response
		case response := <-handler:
			go ProcessResponse(l, es, response)
		// terminate app
		case <-exit:
			go func() {
				DeregisterDevices(l, RegRes)
				os.Exit(0)
			}()

		}
	}
}
