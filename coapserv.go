package main

import (
	"github.com/dustin/go-coap"
	client "github.com/qualiapps/observer/models"
	"github.com/qualiapps/observer/utils"
	"log"
	"net"
	"strings"
	"time"
)

func ServCoap(listener chan *net.UDPConn, handler chan Response, conStr HostPort) {
	sAddr, err := net.ResolveUDPAddr(conStr.Net, conStr.Address)
	if err != nil {
		panic(err)
	}

	l, err := net.ListenUDP(conStr.Net, sAddr)
	if err != nil {
		panic(err)
	}

	listener <- l

	buf := make([]byte, MaxBufSize)
	response := Response{}
	for {
		nr, fromAddr, err := l.ReadFromUDP(buf)
		if err != nil {
			if neterr, ok := err.(net.Error); ok && (neterr.Temporary() || neterr.Timeout()) {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			log.Printf("Can't read from UDP: %#v\n", err)
		}
		tmp := make([]byte, nr)
		copy(tmp, buf)

		response.Data = tmp
		response.FromAddr = fromAddr

		handler <- response
	}

}

func Discover(device client.Client) (interface{}, error) {
	req := coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.GET,
		MessageID: utils.GenMessageID(),
	}

	req.SetOption(coap.MaxAge, 3)
	req.SetPathString(WellKnown)

	conn, err := coap.Dial(Net, strings.Join([]string{device.Host, device.Port}, ":"))
	if err != nil {
		return nil, err
	}

	rv, err := conn.Send(req)
	if err != nil {
		return nil, err
	}

	return string([]byte(rv.Payload)), nil
}
