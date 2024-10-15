package view

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type CreateTeam struct {
	app.Compo

	name     string
	password string
}

func (t *CreateTeam) Render() app.UI {
	return app.Div().Body(
		app.Text("Create Team"),
		app.P().Body(
			app.Input().Type("text").Value(t.name).
				Placeholder("Team name").
				AutoFocus(true).
				OnChange(t.ValueTo(&t.name)),
			app.Input().Type("text").Value(t.password).
				Placeholder("Password").
				AutoFocus(true).
				OnChange(t.ValueTo(&t.password)),
		),
		app.Button().Text("Create").OnClick(func(ctx app.Context, e app.Event) {
			app.Log("Send request")
			res, err := http.Post("http://localhost:3060/team", "application/json",
				strings.NewReader(fmt.Sprintf(
					`{
				"name": "%v" 
				"password": "%v"
				}`,
					t.name, t.password)))
			if err != nil {
				log.Println(err)
				return
			}
			b, err := io.ReadAll(res.Body)
			if err != nil {
				log.Println(err)
			}
			log.Println(string(b))
		}),
	)
}
