package view

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type SignUp struct {
	app.Compo

	email    string
	username string
	password string
}

func (s *SignUp) Render() app.UI {
	return app.Div().Body(
		app.Input().Type("text").Placeholder("Email").Value(s.email).OnChange(s.ValueTo(&s.email)),
		app.Input().Type("text").Placeholder("Username").Value(s.username).OnChange(s.ValueTo(&s.username)),
		app.Input().Type("password").Placeholder("Password").Value(s.password).OnChange(s.ValueTo(&s.password)),
		app.Button().Text("Sign Up").OnClick(s.signUpUser),
		app.Button().Text("Cancel").OnClick(func(ctx app.Context, e app.Event) {
			ctx.Navigate("/")
		}),
	)
}

func (s *SignUp) signUpUser(ctx app.Context, e app.Event) {
	app.Log(s.email, s.username, s.password)
}
