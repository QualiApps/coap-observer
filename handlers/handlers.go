package handlers

import (
	"fmt"
	"github.com/dustin/go-coap"
	"gopkg.in/olivere/elastic.v2"
	"os"
	"strings"
)

type Handlers interface {
	Get() []byte
	Post() []byte
	Put() []byte
	Delete() []byte
}

type ES struct {
	Connection *elastic.Client
	Host       string
	Port       string
	Index      string
	Type       string
}

type (
	call    func(*elastic.Client, *coap.Message) []byte
	es_conn struct{ host, port string }
)

var (
	actions = map[string]call{
		"get":    get,
		"post":   post,
		"put":    put,
		"delete": remove,
	}
	iot_index      = "storage"
	iot_index_type = "item"
)

func InitHostPort() es_conn {
	host := os.Getenv("ES_HOST")
	if is_empty(host) {
		host = "localhost"
	}
	port := os.Getenv("ES_PORT")
	if is_empty(port) {
		port = "9200"
	}

	return es_conn{host, port}
}

func is_empty(str string) bool {
	empty := false
	if len(strings.TrimSpace(str)) == 0 {
		empty = true
	}

	return empty
}

func get_index() string {
	return iot_index
}

func set_index(ind string) {
	iot_index = ind
}

func get_type() string {
	return iot_index_type
}

func (es *ES) Get(r *coap.Message) []byte {
	// Get item with specified ID
	get, err := client.Get().
		Index(get_index()).
		Type(get_type()).
		Id("AVE0qtUTGQFtvo5rFxAt").
		Do()
	if err != nil {
		fmt.Println("GET ERROR! %#v", err)
		// Handle error
		return []byte("error")
	}
	var data []byte
	if get.Found {
		data = []byte(*get.Source)
	}

	return data
}

func (es *ES) Post(r *coap.Message) []byte {
	return []byte("POST")
}

func (es *ES) Put(r *coap.Message) []byte {
	es_put(es.Connection, string(r.Payload))
	return []byte("PUT OK")
}

func (es ES) Remove(r *coap.Message) []byte {
	return []byte("DELETE")
}

func Call(method string, request *coap.Message) ([]byte, error) {
	client, err := init_es()
	if err != nil {
		return nil, fmt.Errorf("Init ES ERROR: %#v", err)
	}

	switch method {
	case "GET":
		Handlers.Get()
	case "POST":
		Handlers.Post()
	case "PUT":
		Handlers.Put()
	case "DELETE":
		Handlers.Delete()
	default:
		return nil, fmt.Errorf("Request method %s does not support!", method)

	}

	return actions[method](client, request), nil
}

func es_check_index(client *elastic.Client) error {
	// Use the IndexExists service to check if a specified index exists.
	index := get_index()
	exists, err := client.IndexExists(index).Do()
	if err != nil {
		// Handle error
		return err
	}
	if !exists {
		// Create a new index.
		createIndex, err := client.CreateIndex(index).Do()
		if err != nil {
			// Handle error
			return err
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}
	}

	return nil
}

func es_put(client *elastic.Client, body string) error {
	put, err := client.Index().
		Index(get_index()).
		Type(get_type()).
		//Id("2").
		BodyString(body).
		Do()
	if err != nil {
		// Handle error
		fmt.Println("ERROR! %#v", err)
		return err
	}
	fmt.Printf("Indexed data %s to index %s, type %s\n", put.Id, put.Index, put.Type)

	return nil
}

func init_es() (*elastic.Client, error) {
	// Create a client
	//conn := InitHostPort()
	var (
		client *elastic.Client
		err    error
	)
	client, err = elastic.NewClient(
		elastic.SetURL("http://localhost:9200"),
	)
	if err != nil {
		return nil, err
	}

	_, _, err = client.Ping().Do()
	if err != nil {
		return nil, err
	}

	err = es_check_index(client)
	if err != nil {
		return nil, err
	}

	return client, nil
}
