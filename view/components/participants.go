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
	pId          int64
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
		app.P().Text(p.errMessage).Class("err-message"),
		app.Button().Text("Exit").OnClick(func(ctx app.Context, e app.Event) {
			p.exitTeam(ctx, e, p.pId)
		}),
		app.Range(p.participants).Slice(func(i int) app.UI {
			if p.participants[i].UserId == p.user.Id {
				p.pId = p.participants[i].Id
			}
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
							p.deleteParticipant(ctx, e, p.participants[i].Id)
						}),
						app.If(!p.participants[i].Admin, func() app.UI {
							return app.Button().Text("Give Admin").OnClick(func(ctx app.Context, e app.Event) {
								p.giveAdmin(ctx, e, p.participants[i].Id)
							})
						}),
					)
				}),
			)
		}),
	)
}

func (p *ParticipantsList) deleteParticipant(ctx app.Context, e app.Event, pId int64) {
	url := fmt.Sprintf("%v/participant/%v/%v", app.Getenv("BACK_URL"), p.teamId, pId)
	app.Log(url)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		app.Log(err)
		p.errMessage = "Could not delete participant"
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", p.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		p.errMessage = "Could not delete participant"
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		app.Log(err)
		p.errMessage = "Failed to read delete participant response"
		return
	}
	if res.StatusCode == http.StatusOK {
		p.errMessage = "Member deleted"
	} else {
		var errResBody errResponseBody
		if err := json.Unmarshal(b, &errResBody); err != nil {
			app.Log(err)
			p.errMessage = "Failed to parse delete participant response data"
			return
		}
		p.errMessage = errResBody.Message
		app.Log(errResBody.Err)
	}
}

func (p *ParticipantsList) giveAdmin(ctx app.Context, e app.Event, pId int64) {
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%v/participant/%v", app.Getenv("BACK_URL"), pId), nil)
	if err != nil {
		app.Log(err)
		p.errMessage = "Could not execute the request"
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", p.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		p.errMessage = "Could not give admin to the member"
		return
	}
	if res.StatusCode == http.StatusOK {
		p.errMessage = "Admin role has given to the member"
	} else {
		app.Log(err)
		p.errMessage = "Could not give admin to the member"
	}
}

func (p *ParticipantsList) exitTeam(ctx app.Context, e app.Event, pId int64) {
	url := fmt.Sprintf("%v/participant/%v/%v", app.Getenv("BACK_URL"), p.teamId, pId)
	app.Log(url)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		app.Log(err)
		p.errMessage = "Could not exit team"
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", p.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		p.errMessage = "Could not exit team"
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		app.Log(err)
		p.errMessage = "Failed to read exit team response"
		return
	}
	if res.StatusCode == http.StatusOK {
		ctx.Navigate("/")
	} else {
		var errResBody errResponseBody
		if err := json.Unmarshal(b, &errResBody); err != nil {
			app.Log(err)
			p.errMessage = "Failed to parse exit team response data"
			return
		}
		p.errMessage = errResBody.Message
		app.Log(errResBody.Err)
	}
}
