package view

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type JoinTeam struct {
	app.Compo

	teamName     string
	password     string
	errorMessage string
	backURL      string
}

func (c *JoinTeam) OnMount(ctx app.Context) {
	c.backURL = app.Getenv("BACK_URL")
}

func (j *JoinTeam) Render() app.UI {
	return app.Div().Body(
		app.Text("Join Team"),
		app.Form().Body(
			app.Input().Type("text").Value(j.teamName).
				Placeholder("Team name").
				AutoFocus(true).
				OnChange(j.ValueTo(&j.teamName)),
			app.Input().Type("text").Value(j.password).
				Placeholder("Password").
				AutoFocus(true).
				OnChange(j.ValueTo(&j.password)),
		),
		app.Button().Text("Join").OnClick(j.CreateAction),
		app.P().Text(j.errorMessage),
	)
}

func (j *JoinTeam) CreateAction(ctx app.Context, e app.Event) {
	if j.teamName == "" || j.password == "" {
		j.errorMessage = "Empty team name or password field"
	} else {
		app.Log("Sending request...")
		go func() {
			res, err := http.Post(fmt.Sprintf("%v/join-team", j.backURL), "application/json",
				strings.NewReader(fmt.Sprintf(
					`{"name":"%v","password":"%v"}`,
					j.teamName, j.password)))
			if err != nil {
				log.Println(err)
				return
			}
			b, err := io.ReadAll(res.Body)
			if err != nil {
				log.Println("Failed to send request:", err)
				log.Println(err)
			}
			log.Println(string(b))
			if res.StatusCode == http.StatusOK {
				app.Log("Request successful")
				ctx.Navigate("dashboard")
				ctx.LocalStorage().Set("teamName", j.teamName)
			} else {
				app.Log("Request failed with status:", res.StatusCode)
			}
		}()
	}
}
