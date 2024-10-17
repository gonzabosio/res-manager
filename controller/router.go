package controller

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gonzabosio/res-manager/controller/handlers"
)

func Routing() *chi.Mux {
	r := chi.NewRouter()
	h, err := handlers.NewHandler()
	if err != nil {
		log.Fatal(err)
	}
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{fmt.Sprintf("http://localhost%v", os.Getenv("FRONT_PORT"))},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	}))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Resources Manager"))
	})
	r.Post("/team", h.CreateTeam)
	r.Get("/team", h.GetTeams)
	r.Get("/team/{team-id}", h.GetTeamByID)
	r.Patch("/team", h.ModifyTeam)
	r.Delete("/team/{team-id}", h.DeleteTeam)

	r.Post("/project", h.CreateProject)
	r.Get("/project/{team-id}", h.GetProjectByTeamID)
	r.Patch("/project", h.ModifyProject)
	r.Delete("/project/{project-id}", h.DeleteProjectByID)
	return r
}
