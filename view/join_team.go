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

type JoinTeam struct {
	app.Compo

	teamName   string
	password   string
	errMessage string
	backURL    string
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
			app.Input().Type("password").Value(j.password).
				Placeholder("Password").
				OnChange(j.ValueTo(&j.password)),
		),
		app.Button().Text("Join").OnClick(j.joinAction),
		app.P().Text(j.errMessage),
	)
}

type errResponseBody struct {
	Message string `json:"message"`
	Err     string `json:"error"`
}

type okResponseBody struct {
	Message string `json:"message"`
	TeamID  int64  `json:"team_id"`
}

func (j *JoinTeam) joinAction(ctx app.Context, e app.Event) {
	if j.teamName == "" || j.password == "" {
		j.errMessage = "Empty team name or password field"
	} else {
		res, err := http.Post(fmt.Sprintf("%v/join-team", j.backURL), "application/json",
			strings.NewReader(fmt.Sprintf(
				`{"name":"%v","password":"%v"}`,
				j.teamName, j.password)))
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
				j.errMessage = "Failed to parse json"
				return
			}
			if err := ctx.LocalStorage().Set("teamName", j.teamName); err != nil {
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
				j.errMessage = "Failed to parse json"
				return
			}
			log.Println("messsage: ", resBody.Message)
			app.Log(resBody.Err)
			j.errMessage = resBody.Message
		}
	}
}
