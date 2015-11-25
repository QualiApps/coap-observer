package main

import (
	//"github.com/dustin/go-coap"
	"github.com/julienschmidt/httprouter"
	"github.com/qualiapps/observer/controllers"

	"flag"
	"log"
	"net/http"
	"strings"
)

func serv_http() {
	HttpHost := flag.String("host", "localhost", "Http server address")
	HttpPort := flag.String("port", "4000", "Http server port")
	ConFile := flag.String("conf", "clients.db", "Config file")

	flag.Parse()

	// init router
	router := httprouter.New()
	// Init controller
	controller := controllers.NewConfigController(*ConFile)

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

func serv_coap() {
}

func main() {
	go serv_http()
	for {
	}
}
