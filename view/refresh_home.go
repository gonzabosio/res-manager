package view

import (
	"log"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type RefreshHome struct {
	app.Compo
}

func (r *RefreshHome) OnMount(ctx app.Context) {
	r.GetAccessToken(ctx)
}

func (r *RefreshHome) Render() app.UI {
	return app.Div().Body(
		app.H1().Text("You have successfully logged in"),
		app.P().Text("Please return to the original website and refresh the page to continue"),
	)
}

func (r *RefreshHome) GetAccessToken(ctx app.Context) {
	accessToken := ctx.Page().URL().Query().Get("access_token")
	log.Println("ACCESS TOKEN", accessToken)
	ctx.LocalStorage().Set("access-token", accessToken)
}
