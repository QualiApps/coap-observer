package main

import (
	"github.com/dustin/go-coap"
	"github.com/qualiapps/observer/corelinks"
	client "github.com/qualiapps/observer/models"
	"github.com/qualiapps/observer/utils"

	"fmt"
	"log"
	"net"
	"strings"
)

const (
	WellKnown  = "/.well-known/core"
	Net        = "udp"
	MaxBufSize = 1500
)

type (
	HostPort struct {
		Net     string
		Address string
	}
	Registered struct {
		Id         string
		RemoteAddr *net.UDPAddr
		Req        []coap.Message
	}
)

var (
	RegRes     []Registered
	ValidToken = make(map[string]bool)
)

func IsValidToken(token []byte) bool {
	valid := false
	if _, ok := ValidToken[string(token)]; ok {
		valid = true
	}
	return valid
}

func RemoveToken(token []byte) {
	delete(ValidToken, string(token))
}

func GetRegClientByKey(key string) (int, *Registered) {
	for id, reg := range RegRes {
		if reg.Id == key {
			return id, &reg
		}
	}

	return 0, nil
}

func Deregister(l *net.UDPConn, addr *net.UDPAddr, req *coap.Message) {
	fmt.Printf("\nDEREGISTER - Resource: %s, Host: %s, Port: %d", req.Option(coap.URIPath), addr.IP, addr.Port)

	req.SetOption(coap.Observe, 1)
	err := coap.Transmit(l, addr, *req)
	if err != nil {
		log.Fatalf("DEREGISTER ERROR: %#v\n", err)
	}
	RemoveToken(req.Token)
}

func DeregisterDevices(l *net.UDPConn, regResources []Registered) {
	for _, res := range regResources {
		for _, r := range res.Req {
			Deregister(l, res.RemoteAddr, &r)
		}
	}

}

func Register(l *net.UDPConn, device client.Client) bool {
	registered := false

	conn := []string{device.Host, device.Port}
	keyID := client.GenerateId(conn)

	if _, ok := GetRegClientByKey(keyID); ok != nil {
		return false
	}

	resources, err := Discover(device)
	if err != nil {
		log.Printf("Do not find device: %s\n", err)
		return false
	}
	coreLinks := resources.(string)
	if coreLinks == "" {
		registered = false
	}
	if endPoints, ok := core.Parse(coreLinks); ok {
		req := coap.Message{
			Type: coap.NonConfirmable,
			Code: coap.GET,
		}
		req.AddOption(coap.Observe, 0)

		RegDev := &Registered{}
		remoteAddr, _ := net.ResolveUDPAddr(Net, strings.Join(conn, ":"))

		for _, res := range endPoints {
			req.Token = []byte(utils.GenToken(8))
			req.MessageID = utils.GenMessageID()

			link := res.GetLink()
			if len(link) > 1 {
				req.SetPathString(link)

				err := coap.Transmit(l, remoteAddr, req)
				if err != nil {
					log.Printf("Error sending request: %v", err)
				}
				RegDev.Req = append(RegDev.Req, req)
				ValidToken[string(req.Token)] = true
			}
		}
		if len(RegDev.Req) > 0 {
			RegDev.Id = keyID
			RegDev.RemoteAddr = remoteAddr
			RegRes = append(RegRes, *RegDev)
			registered = true
		}
	}

	return registered

}

func RegisterDevices(l *net.UDPConn, conf client.Config) {
	for _, device := range conf {
		Register(l, device)
	}

}

func SendAck(l *net.UDPConn, from *net.UDPAddr, mid uint16) error {
	m := coap.Message{
		Type:      coap.Acknowledgement,
		Code:      0,
		MessageID: mid,
	}
	return coap.Transmit(l, from, m)
}
