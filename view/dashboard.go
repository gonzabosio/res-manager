package view

import (
	"fmt"
	"strconv"

	"github.com/gonzabosio/res-manager/view/components"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Dashboard struct {
	app.Compo
	components.ProjectList
	components.ParticipantsList

	showParticipants bool
	teamName         string
	teamID           int64
}

func (t *Dashboard) OnMount(ctx app.Context) {
	if err := ctx.LocalStorage().Get("teamName", &t.teamName); err != nil {
		app.Log("Could not get team name from local storage", err)
	}
	var teamIDstr string
	if err := ctx.LocalStorage().Get("teamID", &teamIDstr); err != nil {
		app.Log("Could not get the team id from local storage", err)
		return
	}
	teamID, err := strconv.ParseInt(teamIDstr, 10, 64)
	if err != nil {
		app.Log("Error parsing teamID to int64:", err)
		return
	}
	t.teamID = teamID
}

func (t *Dashboard) Render() app.UI {
	return app.Div().Body(
		app.H1().Text(fmt.Sprintf("Dashboard of %v", t.teamName)),
		app.If(t.showParticipants, func() app.UI {
			return app.Div().Body(
				app.Button().Text("Projects").OnClick(func(ctx app.Context, e app.Event) {
					t.showParticipants = false
				}),
				&t.ParticipantsList,
			)
		}).Else(func() app.UI {
			return app.Div().Body(
				app.Button().Text("Participants").OnClick(func(ctx app.Context, e app.Event) {
					t.showParticipants = true
				}),
				&t.ProjectList,
			)
		}),
	)
}
