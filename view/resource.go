package view

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gonzabosio/res-manager/model"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Resource struct {
	app.Compo

	editMode       bool
	titleChanges   string
	contentChanges string
	urlChanges     string

	errMessage string
	resource   model.Resource
	project    model.Project
}

func (r *Resource) OnMount(ctx app.Context) {
	if err := ctx.SessionStorage().Get("project", &r.project); err != nil {
		app.Log(fmt.Sprintf("Could not get resource data: %v", err))
	}

	if err := ctx.SessionStorage().Get("resource", &r.resource); err != nil {
		app.Log(fmt.Sprintf("Could not get resource data: %v", err))
	}
	app.Log("Got resource", r.resource)
}

func (r *Resource) Render() app.UI {
	return app.Div().Body(
		app.If(!r.editMode, func() app.UI {
			return app.Div().Body(app.P().Text(r.resource.Title),
				app.Button().Text("Delete").OnClick(r.deleteResource),
				app.Button().Text("Edit").OnClick(func(ctx app.Context, e app.Event) {
					r.titleChanges = r.resource.Title
					r.contentChanges = r.resource.Content
					r.urlChanges = r.resource.URL
					r.editMode = true
				}),
				app.P().Text(r.errMessage),
				app.If(r.resource.URL != "", func() app.UI {
					return app.A().Text(r.resource.URL).Href(r.resource.URL)
				}),
				app.P().Text(r.resource.Content),
			)
		}).Else(func() app.UI {
			return app.Div().Body(
				app.Button().Text("Save").OnClick(r.editResource),
				app.Button().Text("Cancel").OnClick(func(ctx app.Context, e app.Event) {
					r.editMode = false
				}),
				app.P().Text(r.errMessage),
				app.Input().Type("text").Placeholder("Title").Value(r.titleChanges).OnChange(r.ValueTo(&r.titleChanges)),
				app.Input().Type("text").Placeholder("URL").Value(r.urlChanges).OnChange(r.ValueTo(&r.urlChanges)),
				app.Input().Type("text").Placeholder("Content").Value(r.contentChanges).OnChange(r.ValueTo(&r.contentChanges)),
			)
		}),
	)
}

func (r *Resource) deleteResource(ctx app.Context, e app.Event) {
	client := http.Client{}
	req, err := http.NewRequest(
		http.MethodDelete, fmt.Sprintf("%v/resource/%v", app.Getenv("BACK_URL"), r.resource.Id), nil,
	)
	if err != nil {
		r.errMessage = "Failed creating request to delete resource"
		return
	}
	res, err := client.Do(req)
	app.Log(res.Body)
	if err != nil {
		app.Log(err)
		r.errMessage = "Failed to delete resource"
		return
	}
	if res.StatusCode == http.StatusOK {
		app.Log("Resource deleted successfully")
		// ctx.SessionStorage().Set("project", r.project)
		ctx.Navigate("dashboard/project")
	} else {
		app.Log(err)
		r.errMessage = "Failed to delete resource"
	}
}

type resourceResponse struct {
	Message  string              `json:"message"`
	Resource model.PatchResource `json:"resource"`
}

func (r *Resource) editResource(ctx app.Context, e app.Event) {
	app.Log(r.titleChanges, r.contentChanges)
	//save new resource in session storage
	if r.titleChanges == "" {
		r.errMessage = "Title can't be empty"
		return
	}
	var reqBody string
	if r.urlChanges != "" {
		if err := validator.New().Var(r.urlChanges, "url"); err != nil {
			r.errMessage = "Invalid URL"
			return
		}
	}
	reqBody = fmt.Sprintf(`{"id":%v, "title":"%v", "content":"%v", "url":"%v"}`,
		r.resource.Id, r.titleChanges, r.contentChanges, r.urlChanges,
	)
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%v/resource", app.Getenv("BACK_URL")), strings.NewReader(reqBody))
	if err != nil {
		app.Log(err)
		r.errMessage = "Failed modifying resource"
		return
	}
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		r.errMessage = "Failed modifying resource"
		return
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		r.errMessage = "Failed reading the new resource modifications"
		return
	}
	if res.StatusCode == http.StatusOK {
		var body resourceResponse
		if err = json.Unmarshal(b, &body); err != nil {
			app.Log(err)
			r.errMessage = "Could not parse the resource modifications"
			return
		}
		body.Resource.SectionId = r.resource.SectionId
		app.Log("Resource modified", body.Resource)
		ctx.SessionStorage().Set("resource", body.Resource)
		r.errMessage = ""
		r.resource.Title = r.titleChanges
		r.resource.Content = r.contentChanges
		if r.urlChanges != "" {
			r.resource.URL = r.urlChanges
		}
		r.editMode = false
	} else {
		var body errResponseBody
		if err = json.Unmarshal(b, &body); err != nil {
			app.Log(err)
			r.errMessage = "Could not parse the resource modifications"
			return
		}
		app.Log(body.Err)
		r.errMessage = body.Message
	}
}
