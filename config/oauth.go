package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func GoogleConfig() *oauth2.Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	log.Println("Enviroment variables:", os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"), os.Getenv("REDIRECT_URI"))
	return &oauth2.Config{
		RedirectURL:  os.Getenv("REDIRECT_URI"),
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}
