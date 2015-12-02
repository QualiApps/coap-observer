package es

import (
	"encoding/json"
	"log"
)

const dbTypeKey = "type"

var mmap map[string]interface{}

func (e *ES) Processing(payload string) {
	json.Unmarshal([]byte(payload), &mmap)
	// Send data if decode ok (json format)
	if mmap != nil {
		// set default type
		e.SetType(DType)
		// init a particular type
		if val, ok := mmap[dbTypeKey]; ok {
			switch eventType := val.(type) {
			case string:
				e.SetType(eventType)
			}

		}
		// post to db
		if _, ok := e.Post(payload, ""); !ok {
			log.Printf("Can't add item %s to DB!\n", payload)
		}

	}

}
