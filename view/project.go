package view

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

type Project struct {
	app.Compo
	addSectionForm bool
	modProjectForm bool

	project             model.Project
	errMessage          string
	projectNameField    string
	projectDetailsField string

	sectionTitleField string
}

func (p *Project) OnMount(ctx app.Context) {
	id := ctx.Page().URL().Query().Get("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		app.Log(err)
	}
	p.project.Id = idInt
	p.project.Name = ctx.Page().URL().Query().Get("name")
	app.Log(fmt.Sprintf("In %v - %v", p.project.Id, p.project.Name))
	ctx.SessionStorage().Get("project-details", &p.project.Details)
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
	//load sections of project
}

func (p *Project) Render() app.UI {
	return app.Div().Body(
		app.H1().Text(fmt.Sprintf("Project %v", p.project.Name)),
		app.P().Text(p.project.Details),
		app.If(p.addSectionForm, func() app.UI {
			return app.Div().Body(
				app.Input().Type("text").Placeholder("Title").Value(p.sectionTitleField).OnChange(p.ValueTo(&p.sectionTitleField)),
				app.Button().Text("Add").OnClick(p.addNewSection),
				app.Button().Text("Cancel").OnClick(p.toggleSectionForm),
				app.P().Text(p.errMessage),
			)
		}).ElseIf(p.modProjectForm, func() app.UI {
			return app.Div().Body(
				app.Input().Type("text").Placeholder("Name").Value(p.projectNameField).OnChange(p.ValueTo(&p.projectNameField)),
				app.Input().Type("text").Placeholder("Details").Value(p.projectDetailsField).OnChange(p.ValueTo(&p.projectDetailsField)),
				app.Button().Text("Accept").OnClick(p.modifyProject),
				app.Button().Text("Cancel").OnClick(p.toggleProjectForm),
				app.P().Text(p.errMessage),
			)
		}).Else(func() app.UI {
			return app.Div().Body(
				app.Button().Text("Modify project").OnClick(p.toggleProjectForm),
				app.Button().Text("Delete project").OnClick(p.deleteProject),
				app.P().Text(p.errMessage),
				app.P().Text("Sections"),
				app.Button().Text("Add section").OnClick(p.toggleSectionForm),
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
		app.Log("New project", body.Project)
		p.project = body.Project
		ctx.Update()
		p.projectNameField = ""
		p.projectDetailsField = ""
		p.modProjectForm = false
	} else {
		var body errResponseBody
		if err = json.Unmarshal(b, &body); err != nil {
			app.Log(err)
			p.errMessage = "Could not parse the project modifications"
			return
		}
		app.Log(body)
		p.errMessage = "Could not modify the project"
	}
}

func (p *Project) deleteProject(ctx app.Context, e app.Event) {
	app.Log("Delete:", p.project.Name)
	client := http.Client{}
	req, err := http.NewRequest(
		http.MethodDelete, fmt.Sprintf("%v/project/%v", app.Getenv("BACK_URL"), p.project.Id), nil,
	)
	if err != nil {
		p.errMessage = "Failed creating request to delete project"
		return
	}
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		p.errMessage = "Failed deleting project"
		return
	}
	if res.StatusCode == http.StatusOK {
		ctx.Navigate("dashboard")
	} else {
		app.Log(err)
		p.errMessage = "Failed deleting project"
	}
}

func (p *Project) toggleProjectForm(ctx app.Context, e app.Event) {
	p.projectNameField = ""
	p.projectDetailsField = ""
	p.modProjectForm = !p.modProjectForm
}

func (p *Project) toggleSectionForm(ctx app.Context, e app.Event) {
	p.addSectionForm = !p.addSectionForm
}

func (p *Project) addNewSection(ctx app.Context, e app.Event) {
	p.addSectionForm = !p.addSectionForm
}
