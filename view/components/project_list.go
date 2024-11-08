package components

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gonzabosio/res-manager/model"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type ProjectList struct {
	app.Compo
	showAddProjectForm bool

	projectlist       []model.Project
	newProjectName    string
	newProjectDetails string

	teamID     int64
	errMessage string

	accessToken string
}

type projectsResponse struct {
	Message  string          `json:"message"`
	Projects []model.Project `json:"projects"`
}

func (p *ProjectList) OnMount(ctx app.Context) {
	atCookie := app.Window().Call("getAccessTokenCookie")
	if atCookie.IsUndefined() {
		ctx.Navigate("/")
	} else {
		p.accessToken = atCookie.String()
	}
	var teamIDstr string
	if err := ctx.LocalStorage().Get("teamID", &teamIDstr); err != nil {
		app.Log("Could not get the team id from local storage", err)
		return
	}
	teamID, err := strconv.ParseInt(teamIDstr, 10, 64)
	if err != nil {
		app.Log("Error parsing teamID to int64:", err)
		return
	}
	p.teamID = teamID
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/project/%v", app.Getenv("BACK_URL"), p.teamID), nil)
	if err != nil {
		p.errMessage = "Could not load the projects"
		app.Log(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", p.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		p.errMessage = fmt.Sprintf("Failed getting projects: %v", err)
		return
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		app.Log(err)
		p.errMessage = "Failed reading projects"
		return
	}
	if res.StatusCode == http.StatusOK {
		var resBody projectsResponse
		if err = json.Unmarshal(b, &resBody); err != nil {
			app.Log(err)
			p.errMessage = "Failed parsing projects data"
		}
		p.projectlist = resBody.Projects
	}
}

func (p *ProjectList) Render() app.UI {
	return app.Div().Body(
		app.If(p.showAddProjectForm, func() app.UI {
			return app.Div().Body(
				app.Input().Type("text").Placeholder("Project name").
					Value(p.newProjectName).
					AutoFocus(true).
					OnChange(p.ValueTo(&p.newProjectName)),
				app.Input().Type("text").Placeholder("Project details").
					Value(p.newProjectDetails).
					OnChange(p.ValueTo(&p.newProjectDetails)),
				app.Button().Text("Add").OnClick(p.addProject),
				app.Button().Text("Cancel").OnClick(p.switchFormView),
				app.P().Text(p.errMessage).Class("err-message"),
			)
		}).Else(func() app.UI {
			return app.Div().Body(
				app.P().Text("Projects"),
				app.Button().Text("Add Project").OnClick(p.switchFormView),
				app.Range(p.projectlist).Slice(func(i int) app.UI {
					return app.Div().Body(
						app.A().Text(p.projectlist[i].Name).Href("/dashboard/project").OnClick(func(ctx app.Context, e app.Event) {
							app.Log(p.projectlist[i].Id, p.projectlist[i].Name)
							ctx.SessionStorage().Set("project", p.projectlist[i])
						}),
						// Href(fmt.Sprintf("/dashboard/project?id=%d&name=%s", p.projectlist[i].Id, p.projectlist[i].Name))
					)
				}),
				app.P().Text(p.errMessage).Class("err-message"),
			)
		}),
	)
}

func (p *ProjectList) switchFormView(ctx app.Context, e app.Event) {
	p.newProjectName = ""
	p.newProjectDetails = ""
	p.errMessage = ""
	p.showAddProjectForm = !p.showAddProjectForm
}

func (p *ProjectList) addProject(ctx app.Context, e app.Event) {
	if p.newProjectName == "" || p.newProjectDetails == "" {
		p.errMessage = "All fields must be filled"
		return
	}
	app.Log(fmt.Sprintf("Name: %v\nDetails: %v\nTeamID: %v", p.newProjectName, p.newProjectDetails, p.teamID))
	req, err := http.NewRequest(http.MethodPost, app.Getenv("BACK_URL")+"/project", strings.NewReader(fmt.Sprintf(
		`{"name":"%v","details":"%v","team_id":%d}`,
		p.newProjectName, p.newProjectDetails, p.teamID)),
	)
	if err != nil {
		p.errMessage = "Could not add the project"
		app.Log(err)
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", p.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		p.errMessage = "Failed adding new project"
		return
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		p.errMessage = "Failed to read project data"
		app.Log(err)
		return
	}
	if res.StatusCode == http.StatusOK {
		var resBody projectsResponse
		err := json.Unmarshal(b, &resBody)
		if err != nil {
			p.errMessage = "Failed to parse project data"
			app.Log(err)
			return
		}
		p.newProjectName = ""
		p.newProjectDetails = ""
		p.errMessage = ""
		p.projectlist = resBody.Projects
		p.showAddProjectForm = false
		ctx.Reload()
	} else {
		var resBody errResponseBody
		err := json.Unmarshal(b, &resBody)
		if err != nil {
			p.errMessage = "Failed to parse project data"
			app.Log(err)
			return
		}
		p.errMessage = resBody.Message
		app.Log(resBody.Err)
	}
}
