package main

import (
	"encoding/json"
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

func ProcessResponse(l *net.UDPConn, es *db.ES, response Response) {
	rv := coap.Message{}
	err := rv.UnmarshalBinary(response.Data)

	if err == nil {
		if rv.IsObservable() && IsValidToken(rv.Token) {
			// Send ACK
			if rv.IsConfirmable() {
				m := coap.Message{
					Type:      coap.Acknowledgement,
					Code:      0,
					MessageID: rv.MessageID,
				}
				err := coap.Transmit(l, response.FromAddr, m)
				if err != nil {
					log.Printf("Send ACK ERROR: %#v\n", err)
				}

			}
			data := string(rv.Payload)

			if !utils.IsEmpty(data) {
				// Send to DB
				if es != nil {

					var mmap map[string]interface{}
					json.Unmarshal([]byte(rv.Payload), &mmap)
					// Send data if decode ok (json format)
					if mmap != nil {
						es.SetType(db.DType)

						// init a particular type
						if val, ok := mmap["type"]; ok {
							switch eventType := val.(type) {
							case string:
								es.SetType(eventType)
							}

						}

						if _, ok := es.Post(data); !ok {
							log.Printf("Can't add item %d to DB!\n", rv.MessageID)
						}
					}
				}

			}
		}
	}

}
