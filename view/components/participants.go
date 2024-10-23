package components

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type ParticipantsList struct {
	app.Compo
}

func (p *ParticipantsList) OnMount(ctx app.Context) {
	//load participants list
}

func (p *ParticipantsList) Render() app.UI {
	// list := []string{"par1", "par2", "par3"}
	return app.Div().Body(
	// app.Range(list).Slice(func(i int) app.UI {
	// 	return app.Div().Body(
	// 		app.P().Text(list[i]),
	// 	)
	// }),
	// consider user role to show actions
	)
}
