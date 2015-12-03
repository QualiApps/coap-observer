package main

import (
	"flag"
	"fmt"
	"github.com/qualiapps/observer/db"
	client "github.com/qualiapps/observer/models"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var (
	HttpPort     *string
	ListenerPort *string
	ConFile      *string
)

func init() {
	HttpPort = flag.String("port", "4000", "Conf server port")
	ListenerPort = flag.String("lport", "56083", "Lstener port")
	ConFile = flag.String("conf", "clients.db", "Config file")

	flag.Parse()
}

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

	connString := HostPort{Net, ":" + *ListenerPort}
	signal.Notify(exit,
		os.Interrupt,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGINT,
	)

	go ServHttp(conf, register, deregister)
	go ServCoap(listener, handler, connString)

	// Create observe instance
	observe := NewObserve(<-listener)

	fmt.Printf("Observe Service was started...\n")
	fmt.Printf("Config serv :%s\n", *HttpPort)
	fmt.Printf("Listener serv :%s\n", *ListenerPort)

	// Bootstrap
	observe.RegisterDevices(<-conf)

	for {
		select {
		// register new device
		case device := <-register:
			observe.Register(device)
		// remove device
		case keyID := <-deregister:
			id, client := GetRegClientByKey(keyID)
			if client != nil {
				for _, req := range client.Req {
					observe.Deregister(client.RemoteAddr, &req)
				}
				if id >= 0 {
					RegRes = append(RegRes[:id], RegRes[id+1:]...)
				}
			}
		// on response
		case response := <-handler:
			go ProcessResponse(observe.Conn, es, response)
		// terminate app
		case <-exit:
			go func() {
				observe.DeregisterDevices(RegRes)
				os.Exit(0)
			}()

		}
	}
}
