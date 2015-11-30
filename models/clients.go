package models

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"

	"io"
	"io/ioutil"
	"os"
	"strings"
)

type (
	Config map[string]Client

	Client struct {
		Host    string   `json:"host"`
		Port    string   `json:"port"`
		EPoints []string `json:"ep"` // client endpoints (resources)
	}
)

/**
 * Path to conf file
 */
type DbConf struct {
	Name string `default:"clients.db"`
}

var Db = DbConf{"clients.db"}

/**
 * Retrieves all clients
 * @return - returns bytes array
 */
func GetAllClients() []byte {
	file, _ := ioutil.ReadFile(Db.Name)
	return []byte(file)
}

func GetClientById(id string) []byte {
	return []byte("")
}

/**
 * Adds a new client
 * @param io.Reader params - json
 * @return ([]byte, bool)
 */
func AddClient(params io.Reader) ([]byte, bool) {
	var (
		err    error
		client = Client{}
		cl     []byte
	)

	// Populate params
	err = json.NewDecoder(params).Decode(&client)
	if err != nil {
		log.Printf("Json Decode error: %#v\n", err)
		return nil, false
	}

	// Add an Id
	id := GenerateId([]string{client.Host, client.Port})

	clients, added := checkClient(id)
	if added {
		return nil, false
	}
	// The first item (client)
	if clients == nil {
		clients = make(map[string]Client)
	}
	clients[id] = client

	cl, err = json.Marshal(&clients)
	if err != nil {
		log.Printf("Json Marshal error: %#v\n", err)
		return nil, false
	}

	var ac []byte
	if writeConf(cl) {
		ac, err = json.Marshal(client)
		if err != nil {
			log.Printf("Json Marshal error: %#v\n", err)
			return nil, false
		}

	}

	return ac, true
}

/**
 * Removes the client by Id
 * @param string id - md5 hash of host:port
 * @return bool
 */
func DeleteClient(id string) bool {
	clients, ok := checkClient(id)
	if ok {
		delete(clients, id)
		cl, err := json.Marshal(&clients)
		if err != nil {
			log.Printf("Json Marshal error:", err)
			return false
		}
		writeConf(cl)
		return true
	}
	return false
}

/**
 * Writes config to file
 * @param []byte data - incoming data
 * @return bool
 */
func writeConf(data []byte) bool {
	file, err := os.OpenFile(Db.Name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("File open error:", err)
		return false
	}
	defer file.Close()

	if _, err = file.Write(data); err != nil {
		log.Printf("Write to file error:", err)
		return false
	}
	return true

}

/**
 * Checks an item
 * @param string id - client id
 * @return Confug, bool
 */
func checkClient(id string) (Config, bool) {
	exist := false
	var conf Config
	clients := GetAllClients()

	json.Unmarshal(clients, &conf)
	if _, ok := conf[id]; ok {
		exist = true
	}

	return conf, exist
}

/**
 * Generates a hash
 * @param []string data - [host, port]
 * @return string - md5 hash
 */
func GenerateId(data []string) string {
	hash := md5.New()
	hash.Write([]byte(strings.Join(data, ":")))
	return hex.EncodeToString(hash.Sum(nil))
}
