package components

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gonzabosio/res-manager/model"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type ParticipantsList struct {
	app.Compo

	participants []model.ParticipantsResp
	errMessage   string
	accessToken  string
	teamId       int64
	user         model.User
	admin        bool
}

type participantsResponse struct {
	Message      string                   `json:"message"`
	Participants []model.ParticipantsResp `json:"participants"`
}

func (p *ParticipantsList) OnMount(ctx app.Context) {
	if err := ctx.SessionStorage().Get("user", &p.user); err != nil {
		app.Log("Could not get the user data from session storage")
	}
	if err := ctx.SessionStorage().Get("admin", &p.admin); err != nil {
		app.Log("Could not get the admin data from session storage")
	}
	app.Log(p.admin)
	atCookie := app.Window().Call("getAccessTokenCookie")
	app.Log("access token", atCookie.String())
	if atCookie.IsUndefined() {
		ctx.Navigate("dashboard")
	} else {
		p.accessToken = atCookie.String()
	}
	var teamIDstr string
	if err := ctx.LocalStorage().Get("teamID", &teamIDstr); err != nil {
		p.errMessage = "Could not get team identity"
		app.Log(err)
	}
	teamID, err := strconv.ParseInt(teamIDstr, 10, 64)
	if err != nil {
		app.Log("Error parsing teamID to int64:", err)
		return
	}
	p.teamId = teamID
	//load participants list
	getURL := fmt.Sprintf("%v/participant/%v", app.Getenv("BACK_URL"), p.teamId)
	app.Log(getURL)
	req, err := http.NewRequest(http.MethodGet, getURL, nil)
	if err != nil {
		p.errMessage = "Could not build participant request"
		app.Log(err)
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", p.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		p.errMessage = "Failed to get participants"
		app.Log(err)
		return
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		p.errMessage = "Could not marshal participant data"
		app.Log(err)
		return
	}
	var pResp participantsResponse
	err = json.Unmarshal(b, &pResp)
	if err != nil {
		p.errMessage = "Could not unmarshal participant data"
		app.Log(err)
		return
	}
	if res.StatusCode == http.StatusOK {
		app.Log(pResp.Message)
		app.Log(pResp.Participants)
		p.participants = pResp.Participants
	} else {

	}
}

func (p *ParticipantsList) Render() app.UI {
	return app.Div().Body(
		app.P().Text(p.errMessage),
		app.Range(p.participants).Slice(func(i int) app.UI {
			return app.Div().Body(
				app.If(!p.participants[i].Admin, func() app.UI {
					return app.P().Text(p.participants[i].Username)
				}).Else(func() app.UI {
					return app.Div().Body(
						app.P().Text(fmt.Sprintf("%v - Admin", p.participants[i].Username)),
					)
				}),
				app.If(p.participants[i].UserId != p.user.Id && p.admin, func() app.UI {
					return app.Div().Body(
						app.Button().Text("Delete").OnClick(func(ctx app.Context, e app.Event) {
							p.deleteParticipant(ctx, e, p.participants[i].UserId, p.teamId)
						}),
						app.If(!p.participants[i].Admin, func() app.UI {
							return app.Button().Text("Give Admin").OnClick(func(ctx app.Context, e app.Event) {
								p.giveAdmin(ctx, e, p.participants[i].UserId, p.teamId)
							})
						}),
					)
				}),
			)
		}),
	)
}

func (p *ParticipantsList) deleteParticipant(ctx app.Context, e app.Event, userId, teamId int64) {
	//delete participant by user id and team id
}

func (p *ParticipantsList) giveAdmin(ctx app.Context, e app.Event, userId, teamId int64) {
	//give admin to the participant selected
}
