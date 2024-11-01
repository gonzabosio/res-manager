package view

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gonzabosio/res-manager/model"
	"github.com/gonzabosio/res-manager/view/components"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Dashboard struct {
	app.Compo
	components.ProjectList
	components.ParticipantsList

	team             model.Team
	user             model.User
	showParticipants bool
	showTeamForm     bool
	errMessage       string

	newTeamName string
	newPassword string

	accessToken string
	admin       bool
}

func (d *Dashboard) OnMount(ctx app.Context) {
	ctx.Async(func() {
		defer ctx.Update()
		atCookie := app.Window().Call("getAccessTokenCookie")
		if atCookie.IsUndefined() {
			ctx.Navigate("/")
		} else {
			d.accessToken = atCookie.String()
		}
		if err := ctx.SessionStorage().Get("user", &d.user); err != nil {
			app.Log("Could not get user data from session storage", err)
		}
		if err := ctx.LocalStorage().Get("teamName", &d.team.Name); err != nil {
			app.Log("Could not get team name from local storage", err)
		}
		var teamIDstr string
		if err := ctx.LocalStorage().Get("teamID", &teamIDstr); err != nil {
			app.Log("Could not get the team id from local storage", err)
		}
		teamID, err := strconv.ParseInt(teamIDstr, 10, 64)
		if err != nil {
			app.Log("Error parsing teamID to int64:", err)
			return
		}
		d.team.Id = teamID
		ctx.SessionStorage().Get("admin", &d.admin)
	})
}

func (d *Dashboard) Render() app.UI {
	return app.Div().Body(
		app.H1().Text(fmt.Sprintf("Dashboard of %v", d.team.Name)),
		app.If(d.admin, func() app.UI {
			return app.Div().Body(
				app.Button().Text("Edit").OnClick(d.toggleTeamForm),
				app.Button().Text("Delete").OnClick(d.deleteTeam),
			)
		}),
		app.Button().Text("Change Team").OnClick(d.changeTeam),
		app.If(d.showParticipants, func() app.UI {
			return app.Div().Body(
				app.Button().Text("Projects").OnClick(func(ctx app.Context, e app.Event) {
					d.showParticipants = false
				}),
				&d.ParticipantsList,
			)
		}).ElseIf(d.showTeamForm, func() app.UI {
			return app.Div().Body(
				app.Input().Type("text").Placeholder("Team name").
					Value(d.newTeamName).
					AutoFocus(true).
					OnChange(d.ValueTo(&d.newTeamName)),
				app.Input().Type("password").Placeholder("Password").
					Value(d.newPassword).
					AutoFocus(true).
					OnChange(d.ValueTo(&d.newPassword)),
				app.Button().Text("Accept").OnClick(d.editTeam),
				app.Button().Text("Cancel").OnClick(d.toggleTeamForm),
				app.P().Text(d.errMessage),
			)
		}).Else(func() app.UI {
			return app.Div().Body(
				app.Button().Text("Participants").OnClick(func(ctx app.Context, e app.Event) {
					d.showParticipants = true
				}),
				&d.ProjectList,
			)
		}),
	)
}

type teamResponse struct {
	Message string     `json:"message"`
	Team    model.Team `json:"team"`
}

func (d *Dashboard) editTeam(ctx app.Context, e app.Event) {
	if d.newPassword != "" && len(d.newPassword) < 4 {
		d.errMessage = "Password length must be at least 4 characters"
		return
	}
	reqBody := fmt.Sprintf(`{"id":%v, "name":"%v","password":"%v"}`, d.team.Id, d.newTeamName, d.newPassword)
	reader := strings.NewReader(reqBody)
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%v/team", app.Getenv("BACK_URL")), reader)
	app.Log("Edit team with the new name:", d.newTeamName)
	if err != nil {
		app.Log(err)
		d.errMessage = "Failed modifying team"
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", d.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		d.errMessage = "Failed modifying team"
		return
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		d.errMessage = "Failed reading the new team modifications"
		return
	}
	if res.StatusCode == http.StatusOK {
		var body teamResponse
		if err = json.Unmarshal(b, &body); err != nil {
			app.Log(err)
			d.errMessage = "Could not parse the team modifications"
			return
		}
		app.Log("New team", body.Team)
		d.team = body.Team
		ctx.LocalStorage().Set("teamName", d.team.Name)
		teamIDstr := strconv.FormatInt(d.team.Id, 10)
		ctx.LocalStorage().Set("teamID", teamIDstr)
		d.newPassword = ""
		d.newTeamName = ""
		d.errMessage = ""
		d.showTeamForm = !d.showTeamForm
	} else {
		var body errResponseBody
		if err = json.Unmarshal(b, &body); err != nil {
			app.Log(err)
			d.errMessage = "Could not parse the team modifications"
		}
		app.Log(body.Err)
		d.errMessage = body.Message
		return
	}
	d.newPassword = ""
	d.newTeamName = ""
	d.errMessage = ""
}

func (d *Dashboard) deleteTeam(ctx app.Context, e app.Event) {
	app.Log("Delete:", d.team.Name)
	client := http.Client{}
	req, err := http.NewRequest(
		http.MethodDelete, fmt.Sprintf("%v/team/%v", app.Getenv("BACK_URL"), d.team.Id), nil,
	)
	if err != nil {
		d.errMessage = "Failed creating request to delete team"
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", d.accessToken))
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		d.errMessage = "Failed deleting team"
		return
	}
	if res.StatusCode == http.StatusOK {
		ctx.Navigate("/")
	} else {
		app.Log(err)
		d.errMessage = "Failed deleting team"
	}
}

func (d *Dashboard) toggleTeamForm(ctx app.Context, e app.Event) {
	d.showParticipants = false
	d.showTeamForm = !d.showTeamForm
}

func (d *Dashboard) changeTeam(ctx app.Context, e app.Event) {
	ctx.SessionStorage().Del("teamID")
	ctx.SessionStorage().Del("teamName")
	ctx.Navigate("/")
}
