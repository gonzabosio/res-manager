package controller

import (
	"log"
	"net/http"
	"os"
)

func InitBackend() error {
	r := Routing()
	log.Println("Running backend server...")
	return http.ListenAndServe(os.Getenv("BACK_PORT"), r)
}
