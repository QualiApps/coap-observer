package controllers

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/qualiapps/observer/models"
)

type (
	// ConfController represents the controller for operating on the Config resource
	ConfigController struct {
		ConFile string
	}
)

func NewConfigController(conFile string) *ConfigController {
	return &ConfigController{
		ConFile: conFile,
	}
}

/**
 * Retrieves all clients
 * @return json
 */
func (c *ConfigController) GetClients(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	models.DbFile = c.ConFile
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
	response, _ := models.AddClient(r.Body)

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
	res := models.DeleteClient(p.ByName("id"))
	if res {
		status = 200
	}
	w.WriteHeader(status)
}
