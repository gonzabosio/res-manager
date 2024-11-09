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

type CreateTeam struct {
	app.Compo

	user        model.User
	teamName    string
	password    string
	errMessage  string
	accessToken string
}

type participantResponse struct {
	Message     string            `json:"message"`
	Participant model.Participant `json:"participant"`
}

func (c *CreateTeam) OnMount(ctx app.Context) {
	if err := ctx.SessionStorage().Get("user", &c.user); err != nil {
		app.Log("Could not get user from local storage")
	}
	atCookie := app.Window().Call("getAccessTokenCookie")
	// app.Log("access token", atCookie.String())
	if atCookie.IsUndefined() {
		ctx.Navigate("/")
	} else {
		c.accessToken = atCookie.String()
	}
}

func (c *CreateTeam) Render() app.UI {
	return app.Div().Body(
		app.Text("Create Team"),
		app.Form().Body(
			app.Input().Type("text").Value(c.teamName).MaxLength(30).
				Placeholder("Team name").
				AutoFocus(true).
				OnChange(c.ValueTo(&c.teamName)),
			app.Input().Type("password").Value(c.password).
				Placeholder("Password").
				OnChange(c.ValueTo(&c.password)),
		),
		app.Button().Text("Create").OnClick(c.createAction).Class("global-btn"),
		app.P().Text(c.errMessage).Class("err-message"),
	)
}

func (c *CreateTeam) createAction(ctx app.Context, e app.Event) {
	if c.teamName == "" || c.password == "" {
		c.errMessage = "Empty team name or password field"
	} else {
		ctx.Async(func() {
			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/team", app.Getenv("BACK_URL")), strings.NewReader(fmt.Sprintf(
				`{"name":"%v","password":"%v"}`,
				c.teamName, c.password)),
			)
			if err != nil {
				app.Log(err)
				c.errMessage = "Could not build the team request"
				return
			}
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
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
				log.Println("Failed to read response body:", err)
				return
			}
			if res.StatusCode == http.StatusOK {
				// app.Log("Request successful:", string(b))
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
				ctx.Async(func() {
					// add participant with admin role
					req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/participant", app.Getenv("BACK_URL")), strings.NewReader(fmt.Sprintf(`
					{"admin":%v,"user_id":%v,"team_id":%v}
					`, true, c.user.Id, resBody.TeamID)))
					if err != nil {
						app.Log(err)
						c.errMessage = "Could not build the participant request"
						return
					}
					req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
					req.Header.Add("Content-Type", "application/json")
					client := http.Client{}
					res, err := client.Do(req)
					if err != nil {
						app.Log(err)
						c.errMessage = "Failed to add participant"
						return
					}
					defer res.Body.Close()
					b, err := io.ReadAll(res.Body)
					if err != nil {
						c.errMessage = "Failed reading the partipant response"
						return
					}

					if res.StatusCode == http.StatusOK {
						var body participantResponse
						if err = json.Unmarshal(b, &body); err != nil {
							app.Log(err)
							c.errMessage = "Could not parse the participant response"
							return
						}
						// app.Log("Participant retrieved", body.Participant)
						ctx.SessionStorage().Set("admin", true)
						ctx.Navigate("dashboard")
					} else {
						var body errResponseBody
						if err = json.Unmarshal(b, &body); err != nil {
							app.Log(err)
							c.errMessage = "Could not parse the participant data"
							return
						}
						app.Log(body.Err)
						c.errMessage = body.Message
					}
				})
			} else {
				var resBody errResponseBody
				if err := json.Unmarshal(b, &resBody); err != nil {
					c.errMessage = "Failed to parse json"
					app.Log(resBody.Err)
					return
				}
				app.Log(resBody.Err)
				c.errMessage = resBody.Message
				ctx.Dispatch(func(ctx app.Context) {})
			}
		})
	}
}
