package view

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Login struct {
	app.Compo

	email    string
	username string
	password string
}

func (l *Login) Render() app.UI {
	return app.Div().Body(
		app.Input().Type("text").Placeholder("Email").Value(l.email).OnChange(l.ValueTo(&l.email)),
		app.Input().Type("text").Placeholder("Username").Value(l.username).OnChange(l.ValueTo(&l.username)),
		app.Input().Type("password").Placeholder("Password").Value(l.password).OnChange(l.ValueTo(&l.password)),
		app.Button().Text("Login").OnClick(l.loginUser),
		app.Button().Text("Cancel").OnClick(func(ctx app.Context, e app.Event) {
			ctx.Navigate("/")
		}),
	)
}

func (l *Login) loginUser(ctx app.Context, e app.Event) {
	app.Log(l.email, l.username, l.password)
}
