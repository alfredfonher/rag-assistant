package main

import (
	"errors"
	"log"
	"net/http"

	"rag-assistant/service/internal/bootstrap"
)

func main() {
	app, err := bootstrap.New()
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:    app.Config.HTTPAddr,
		Handler: app.Server.Handler(),
	}

	log.Printf("starting %s on %s", app.Config.ServiceName, app.Config.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
