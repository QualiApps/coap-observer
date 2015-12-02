package main

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/qualiapps/observer/controllers"
	client "github.com/qualiapps/observer/models"
	"log"
	"net/http"
	"strings"
)

func ServHttp(confChan chan client.Config, reg chan client.Client, rm chan string) {
	// init router
	router := httprouter.New()

	register := func(c client.Client) {
		reg <- c
	}

	delete := func(id string) {
		rm <- id
	}

	// Init controller
	controller := controllers.NewConfigController(*ConFile, register, delete)

	// Init bootstrap
	clients := client.GetAllClients()
	var conf client.Config
	json.Unmarshal(clients, &conf)
	confChan <- conf
	close(confChan)

	// Get clients list
	router.GET("/clients", controller.GetClients)

	// Add a new client
	router.POST("/clients/add", controller.AddClient)

	// Remove a client
	router.DELETE("/clients/delete/:id", controller.RemoveClient)

	log.Fatal(
		http.ListenAndServe(
			strings.Join([]string{*HttpHost, *HttpPort}, ":"),
			router,
		),
	)

}
