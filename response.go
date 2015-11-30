package main

import (
	"fmt"
	"github.com/dustin/go-coap"
	"net"
)

type Response struct {
	Data     []byte
	FromAddr *net.UDPAddr
}

func ProcessResponse(l *net.UDPConn, response Response) {
	rv := coap.Message{}
	err := rv.UnmarshalBinary(response.Data)

	if err == nil {
		// Send ACK
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
		// Send to DB
		fmt.Printf("------TO DB------------------------------------------------------\n")
	}

}
