package controller

import (
	"log"
	"net/http"
	"os"
)

func InitBackend() error {
	r := Routing()
	log.Printf("Running backend server on %v\n", os.Getenv("BACK_URL"))
	return http.ListenAndServe(os.Getenv("BACK_PORT"), r)
}
