package view

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type RefreshHome struct {
	app.Compo
}

func (r *RefreshHome) Render() app.UI {
	return app.Div().Body(
		app.H1().Text("You have successfully logged in"),
		app.P().Text("Please return to the original website and refresh the page to continue"),
	)
}
