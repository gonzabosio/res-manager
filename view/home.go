package view

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gonzabosio/res-manager/model"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Home struct {
	app.Compo
	errMessage  string
	user        model.User
	accessToken string
}

type errResponseBody struct {
	Message string `json:"message"`
	Err     string `json:"error"`
}

type okResponseBody struct {
	Message string `json:"message"`
	TeamID  int64  `json:"team_id"`
}

type userResponse struct {
	Message string     `json:"message"`
	User    model.User `json:"user"`
}

func (h *Home) OnMount(ctx app.Context) {
	if err := ctx.SessionStorage().Get("user", &h.user); err != nil {
		app.Log("Could not get user from local storage")
	}
	atCookie := app.Window().Call("getAccessTokenCookie")
	app.Log("access token", atCookie.String())
	if atCookie.IsUndefined() {
		h.accessToken = ""
	} else {
		h.accessToken = atCookie.String()
		var googleUser model.GoogleUser
		ctx.Async(func() {
			resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + h.accessToken)
			if err != nil {
				h.errMessage = ""
				return
			}
			defer resp.Body.Close()

			userData, err := io.ReadAll(resp.Body)
			if err != nil {
				h.errMessage = "Response JSON Parsing Failed"
				return
			}
			app.Log(string(userData))
			if err := json.Unmarshal(userData, &googleUser); err != nil {
				h.errMessage = "User JSON Parsing Failed"
				return
			}
			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("%v/user", app.Getenv("BACK_URL")),
				strings.NewReader(fmt.Sprintf(`{"username":"%v","email":"%v"}`, googleUser.Name, googleUser.Email)),
			)
			if err != nil {
				h.errMessage = "Failed to build user request"
				app.Log(err)
			}
			token := fmt.Sprintf("Bearer %s", h.accessToken)
			app.Log(token)
			req.Header.Add("Authorization", token)
			req.Header.Add("Content-Type", "application/json")
			client := http.Client{}
			res, err := client.Do(req)
			if err != nil {
				app.Log(err)
				h.errMessage = "Failed registering user"
				return
			}
			b, err := io.ReadAll(res.Body)
			if err != nil {
				h.errMessage = "Failed reading the user response"
				return
			}
			if res.StatusCode == http.StatusOK {
				var body userResponse
				if err = json.Unmarshal(b, &body); err != nil {
					app.Log(err)
					h.errMessage = "Could not parse the user response"
					return
				}
				app.Log("User retrieved", body.User)
				h.user = body.User
				ctx.SessionStorage().Set("user", h.user)
			} else {
				var body errResponseBody
				if err = json.Unmarshal(b, &body); err != nil {
					app.Log(err)
					h.errMessage = "Could not parse the user data"
					return
				}
				app.Log(body.Err)
				h.errMessage = body.Message
			}
		})
	}
}

func (h *Home) Render() app.UI {
	return app.Div().Body(
		app.H1().Text("Resources Manager"),
		app.If(h.accessToken != "", func() app.UI {
			return app.Div().Body(
				app.P().Text("Hello "+h.user.Username),
				app.Button().Text("Create Team").OnClick(func(ctx app.Context, e app.Event) {
					ctx.Navigate("/create-team")
				}),
				app.Button().Text("Join Team").OnClick(func(ctx app.Context, e app.Event) {
					ctx.Navigate("/join-team")
				}),
				app.Div().Body(
					app.Button().Text("Modify username"),
					app.Button().Text("Delete user"),
				),
			)
		}).Else(func() app.UI {
			return app.Div().Body(
				app.A().Text("Sign in with Google").Href(fmt.Sprintf("%v/auth/google_login", app.Getenv("BACK_URL"))),
			)
		}),
		app.P().Text(h.errMessage),
	)
}
