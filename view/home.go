package view

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type Home struct {
	app.Compo
}

func (h *Home) Render() app.UI {
	return app.Div().Body(
		app.H1().Text("Resources Manager"),
	)
}
