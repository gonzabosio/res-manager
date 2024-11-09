package view

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gonzabosio/res-manager/model"
	"github.com/gonzabosio/res-manager/view/components"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Project struct {
	app.Compo
	components.Sections
	showAddSectionForm bool
	modProjectForm     bool

	project             model.Project
	errMessage          string
	projectNameField    string
	projectDetailsField string

	sectionTitleField string

	accessToken string
}

func (p *Project) OnMount(ctx app.Context) {
	atCookie := app.Window().Call("getAccessTokenCookie")
	if atCookie.IsUndefined() {
		ctx.Navigate("/")
	} else {
		p.accessToken = atCookie.String()
	}
	if err := ctx.SessionStorage().Get("project", &p.project); err != nil {
		app.Log("Could not get the project from session storage", err)
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
	p.project.TeamId = teamID
}

func (p *Project) Render() app.UI {
	return app.Div().Body(
		app.H1().Text(fmt.Sprintf("Project %v", p.project.Name)),
		app.P().Text(p.project.Details),
		app.If(p.showAddSectionForm, func() app.UI {
			return app.Div().Body(
				app.Input().Type("text").Placeholder("Title").Value(p.sectionTitleField).OnChange(p.ValueTo(&p.sectionTitleField)),
				app.Button().Text("Add").OnClick(p.addNewSection).Class("global-btn"),
				app.Button().Text("Cancel").OnClick(p.toggleAddSectionForm).Class("global-btn"),
				app.P().Text(p.errMessage).Class("err-message"),
			)
		}).ElseIf(p.modProjectForm, func() app.UI {
			return app.Div().Body(
				app.Input().Type("text").Placeholder("Name").Value(p.projectNameField).OnChange(p.ValueTo(&p.projectNameField)),
				app.Input().Type("text").Placeholder("Details").Value(p.projectDetailsField).OnChange(p.ValueTo(&p.projectDetailsField)),
				app.Button().Text("Accept").OnClick(p.modifyProject).Class("global-btn"),
				app.Button().Text("Cancel").OnClick(p.toggleProjectForm).Class("global-btn"),
				app.P().Text(p.errMessage).Class("err-message"),
			)
		}).Else(func() app.UI {
			return app.Div().Body(
				app.Button().Text("Modify project").OnClick(p.toggleProjectForm).Class("global-btn"),
				app.Button().Text("Delete project").OnClick(p.deleteProject).Class("global-btn"),
				app.P().Text(p.errMessage).Class("err-message"),
				app.P().Text("Sections"),
				app.Button().Text("Add section").OnClick(p.toggleAddSectionForm).Class("global-btn"),
				&p.Sections,
			)
		}),
	)
}

type projectResponse struct {
	Message string        `json:"message"`
	Project model.Project `json:"project"`
}

func (p *Project) modifyProject(ctx app.Context, e app.Event) {
	if p.projectNameField == "" && p.projectDetailsField == "" {
		p.errMessage = "At least one field should be edited"
		return
	}
	reqBody := fmt.Sprintf(`{"id":%v,"name":"%v","details":"%v","team_id":%v}`, p.project.Id, p.projectNameField, p.projectDetailsField, p.project.TeamId)
	reader := strings.NewReader(reqBody)
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%v/project", app.Getenv("BACK_URL")), reader)
	if err != nil {
		app.Log(err)
		p.errMessage = "Failed modifying project"
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", p.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		p.errMessage = "Failed modifying project"
		return
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		p.errMessage = "Failed reading the new project modifications"
		return
	}
	if res.StatusCode == http.StatusOK {
		var body projectResponse
		if err = json.Unmarshal(b, &body); err != nil {
			app.Log(err)
			p.errMessage = "Could not parse the project modifications"
			return
		}
		p.project = body.Project
		url := url.URL{
			Path: "/dashboard/project",
			// RawQuery: fmt.Sprintf("id=%d&name=%s", p.project.Id, p.project.Name),
		}
		ctx.Page().ReplaceURL(&url)
		ctx.SessionStorage().Set("project", p.project)
		p.errMessage = ""
		p.projectNameField = ""
		p.projectDetailsField = ""
		p.modProjectForm = false
	} else {
		var body errResponseBody
		if err = json.Unmarshal(b, &body); err != nil {
			app.Log(body.Err)
			p.errMessage = body.Message
			return
		}
		app.Log(body.Err)
		p.errMessage = body.Message
	}
}

func (p *Project) deleteProject(ctx app.Context, e app.Event) {
	app.Log("Delete:", p.project.Name)
	client := http.Client{}
	req, err := http.NewRequest(
		http.MethodDelete, fmt.Sprintf("%v/project/%v", app.Getenv("BACK_URL"), p.project.Id), nil,
	)
	if err != nil {
		p.errMessage = "Failed to create request to delete project"
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", p.accessToken))
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		p.errMessage = "Failed to delete project"
		return
	}
	if res.StatusCode == http.StatusOK {
		ctx.Navigate("dashboard")
	} else {
		app.Log(err)
		p.errMessage = "Failed to delete project"
	}
}

func (p *Project) toggleProjectForm(ctx app.Context, e app.Event) {
	p.projectNameField = ""
	p.projectDetailsField = ""
	p.errMessage = ""
	p.modProjectForm = !p.modProjectForm
}

func (p *Project) toggleAddSectionForm(ctx app.Context, e app.Event) {
	p.sectionTitleField = ""
	p.errMessage = ""
	p.showAddSectionForm = !p.showAddSectionForm
}

func (p *Project) addNewSection(ctx app.Context, e app.Event) {
	if p.sectionTitleField == "" {
		p.errMessage = "The field is empty"
		return
	}
	app.Log(fmt.Sprintf("Title: %v\nProjectID: %v", p.sectionTitleField, p.project.Id))
	req, err := http.NewRequest(http.MethodPost, app.Getenv("BACK_URL")+"/section", strings.NewReader(fmt.Sprintf(
		`{"title":"%v","project_id":%d}`,
		p.sectionTitleField, p.project.Id)),
	)
	if err != nil {
		app.Log(err)
		p.errMessage = "Could not add new section"
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", p.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		p.errMessage = "Failed adding new section"
		return
	}
	if res.StatusCode == http.StatusOK {
		p.sectionTitleField = ""
		p.errMessage = ""
		p.showAddSectionForm = false
	} else {
		p.errMessage = "Could not add the section"
	}
}
