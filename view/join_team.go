package view

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gonzabosio/res-manager/model"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type JoinTeam struct {
	app.Compo

	user        model.User
	teamName    string
	password    string
	errMessage  string
	accessToken string
}

func (c *JoinTeam) OnMount(ctx app.Context) {
	if err := ctx.SessionStorage().Get("user", &c.user); err != nil {
		app.Log("Could not get user from local storage")
	}
	atCookie := app.Window().Call("getAccessTokenCookie")
	app.Log("access token", atCookie.String())
	if atCookie.IsUndefined() {
		ctx.Navigate("/")
	} else {
		c.accessToken = atCookie.String()
	}
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

func (j *JoinTeam) joinAction(ctx app.Context, e app.Event) {
	if j.teamName == "" || j.password == "" {
		j.errMessage = "Empty team name or password field"
	} else {
		ctx.Async(func() {
			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/join-team", app.Getenv("BACK_URL")), strings.NewReader(fmt.Sprintf(
				`{"name":"%v","password":"%v"}`,
				j.teamName, j.password)))
			if err != nil {
				app.Log(err)
				j.errMessage = "Could not build the team request"
				return
			}
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", j.accessToken))
			req.Header.Add("Content-Type", "application/json")
			client := http.Client{}
			res, err := client.Do(req)
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
				ctx.Async(func() {
					req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/participant", app.Getenv("BACK_URL")), strings.NewReader(fmt.Sprintf(`
					{"admin":%v,"user_id":%v,"team_id":%v}
					`, false, j.user.Id, resBody.TeamID,
					)))
					if err != nil {
						app.Log(err)
						j.errMessage = "Could not build the participant request"
						return
					}
					req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", j.accessToken))
					req.Header.Add("Content-Type", "application/json")
					client := http.Client{}
					res, err := client.Do(req)
					if err != nil {
						app.Log(err)
						j.errMessage = "Failed to add participant"
						return
					}
					defer res.Body.Close()
					b, err := io.ReadAll(res.Body)
					if err != nil {
						app.Log(err)
						j.errMessage = "Failed reading the partipant response"
						return
					}

					if res.StatusCode == http.StatusOK {
						var body participantResponse
						if err = json.Unmarshal(b, &body); err != nil {
							app.Log(err)
							j.errMessage = "Could not parse the participant response"
							return
						}
						app.Log("Participant retrieved", body.Participant)
						ctx.SessionStorage().Set("admin", body.Participant.Admin)
						ctx.Navigate("dashboard")
					} else {
						var body errResponseBody
						if err = json.Unmarshal(b, &body); err != nil {
							app.Log(err)
							j.errMessage = "Could not parse the participant data"
							return
						}
						app.Log(body.Message, body.Err)
						j.errMessage = body.Message
					}
				})
			} else {
				app.Log("Request failed with status:", res.StatusCode)
				var resBody errResponseBody
				if err := json.Unmarshal(b, &resBody); err != nil {
					j.errMessage = "Failed to parse json"
					return
				}
				app.Log(resBody.Err)
				j.errMessage = resBody.Message
				ctx.Dispatch(func(ctx app.Context) {})
			}
		})
	}
}
