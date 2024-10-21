package view

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

var BACK_URL string

type Home struct {
	app.Compo
}

func (h *Home) Render() app.UI {
	return app.Div().Body(
		app.H1().Text("Resources Manager"),
		app.A().Text("Create Team").Href("/create-team"),
		app.A().Text("Join Team").Href("/join-team"),
	)
}
