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

	teamName     string
	password     string
	errorMessage string
	backURL      string
}

func (c *CreateTeam) OnMount(ctx app.Context) {
	c.backURL = app.Getenv("BACK_URL")
}

func (c *CreateTeam) Render() app.UI {
	return app.Div().Body(
		app.Text("Create Team"),
		app.Form().Body(
			app.Input().Type("text").Value(c.teamName).
				Placeholder("Team name").
				AutoFocus(true).
				OnChange(c.ValueTo(&c.teamName)),
			app.Input().Type("text").Value(c.password).
				Placeholder("Password").
				AutoFocus(true).
				OnChange(c.ValueTo(&c.password)),
		),
		app.Button().Text("Create").OnClick(c.CreateAction),
		app.P().Text(c.errorMessage),
	)
}

func (c *CreateTeam) CreateAction(ctx app.Context, e app.Event) {
	if c.teamName == "" || c.password == "" {
		c.errorMessage = "Empty team name or password field"
	} else {
		app.Log("Sending request...")
		go func() {
			res, err := http.Post(fmt.Sprintf("%v/team", c.backURL), "application/json",
				strings.NewReader(fmt.Sprintf(
					`{"name":"%v","password":"%v"}`,
					c.teamName, c.password)))
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
				ctx.LocalStorage().Set("teamName", c.teamName)
			} else {
				app.Log("Request failed with status:", res.StatusCode)
			}
		}()
	}
}
