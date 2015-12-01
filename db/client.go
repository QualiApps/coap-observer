package db

import (
	"github.com/qualiapps/observer/db/es"
	"github.com/qualiapps/observer/utils"

	"log"
	"os"
)

type (
	DBClient interface {
		Get(id string) []byte
		Post(data, id string) (string, bool)
		Put(data, id string) (string, bool)
		Delete(id string) bool
		Processing(data string)
	}
)

func GetClient() DBClient {
	var cl DBClient
	inst, err := InitES()

	if err != nil {
		log.Fatalf("Can't init DB: %#v\n", err)
	}
	cl = inst
	return cl
}

func InitES() (*es.ES, error) {
	host := os.Getenv("ES_HOST")
	if utils.IsEmpty(host) {
		host = "localhost"
	}

	port := os.Getenv("ES_PORT")
	if utils.IsEmpty(port) {
		port = "9200"
	}

	index := os.Getenv("ES_INDEX")
	if utils.IsEmpty(index) {
		index = es.DIndex
	}

	es, err := es.NewES(host, port)
	if err != nil {
		return nil, err
	}

	err = es.CreateIndex(index)
	if err != nil {
		return nil, err
	}

	return es, nil
}
