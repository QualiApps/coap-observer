package controllers

import (
	"fmt"
	"net/http"

	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/qualiapps/observer/models"
)

type (
	// ConfController represents the controller for operating on the Config resource
	ConfigController struct {
		OnRegister func(c models.Client)
		OnDelete   func(id string)
	}
)

func NewConfigController(conFile string, reg func(c models.Client), rm func(id string)) *ConfigController {
	models.Db.Name = conFile
	return &ConfigController{reg, rm}
}

/**
 * Retrieves all clients
 * @return json
 */
func (c *ConfigController) GetClients(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	clients := models.GetAllClients()

	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	fmt.Fprintf(w, "%s", clients)
}

/**
 * Adds a new client
 */
func (c *ConfigController) AddClient(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	response, ok := models.AddClient(r.Body)
	if ok {
		var client models.Client
		json.Unmarshal(response, &client)
		c.Register(client)
	}

	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", response)
}

/**
 * Removes an existing client
 */
func (c *ConfigController) RemoveClient(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	status := 400
	id := p.ByName("id")
	res := models.DeleteClient(id)
	if res {
		status = 200
		c.Delete(id)
	}
	w.WriteHeader(status)
}

func (c *ConfigController) Register(res models.Client) {
	c.OnRegister(res)
}

func (c *ConfigController) Delete(id string) {
	c.OnDelete(id)
}
