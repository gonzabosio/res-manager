package view

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type CreateTeam struct {
	app.Compo

	teamName   string
	password   string
	errMessage string
	backURL    string
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
			app.Input().Type("password").Value(c.password).
				Placeholder("Password").
				OnChange(c.ValueTo(&c.password)),
		),
		app.Button().Text("Create").OnClick(c.createAction),
		app.P().Text(c.errMessage),
	)
}

func (c *CreateTeam) createAction(ctx app.Context, e app.Event) {
	if c.teamName == "" || c.password == "" {
		c.errMessage = "Empty team name or password field"
	} else {
		res, err := http.Post(fmt.Sprintf("%v/team", c.backURL), "application/json",
			strings.NewReader(fmt.Sprintf(
				`{"name":"%v","password":"%v"}`,
				c.teamName, c.password)))
		if err != nil {
			log.Println(err)
			return
		}
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		if err != nil {
			log.Println("Failed to send request:", err)
			return
		}
		if res.StatusCode == http.StatusOK {
			app.Log("Request successful:", string(b))
			var resBody okResponseBody
			if err := json.Unmarshal(b, &resBody); err != nil {
				app.Log(err)
				c.errMessage = "Failed to parse json"
				return
			}
			if err := ctx.LocalStorage().Set("teamName", c.teamName); err != nil {
				app.Log(err)
			}
			teamIDstr := strconv.FormatInt(resBody.TeamID, 10)
			if err := ctx.LocalStorage().Set("teamID", teamIDstr); err != nil {
				app.Log(err)
			}
			ctx.Navigate("dashboard")
		} else {
			app.Log("Request failed with status:", res.StatusCode)
			var resBody errResponseBody
			if err := json.Unmarshal(b, &resBody); err != nil {
				c.errMessage = "Failed to parse json"
				return
			}
			log.Println("messsage: ", resBody.Message)
			app.Log(resBody.Err)
			c.errMessage = resBody.Message
		}
	}
}
