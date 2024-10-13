package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gonzabosio/res-manager/view"
	"github.com/joho/godotenv"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading enviroment variables: %s", err)
	}

	app.Route("/", func() app.Composer { return &view.Home{} })
	app.RunWhenOnBrowser()

	http.Handle("/", &app.Handler{
		Name:        "Resources Manager",
		Description: "Projects requirements and resources manager",
		Title:       "Resources Manager",
		Icon:        app.Icon{SVG: "https://svgshare.com/i/1BQ6.svg"},
		Styles:      []string{"/web/style/global.css"},
	})

	if err := http.ListenAndServe(os.Getenv("FRONT_PORT"), nil); err != nil {
		log.Fatalf("Server down: %s", err)
	}
}
