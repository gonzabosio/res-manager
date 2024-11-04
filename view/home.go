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
	errMessage         string
	user               model.User
	accessToken        string
	newUsername        string
	showUpdateUserForm bool
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
	ctx.Async(func() {
		defer ctx.Update()
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
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", h.accessToken))
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
		}
	})
}

func (h *Home) Render() app.UI {
	return app.Div().Body(
		app.H1().Text("Resources Manager"),
		app.If(h.accessToken != "", func() app.UI {
			return app.Div().Body(
				app.If(h.showUpdateUserForm, func() app.UI {
					return app.Div().Body(
						app.Input().Placeholder("New username").Value(h.newUsername).OnChange(h.ValueTo(&h.newUsername)),
						app.Button().Text("Save").OnClick(h.modifyUser),
						app.Button().Text("Cancel").OnClick(func(ctx app.Context, e app.Event) {
							h.newUsername = ""
							h.showUpdateUserForm = false
						}),
					)
				}).Else(func() app.UI {
					return app.Div().Body(
						app.P().Text("Hello "+h.user.Username),
						app.Button().Text("Create Team").OnClick(func(ctx app.Context, e app.Event) {
							ctx.Navigate("/create-team")
						}),
						app.Button().Text("Join Team").OnClick(func(ctx app.Context, e app.Event) {
							ctx.Navigate("/join-team")
						}),
						app.Div().Body(
							app.Button().Text("Modify username").OnClick(func(ctx app.Context, e app.Event) {
								h.showUpdateUserForm = true
							}),
							app.Button().Text("Delete user").OnClick(h.deleteUser),
							app.Button().Text("Sign Out").OnClick(func(ctx app.Context, e app.Event) {
								app.Window().Call("deleteAccessTokenCookie")
								ctx.SessionStorage().Del("user")
								ctx.Reload()
							}),
						),
					)
				}),
			)
		}).Else(func() app.UI {
			return app.Div().Body(
				app.A().Text("Sign in with Google").Href(fmt.Sprintf("%v/auth/google_login", app.Getenv("BACK_URL"))),
			)
		}),
		app.P().Text(h.errMessage),
	)
}

func (h *Home) deleteUser(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%v/user/%v", app.Getenv("BACK_URL"), h.user.Id), nil)
		if err != nil {
			app.Log(err)
			h.errMessage = "Could not build the request to delete user"
			return
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", h.accessToken))
		req.Header.Add("Content-Type", "application/json")
		client := http.Client{}
		res, err := client.Do(req)
		if err != nil {
			app.Log(err)
			h.errMessage = "Failed to execute the request to delete user"
			return
		}
		b, err := io.ReadAll(res.Body)
		if err != nil {
			app.Log(err)
			h.errMessage = "Failed to read user delete response"
			return
		}
		if res.StatusCode == http.StatusOK {
			app.Log("User deleted successfully")
			app.Window().Call("deleteAccessTokenCookie")
			h.accessToken = ""
			ctx.Dispatch(func(ctx app.Context) {})
		} else {
			var respBody errResponseBody
			err := json.Unmarshal(b, &respBody)
			if err != nil {
				h.errMessage = "Failed to parse error response body"
			}
			h.errMessage = respBody.Message
			app.Log(respBody.Err)
		}
	})
}

func (h *Home) modifyUser(ctx app.Context, e app.Event) {
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%v/user", app.Getenv("BACK_URL")),
		strings.NewReader(
			fmt.Sprintf(`{"id":%v,"username":"%v"}`, h.user.Id, h.newUsername),
		),
	)
	if err != nil {
		app.Log(err)
		h.errMessage = "Failed to build user update request"
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", h.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		h.errMessage = "Failed modifying user"
		return
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		h.errMessage = "Failed reading the new user modifications"
		return
	}
	if res.StatusCode == http.StatusOK {
		var body userResponse
		if err = json.Unmarshal(b, &body); err != nil {
			app.Log(err)
			h.errMessage = "Could not parse the user modifications"
			return
		}
		h.user = body.User
		ctx.SessionStorage().Set("user", h.user)
		h.showUpdateUserForm = false
		app.Log(h.showUpdateUserForm)
	} else {
		var body errResponseBody
		if err = json.Unmarshal(b, &body); err != nil {
			app.Log(err)
			h.errMessage = "Could not parse the user modifications"
			return
		}
		app.Log(body.Err)
		h.errMessage = body.Message
	}
}
