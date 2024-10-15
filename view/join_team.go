package view

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type JoinTeam struct {
	app.Compo

	name     string
	password string
}

func (j *JoinTeam) Render() app.UI {
	return app.Div().Body(
		app.Text("Join Team"),
		app.P().Body(
			app.Input().Type("text").Value(j.name).
				Placeholder("Team name").
				AutoFocus(true).
				OnChange(j.ValueTo(&j.name)),
			app.Input().Type("text").Value(j.password).
				Placeholder("Password").
				AutoFocus(true).
				OnChange(j.ValueTo(&j.password)),
		),
		app.Button().Text("Join").OnClick(func(ctx app.Context, e app.Event) {
			app.Log("Send request")
		}),
	)
}
