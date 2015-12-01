package db

import (
	"gopkg.in/olivere/elastic.v2"
	"log"
	"net/url"
	"strings"
)

var (
	dIndex = "index" //default index name
	dType  = "type"  // default type name
)

type ES struct {
	Conn  *elastic.Client
	Host  string
	Port  string
	Index string
	Type  string
}

func NewES(host, port string) (*ES, error) {
	es := &ES{Host: host, Port: port}

	url := es.GetUrl(host, port)

	var err error
	es.Conn, err = elastic.NewClient(
		elastic.SetURL(url),
	)
	if err != nil {
		return nil, err
	}

	_, _, err = es.Conn.Ping().Do()
	if err != nil {
		return nil, err
	}

	return es, nil
}

func (e *ES) GetUrl(h, p string) string {
	u := &url.URL{
		Scheme: "http",
		Host:   strings.Join([]string{h, p}, ":"),
	}

	return u.String()
}

func (e *ES) GetType() string {
	if e.isEmpty(e.Type) {
		e.Type = dType
	}
	return e.Type
}

func (e *ES) SetType(t string) {
	e.Type = t
}

func (e *ES) GetIndex() string {
	if e.isEmpty(e.Index) {
		e.Index = dIndex
	}
	return e.Index
}

func (e *ES) SetIndex(i string) {
	e.Index = i
}

func (e *ES) Get(id string) []byte {
	var data = []byte("")
	// Get item with specified ID
	get, err := e.Conn.Get().
		Index(e.GetIndex()).
		Type(e.GetType()).
		Id(id).
		Do()
	if err != nil {
		log.Printf("GET ERROR! %#v", err)
	}
	if get.Found {
		data = []byte(*get.Source)
	}

	return data
}

func (e *ES) Post(data string) bool {
	return e.send(data)
}

func (e *ES) Put(data string) bool {
	return e.send(data)
}

/**
 * @TODO - needs to implement
 */
func (e ES) Remove(id string) []byte {
	return []byte("This operation is not allowed!")
}

func (e *ES) CreateIndex(i string) error {
	return e.CheckIndex(i)
}

func (e *ES) DeleteIndex(i string) bool {
	isDelete := true
	_, err := e.Conn.DeleteIndex(i).Do()
	if err != nil {
		log.Printf("DELETE INDEX ERROR: %#v\n", err)
		isDelete = false
	}
	return isDelete
}

func (e *ES) CheckIndex(i string) error {
	// Use the IndexExists service to check if a specified index exists.
	exist, err := e.Conn.IndexExists(i).Do()
	if err != nil {
		// Set error
		return err
	}
	if !exist {
		// Create a new index.
		createIndex, err := e.Conn.CreateIndex(i).Do()
		if err != nil {
			// set error
			return err
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}
	}
	e.SetIndex(i)

	return nil
}

func (e *ES) send(data string) bool {
	_, err := e.Conn.Index().
		Index(e.GetIndex()).
		Type(e.GetType()).
		BodyString(data).
		Do()
	if err != nil {
		log.Printf("POST ERROR: %#v\n", err)
		return false
	}

	return true
}

func (e *ES) isEmpty(s string) bool {
	empty := false
	if len(strings.TrimSpace(s)) == 0 {
		empty = true
	}

	return empty
}
