package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gonzabosio/res-manager/controller"
	"github.com/gonzabosio/res-manager/view"
	"github.com/joho/godotenv"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading enviroment variables: %s", err)
	}

	app.Route("/", func() app.Composer { return &view.Home{} })
	app.Route("/create-team", func() app.Composer { return &view.CreateTeam{} })
	app.Route("/join-team", func() app.Composer { return &view.JoinTeam{} })
	app.Route("/dashboard", func() app.Composer { return &view.Dashboard{} })
	app.Route("/dashboard/project", func() app.Composer { return &view.Project{} })
	app.Route("/dashboard/project/res", func() app.Composer { return &view.Resource{} })
	app.Route("/refresh", func() app.Composer { return &view.RefreshHome{} })
	app.RunWhenOnBrowser()

	http.Handle("/", &app.Handler{
		Name:        "Resur",
		Description: "Projects requirements and resources manager",
		Title:       "Resur",
		Icon:        app.Icon{SVG: "https://www.svgrepo.com/show/449746/files.svg"},
		Styles:      []string{"/web/style/global.css", "/web/style/home.css", "/web/style/resource.css"},
		Env: map[string]string{
			"BACK_URL": os.Getenv("BACK_URL"),
		},
		Scripts: []string{
			"/web/script/main.js",
		},
	})
	go func() {
		if err := controller.InitBackend(); err != nil {
			log.Fatalf("backend server connection failed: %v", err)
		}
	}()

	if err := http.ListenAndServe(os.Getenv("FRONT_PORT"), nil); err != nil {
		log.Fatalf("wasm server down: %v", err)
	}
}
