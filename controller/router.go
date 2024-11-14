package controller

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/gonzabosio/res-manager/controller/handlers"
	middlewares "github.com/gonzabosio/res-manager/controller/middleware"
)

func Routing() *chi.Mux {
	r := chi.NewRouter()
	h, err := handlers.NewHandler()
	if err != nil {
		log.Fatal(err)
	}
	r.Use(middleware.Logger)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{os.Getenv("FRONT_URL")},
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
	}))
	r.Head("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Resur"))
	})
	r.Get("/teams", h.GetTeams)

	r.Route("/auth", func(r chi.Router) {
		r.Get("/google_login", h.GoogleLoginHandler)
		r.Get("/google_callback", h.GoogleCallbackHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(middlewares.OAuthMiddleware)
		r.Post("/team", h.CreateTeam)
		r.Post("/join-team", h.VerifyTeamByName)
		r.Patch("/team", h.ModifyTeam)
		r.Delete("/team/{team-id}", h.DeleteTeam)

		r.Post("/project", h.CreateProject)
		r.Get("/project/{team-id}", h.GetProjectsByTeamID)
		r.Patch("/project", h.ModifyProject)
		r.Delete("/project/{project-id}", h.DeleteProject)

		r.Post("/section", h.CreateSection)
		r.Get("/section/{project-id}", h.GetSectionsByProjectID)
		r.Put("/section", h.ModifySection)
		r.Delete("/section/{section-id}", h.DeleteSection)

		r.Post("/resource", h.CreateResource)
		r.Get("/resource/{section-id}", h.GetResourcesBySectionID)
		r.Patch("/resource", h.ModifyResource)
		r.Delete("/resource/{resource-id}", h.DeleteResource)

		r.Post("/user", h.RegisterUser) // if user already exists retrieve info
		r.Get("/user", h.GetUsers)
		r.Patch("/user", h.ModifyUser)
		r.Delete("/user/{user-id}", h.DeleteUser)

		r.Post("/participant", h.AddParticipant)
		r.Get("/participant/{team-id}", h.GetParticipants)
		r.Patch("/participant/{participant-id}", h.GiveAdmin)
		r.Delete("/participant/{team-id}/{participant-id}", h.DeleteParticipant)

		r.Post("/csv", h.UploadCSV)

		r.Post("/image", h.UploadImage)
		r.Get("/image/{resource-id}", h.GetImages)
		r.Delete("/image", h.DeleteImage)
	})
	return r
}
