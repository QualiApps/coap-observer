package db

import (
	"github.com/qualiapps/observer/utils"
	"os"
)

func GetClient() (*ES, error) {
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
		index = DIndex
	}

	es, err := NewES(host, port)
	if err != nil {
		return nil, err
	}

	err = es.CreateIndex(index)
	if err != nil {
		return nil, err
	}

	return es, nil
}
