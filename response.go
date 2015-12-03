package main

import (
	"github.com/dustin/go-coap"
	"github.com/qualiapps/observer/db"
	"github.com/qualiapps/observer/utils"

	"log"
	"net"
)

type Response struct {
	Data     []byte
	FromAddr *net.UDPAddr
}

/**
 * Processing response
 * @param *net.UDPConn l - connection instance
 * @param dbClient - db instance
 * @param Response response - res data
 */
func ProcessResponse(l *net.UDPConn, dbClient db.DBClient, response Response) {
	// parse to CoAP struct
	rv := coap.Message{}
	err := rv.UnmarshalBinary(response.Data)
	if err == nil {
		if IsValidToken(rv.Token) {
			if rv.IsObservable() {
				// Send ACK
				if rv.IsConfirmable() {
					err := SendAck(l, response.FromAddr, rv.MessageID)
					if err != nil {
						log.Printf("Send ACK ERROR: %#v\n", err)
					}
				}
				// processing payload
				data := string(rv.Payload)
				if !utils.IsEmpty(data) {
					// Send to DB
					if dbClient != nil {
						dbClient.Processing(data)
					}
				}
			} else {
				go RemoveUnObservable(&rv, response.FromAddr)
			}
		}
	}
}
