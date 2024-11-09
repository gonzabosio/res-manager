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

	teamsList []model.TeamView

	teamFilter     string
	pageNumber     int
	pageNumberList []int
	totalPages     int
	offset         int
	limit          int
	totalTeams     int
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
	h.pageNumber = 1
	h.limit = 3
	h.loadTeamsList(ctx)
	if err := ctx.SessionStorage().Get("user", &h.user); err != nil {
		app.Log("Could not get user from local storage")
	}
	atCookie := app.Window().Call("getAccessTokenCookie")
	// app.Log("access token", atCookie.String())
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
}

func (h *Home) Render() app.UI {
	return app.Div().Class("container-home").Body(
		app.Div().Class("title-container").Body(
			app.Img().Src("https://www.svgrepo.com/show/449746/files.svg").Width(100),
			app.H1().Text("RESUR").ID("title"),
		),
		app.If(h.accessToken != "", func() app.UI {
			return app.Div().Body(
				app.If(h.showUpdateUserForm, func() app.UI {
					return app.Div().Body(
						app.Input().Placeholder("New username").Value(h.newUsername).OnChange(h.ValueTo(&h.newUsername)),
						app.Button().Text("Save").OnClick(h.modifyUser).Class("global-btn"),
						app.Button().Text("Cancel").Class("global-btn").OnClick(func(ctx app.Context, e app.Event) {
							h.newUsername = ""
							h.showUpdateUserForm = false
						}),
					)
				}).Else(func() app.UI {
					return app.Div().Body(
						app.P().Text("Hello "+h.user.Username),
						app.Button().Text("Create Team").Class("global-btn").OnClick(func(ctx app.Context, e app.Event) {
							ctx.Navigate("/create-team")
						}),
						app.Button().Text("Join Team").Class("global-btn").OnClick(func(ctx app.Context, e app.Event) {
							ctx.Navigate("/join-team")
						}),
						app.Div().Body(
							app.Button().Text("Modify username").Class("global-btn").OnClick(func(ctx app.Context, e app.Event) {
								h.showUpdateUserForm = true
							}),
							app.Button().Text("Delete user").Class("global-btn").OnClick(h.deleteUser),
							app.Button().Text("Sign Out").Class("global-btn").OnClick(func(ctx app.Context, e app.Event) {
								app.Window().Call("deleteAccessTokenCookie")
								ctx.SessionStorage().Del("user")
								ctx.Reload()
							}),
						),
					)
				}),
			)
		}).Else(func() app.UI {
			return app.A().Text("Sign in with Google").Href(fmt.Sprintf("%v/auth/google_login", app.Getenv("BACK_URL"))).
				Class("global-btn").ID("google-login")
		}),
		app.P().Text(h.errMessage).Class("err-message"),
		app.P().Text("Teams List"),
		app.Div().Body(
			app.Input().Type("text").Placeholder("Team name").Value(h.teamFilter).MaxLength(30).OnChange(h.ValueTo(&h.teamFilter)).Class("search-bar"),
			app.Button().Text("Search").Class("global-btn").OnClick(func(ctx app.Context, e app.Event) {
				h.offset = 0
				h.loadTeamsList(ctx)
				h.pageNumber = 1
			}),
		).Class("search-filter"),
		app.Div().Class("card-container").Body(
			app.Range(h.teamsList).Slice(func(i int) app.UI {
				return app.Div().Class("card").Body(
					app.P().Text(h.teamsList[i].Name),
					app.P().Text(fmt.Sprintf("%d participant/s", h.teamsList[i].ParticipantsNumber)),
				)
			}),
		),
		app.Div().Body(
			app.Button().Text("<").Class("pagination-btn").OnClick(func(ctx app.Context, e app.Event) {
				if h.pageNumber > 1 {
					h.pageNumber--
					h.offset -= 3
					h.loadTeamsList(ctx)
				}
			}),
			app.Range(h.pageNumberList).Slice(func(i int) app.UI {
				return app.Button().Class("pagination-btn").Text(h.pageNumberList[i]).Disabled(h.pageNumberList[i] == h.pageNumber).
					OnClick(func(ctx app.Context, e app.Event) {
						h.paginationBtn(ctx, i)
					})
			}),
			app.Button().Text(">").Class("pagination-btn").OnClick(func(ctx app.Context, e app.Event) {
				if h.pageNumber < h.totalPages {
					h.pageNumber++
					h.offset += 3
					h.loadTeamsList(ctx)
				} else {
				}
			}),
		).Class("pagination-bar"),
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
		// app.Log(h.showUpdateUserForm)
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

func (h *Home) loadTeamsList(ctx app.Context) {
	// app.Log(h.limit, h.offset, h.teamFilter)
	var (
		res *http.Response
		err error
	)
	if h.teamFilter == "" {
		res, err = http.Get(fmt.Sprintf("%v/teams?offset=%v&limit=%v", app.Getenv("BACK_URL"), h.offset, h.limit))
	} else {
		res, err = http.Get(fmt.Sprintf("%v/teams?offset=%v&limit=%v&filter=%s", app.Getenv("BACK_URL"), h.offset, h.limit, h.teamFilter))
	}
	if err != nil {
		h.errMessage = "Failed to load teams list"
		return
	}
	type teamListResponse struct {
		Message string           `json:"message"`
		Teams   []model.TeamView `json:"teams"`
		Count   int              `json:"count"`
	}
	if res.StatusCode == http.StatusOK {
		var teamsList teamListResponse
		if err = json.NewDecoder(res.Body).Decode(&teamsList); err != nil {
			h.errMessage = "Failed to load teams list"
			app.Log(err)
			return
		}
		h.teamsList = teamsList.Teams
		h.totalTeams = teamsList.Count
		// app.Log("Total teams", h.totalTeams)
		h.pageNumberList = []int{}
		h.totalPages = h.totalTeams / 3
		if h.totalTeams%3 != 0 {
			h.totalPages++
		}
		for i := 1; i <= h.totalPages; i++ {
			h.pageNumberList = append(h.pageNumberList, i)
		}
		// app.Log("Page number List:", h.pageNumberList)
		ctx.Dispatch(func(ctx app.Context) {})
	} else {
		var errResBody errResponseBody
		if err = json.NewDecoder(res.Body).Decode(&errResBody); err != nil {
			h.errMessage = "Failed to load teams list"
			app.Log(err)
			return
		}
		h.errMessage = errResBody.Message
		app.Log(errResBody.Err)
	}

}

func (h *Home) paginationBtn(ctx app.Context, i int) {
	if h.pageNumber != h.pageNumberList[i] {
		h.pageNumber = h.pageNumberList[i]
		if h.pageNumberList[i] == 1 {
			h.pageNumber = h.pageNumberList[i]
			h.offset = 0
			h.loadTeamsList(ctx)
		} else {
			h.offset = 0
			h.pageNumber = h.pageNumberList[i]
			if h.pageNumber%2 == 0 {
				h.offset = h.pageNumber*2 - 1
			} else {
				h.offset = h.pageNumber * 2
			}
			h.loadTeamsList(ctx)
		}
	}
}
