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
var DbFile string

/**
 * Retrieves all clients
 * @return - returns bytes array
 */
func GetAllClients() []byte {
	file, _ := ioutil.ReadFile(DbFile)
	return []byte(file)
}

/**
 * Adds a new client
 * @TODO - return item which has been created
 * @param io.Reader params - json
 * @return ([]byte, bool)
 */
func AddClient(params io.Reader) ([]byte, bool) {
	client := Client{}

	// Populate params
	json.NewDecoder(params).Decode(&client)

	// Add an Id
	id := generateId([]string{client.Host, client.Port})

	clients, added := checkClient(id)
	if added {
		return nil, false
	}
	// The first item (client)
	if clients == nil {
		clients = make(map[string]Client)
	}
	clients[id] = client

	cl, err := json.Marshal(&clients)
	if err != nil {
		log.Printf("Json Marshal error:", err)
		return nil, false
	}
	writeConf(cl)

	return cl, true
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
	}
	return true
}

/**
 * Writes config to file
 * @param []byte data - incoming data
 * @return bool
 */
func writeConf(data []byte) bool {
	file, err := os.OpenFile(DbFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
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
func generateId(data []string) string {
	hash := md5.New()
	hash.Write([]byte(strings.Join(data, ":")))
	return hex.EncodeToString(hash.Sum(nil))
}
