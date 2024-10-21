package view

import (
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Dashboard struct {
	app.Compo
	teamName string
}

func (t *Dashboard) OnMount(ctx app.Context) {
	if err := ctx.LocalStorage().Get("teamName", &t.teamName); err != nil {
		app.Log("Could not get team name from local storage", err)
	}
}

func (t *Dashboard) Render() app.UI {
	return app.Div().Body(
		app.H1().Text(fmt.Sprintf("Dashboard of %v", t.teamName)),
	)
}
