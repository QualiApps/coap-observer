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
	Observe struct {
		Conn *net.UDPConn
	}
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
	RegRes      []Registered
	ValidTokens = make(map[string]bool)
)

func NewObserve(l *net.UDPConn) *Observe {
	return &Observe{l}
}

func (o *Observe) RegisterDevices(conf client.Config) {
	for _, device := range conf {
		o.Register(device)
	}
}

func (o *Observe) Register(device client.Client) bool {
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
		return false
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
			link := res.GetLink()
			if len(link) > 1 {
				req.Token = []byte(utils.GenToken(8))
				req.MessageID = utils.GenMessageID()
				req.SetPathString(link)

				err := coap.Transmit(o.Conn, remoteAddr, req)
				if err != nil {
					log.Printf("Error sending request: %v", err)
				}

				RegDev.Req = append(RegDev.Req, req)
				ValidTokens[string(req.Token)] = true
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

func (o *Observe) Deregister(addr *net.UDPAddr, req *coap.Message) {
	log.Printf("DEREGISTER - Resource: %s, Host: %s, Port: %d\n", req.Option(coap.URIPath), addr.IP, addr.Port)

	req.SetOption(coap.Observe, 1)
	err := coap.Transmit(o.Conn, addr, *req)
	if err != nil {
		log.Fatalf("DEREGISTER ERROR: %#v\n", err)
	}
	RemoveToken(req.Token)
}

func (o *Observe) DeregisterDevices(regResources []Registered) {
	for _, res := range regResources {
		for _, r := range res.Req {
			o.Deregister(res.RemoteAddr, &r)
		}
	}
}

func RemoveUnObservable(m *coap.Message, from *net.UDPAddr) {
	RemoveToken(m.Token)

	conn := []string{fmt.Sprint(from.IP), fmt.Sprint(from.Port)}
	keyID := client.GenerateId(conn)
	id, client := GetRegClientByKey(keyID)
	if id >= 0 {
		for i, reg := range client.Req {
			if string(reg.Token) == string(m.Token) {
				client.Req = append(client.Req[:i], client.Req[i+1:]...)
			}
		}
	}
}

func IsValidToken(token []byte) bool {
	valid := false
	if _, ok := ValidTokens[string(token)]; ok {
		valid = true
	}
	return valid
}

func RemoveToken(token []byte) {
	delete(ValidTokens, string(token))
}

func GetRegClientByKey(key string) (int, *Registered) {
	for id, reg := range RegRes {
		if reg.Id == key {
			return id, &RegRes[id]
		}
	}

	return -1, nil
}
