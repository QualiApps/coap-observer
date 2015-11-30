package main

import (
	"github.com/dustin/go-coap"
	"github.com/julienschmidt/httprouter"
	"github.com/qualiapps/observer/controllers"
	"github.com/qualiapps/observer/corelinks"
	"github.com/qualiapps/observer/models"
	"github.com/qualiapps/observer/utils"

	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func servHttp(confChan chan models.Config, reg chan models.Client, rm chan string) {
	HttpHost := flag.String("host", "localhost", "Http server address")
	HttpPort := flag.String("port", "4000", "Http server port")
	ConFile := flag.String("conf", "clients.db", "Config file")

	flag.Parse()

	// init router
	router := httprouter.New()

	register := func(c models.Client) {
		fmt.Printf("REG")
		reg <- c
	}

	delete := func(id string) {
		fmt.Printf("RM")
		rm <- id
	}

	// Init controller
	controller := controllers.NewConfigController(*ConFile, register, delete)

	clients := models.GetAllClients()
	var conf models.Config
	json.Unmarshal(clients, &conf)
	confChan <- conf
	close(confChan)

	// Get clients list
	router.GET("/clients", controller.GetClients)

	// Add a new client
	router.POST("/clients/add", controller.AddClient)

	// Remove a client
	router.DELETE("/clients/delete/:id", controller.RemoveClient)

	log.Fatal(
		http.ListenAndServe(
			strings.Join([]string{*HttpHost, *HttpPort}, ":"),
			router,
		),
	)

}

func Discover(device models.Client) (interface{}, error) {
	req := coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.GET,
		MessageID: utils.GenMessageID(),
	}

	req.SetOption(coap.MaxAge, 3)
	req.SetPathString("/.well-known/core")

	c, err := coap.Dial("udp", strings.Join([]string{device.Host, device.Port}, ":"))
	if err != nil {
		return nil, err
	}

	rv, err := c.Send(req)
	if err != nil {
		return nil, err
	}

	return string([]byte(rv.Payload)), nil
}

type Registered struct {
	Id         string
	RemoteAddr *net.UDPAddr
	Req        []coap.Message
}

func GetRegClientByKey(key string) *Registered {
	for _, reg := range reg_res {
		if reg.Id == key {
			return &reg
		}
	}

	return nil
}

func Deregister(l *net.UDPConn, addr *net.UDPAddr, req *coap.Message) {
	fmt.Printf("\nDEREGISTER - Resource: %s\n", req.Option(coap.URIPath))

	err := coap.Transmit(l, addr, *req)
	if err != nil {
		log.Fatalf("DEREGISTER ERROR: %#v\n", err)
	}
}

func DeregisterDevices(l *net.UDPConn /*regResources []Registered*/) {
	/*	for _, res := range regResources {
		//conn, found := ClientDict[res.Id]
		//if found {
		for _, r := range res.Req {
			//fmt.Printf("\nDEREGISTER - Host: %s:%s, Res: %s\n", res.Host, res.Port, r.Option(coap.URIPath))
			fmt.Printf("---%#v\n", r)
			//r.SetOption(coap.Observe, 1)
			//_, err := conn.Send(r)
			//if err != nil {
			//	fmt.Println(err)
			//}
			time.Sleep(time.Second)
		}
		//}
	}*/

}

var ClientDict map[string]*coap.Conn = make(map[string]*coap.Conn)
var reg_res []Registered

func ProcessResponse(l *net.UDPConn, response Response) {
	rv := coap.Message{}
	err := rv.UnmarshalBinary(response.Data)

	if err == nil {
		if rv.IsConfirmable() {
			m := coap.Message{
				Type:      coap.Acknowledgement,
				Code:      0,
				MessageID: rv.MessageID,
			}
			err := coap.Transmit(l, response.FromAddr, m)
			if err != nil {
				fmt.Println(err)
			}

		}
		log.Printf("Got %s", rv.Payload)
		fmt.Printf("CON: %#v\n", rv.IsConfirmable())
		fmt.Printf("TOKEN: %#v\n", rv.Token)
		fmt.Printf("MID: %#v\n", rv.MessageID)
		fmt.Printf("------------------------------------------------------------\n")
	}

}

func UpdateDevices(l *net.UDPConn, conf models.Config) {
	return
}

func CheckReg(id string) bool {
	// Checks registered device
	available := false
	for _, reg := range reg_res {
		if reg.Id == id {
			available = true
			break
		}
	}

	return available
}

func Register(l *net.UDPConn, device models.Client) bool {
	registered := false
	fmt.Printf("Send REG - Host: %s, Port: %s\n", device.Host, device.Port)

	conn := []string{device.Host, device.Port}
	keyID := models.GenerateId(conn)

	if CheckReg(keyID) {
		registered = true
	}

	resources, err := Discover(device)
	if err != nil {
		log.Fatalf("Do not find device: %s\n", err)
		return false
	}
	coreLinks := resources.(string)
	if coreLinks == "" {
		registered = false
	}
	endPoints := core.Parse(coreLinks)

	if endPoints != nil {
		req := coap.Message{
			Type: coap.NonConfirmable,
			Code: coap.GET,
		}
		req.AddOption(coap.Observe, 0)

		RegDev := &Registered{}
		remoteAddr, _ := net.ResolveUDPAddr("udp", strings.Join(conn, ":"))

		for _, res := range endPoints {
			req.Token = []byte(utils.GenToken(8))
			req.MessageID = utils.GenMessageID()

			link := res.GetLink()
			if len(link) > 1 {
				req.SetPathString(link)

				err := coap.Transmit(l, remoteAddr, req)
				if err != nil {
					log.Fatalf("Error sending request: %v", err)
				}
				RegDev.Req = append(RegDev.Req, req)
			}
		}
		if len(RegDev.Req) > 0 {
			RegDev.Id = keyID
			RegDev.RemoteAddr = remoteAddr
			reg_res = append(reg_res, *RegDev)
			registered = true
		}
	}

	return registered

}

func RegisterDevices(l *net.UDPConn, conf models.Config) {
	for _, device := range conf {
		Register(l, device)
	}

}

func UDPListener(listener chan *net.UDPConn, handler chan Response, conStr HostPort) {
	sAddr, err := net.ResolveUDPAddr(conStr.Net, conStr.Address)
	if err != nil {
		panic(err)
	}

	l, err := net.ListenUDP(conStr.Net, sAddr)
	if err != nil {
		panic(err)
	}

	listener <- l

	buf := make([]byte, 1500)
	response := Response{}
	for {
		nr, fromAddr, err := l.ReadFromUDP(buf)
		if err != nil {
			if neterr, ok := err.(net.Error); ok && (neterr.Temporary() || neterr.Timeout()) {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			panic(err)
		}
		tmp := make([]byte, nr)
		copy(tmp, buf)

		response.Data = tmp
		response.FromAddr = fromAddr

		handler <- response
	}

}

type (
	Response struct {
		Data     []byte
		FromAddr *net.UDPAddr
	}
	HostPort struct {
		Net     string
		Address string
	}
)

func main() {
	var (
		conf       = make(chan models.Config)
		register   = make(chan models.Client)
		deregister = make(chan string)
		handler    = make(chan Response)
		listener   = make(chan *net.UDPConn)
		exit       = make(chan os.Signal, 1)
	)

	connString := HostPort{"udp", ":10001"}
	signal.Notify(exit, os.Interrupt)
	signal.Notify(exit, syscall.SIGTERM)

	go UDPListener(listener, handler, connString)
	go servHttp(conf, register, deregister)

	l := <-listener
	RegisterDevices(l, <-conf)

	for {
		select {
		case device := <-register:
			Register(l, device)
		case keyID := <-deregister:
			client := GetRegClientByKey(keyID)
			if client != nil {
				for _, req := range client.Req {
					Deregister(l, client.RemoteAddr, &req)
				}
			}
		case response := <-handler:
			go ProcessResponse(l, response)
		case <-exit:
			go func() {
				DeregisterDevices(l)
				os.Exit(0)
			}()

		}
	}
}
