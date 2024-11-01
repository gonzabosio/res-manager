package components

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gonzabosio/res-manager/model"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Sections struct {
	app.Compo

	showResourcesList bool
	showResourcesForm bool
	showSectionForm   bool

	section         model.Section
	sectionsList    []model.Section
	sectionIdx      int
	newSectionTitle string

	errMessage    string
	user          model.User
	project       model.Project
	resource      model.Resource
	resourcesList []model.Resource

	accessToken string
}

type errResponseBody struct {
	Message string `json:"message"`
	Err     string `json:"error"`
}

type sectionsResponse struct {
	Message  string          `json:"message"`
	Sections []model.Section `json:"sections"`
}

type sectionResponse struct {
	Message string        `json:"message"`
	Section model.Section `json:"section"`
}

type resourcesResponse struct {
	Message   string           `json:"message"`
	Resources []model.Resource `json:"resources"`
}

type resourceResponse struct {
	Message  string         `json:"message"`
	Resource model.Resource `json:"resource"`
}

func (s *Sections) OnMount(ctx app.Context) {
	if err := ctx.SessionStorage().Get("user", &s.user); err != nil {
		app.Log("Could not get user from local storage")
	}
	atCookie := app.Window().Call("getAccessTokenCookie")
	if atCookie.IsUndefined() {
		ctx.Navigate("/")
	} else {
		s.accessToken = atCookie.String()
	}
	ctx.SessionStorage().Get("project", &s.project)
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/section/%v", app.Getenv("BACK_URL"), s.project.Id), nil)
	if err != nil {
		app.Log(err)
		s.errMessage = "Could not get sections"
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		s.errMessage = fmt.Sprintf("Failed getting sections: %v", err)
		return
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		app.Log(err)
		s.errMessage = "Failed reading sections"
		return
	}
	if res.StatusCode == http.StatusOK {
		var resBody sectionsResponse
		if err = json.Unmarshal(b, &resBody); err != nil {
			app.Log(err)
			s.errMessage = "Failed parsing sections data"
		}
		s.sectionsList = resBody.Sections
	} else {
		var resBody errResponseBody
		if err = json.Unmarshal(b, &resBody); err != nil {
			app.Log(err)
			s.errMessage = "Failed parsing sections data"
		}
		s.errMessage = resBody.Message
		app.Log(resBody.Err)
	}
}

func (s *Sections) Render() app.UI {
	return app.Div().Body(
		app.If(s.showResourcesList, func() app.UI {
			return app.Div().Body(
				app.Button().Text("Create resource").OnClick(s.toggleShowResourceForm),
				app.Button().Text("Back to sections").OnClick(s.toggleResourcesListView),
				app.Range(s.resourcesList).Slice(func(i int) app.UI {
					return app.Div().Body(
						app.A().Text(s.resourcesList[i].Title).Href(fmt.Sprintf("/dashboard/project/res?sid=%d&stitle=%s", s.section.Id, url.QueryEscape(s.section.Title)))).OnClick(func(ctx app.Context, e app.Event) {
						ctx.SessionStorage().Set("resource", s.resourcesList[i])
					})
				}),
			)
		}).ElseIf(s.showResourcesForm, func() app.UI {
			return app.Div().Body(
				app.Input().Type("text").Placeholder("Title").Value(s.resource.Title).OnChange(s.ValueTo(&s.resource.Title)),
				app.Button().Text("Create").OnClick(s.addResource),
				app.Button().Text("Cancel").OnClick(s.toggleShowResourceForm),
				app.Form().EncType("multipart/form-data").Body(
					app.Input().Type("file").ID("fileInput").Name("file").Accept(".csv"),
					app.Button().Type("submit").Text("Upload CSV").OnClick(func(ctx app.Context, e app.Event) {
						e.PreventDefault()
						sectionIdStr := strconv.Itoa(int(s.section.Id))
						app.Window().Call("uploadCSV", app.Getenv("BACK_URL"), s.accessToken, s.user.Username, sectionIdStr)
						time.Sleep(500 * time.Millisecond)
						ctx.Navigate(fmt.Sprintf("/dashboard/project/res?sid=%d&stitle=%s", s.section.Id, url.QueryEscape(s.section.Title)))
					}),
				),
			)
		}).ElseIf(s.showSectionForm, func() app.UI {
			return app.Div().Body(
				app.Input().Type("text").Placeholder("Title").Value(s.newSectionTitle).OnChange(s.ValueTo(&s.newSectionTitle)),
				app.Button().Text("Edit").OnClick(s.modifySection),
				app.Button().Text("Cancel").OnClick(s.toggleUpdateSectionForm),
			)
		}).Else(func() app.UI {
			return app.Div().Body(
				app.Range(s.sectionsList).Slice(func(i int) app.UI {
					return app.Div().OnClick(func(ctx app.Context, e app.Event) {
						s.toggleResourcesListView(ctx, e)
						s.section.Id = s.sectionsList[i].Id
						s.section.Title = s.sectionsList[i].Title
						s.loadResources(ctx, e)
					}).Body(
						app.P().Text(s.sectionsList[i].Title),
						app.Button().Text("Edit").OnClick(func(ctx app.Context, e app.Event) {
							e.StopImmediatePropagation()
							s.section = s.sectionsList[i]
							s.sectionIdx = i
							s.toggleUpdateSectionForm(ctx, e)
						}),
						app.Button().Text("Delete").OnClick(func(ctx app.Context, e app.Event) {
							e.StopImmediatePropagation()
							s.section = s.sectionsList[i]
							s.sectionIdx = i
							s.deleteSection(ctx, e)
						}),
					)
				}),
			)
		}),
		app.P().Text(s.errMessage),
	)
}

func (s *Sections) loadResources(ctx app.Context, e app.Event) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/resource/%v", app.Getenv("BACK_URL"), s.section.Id), nil)
	if err != nil {
		s.errMessage = "Could not load resources"
		app.Log(err)
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		s.errMessage = fmt.Sprintf("Failed getting resources: %v", err)
		return
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		app.Log(err)
		s.errMessage = "Failed reading resources"
		return
	}
	if res.StatusCode == http.StatusOK {
		var resBody resourcesResponse
		if err = json.Unmarshal(b, &resBody); err != nil {
			app.Log(err)
			s.errMessage = "Failed parsing resources data"
		}
		s.resourcesList = resBody.Resources
	} else {
		var errRes errResponseBody
		s.errMessage = errRes.Message
		app.Log(errRes.Err)
	}
}

func (s *Sections) addResource(ctx app.Context, e app.Event) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/resource", app.Getenv("BACK_URL")), strings.NewReader(fmt.Sprintf(
		`{"title":"%v","last_edition_by":"%v","section_id":%d}`,
		s.resource.Title, s.user.Username, s.section.Id,
	)))
	if err != nil {
		app.Log(err)
		s.errMessage = "Could not add resource"
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)

	if err != nil {
		app.Log(err)
		s.errMessage = "Failed to create new resource"
		return
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Failed to read response body:", err)
		return
	}
	if res.StatusCode == http.StatusOK {
		var resBody resourceResponse
		if err := json.Unmarshal(b, &resBody); err != nil {
			app.Log(err)
			s.errMessage = "Failed to parse resource data"
			return
		}
		s.resource.SectionId = s.section.Id
		s.resource.Id = resBody.Resource.Id
		ctx.SessionStorage().Set("resource", s.resource)
		ctx.Navigate(fmt.Sprintf("/dashboard/project/res?sid=%d&stitle=%s", s.section.Id, url.QueryEscape(s.section.Title)))
		// ctx.Navigate("/dashboard/project/res")
	} else {
		app.Log("Request failed with status:", res.StatusCode)
		var resBody errResponseBody
		if err := json.Unmarshal(b, &resBody); err != nil {
			s.errMessage = "Failed to parse json"
			return
		}
		app.Log(resBody.Err)
		s.errMessage = resBody.Message
	}
}

func (s *Sections) toggleResourcesListView(ctx app.Context, e app.Event) {
	s.showResourcesForm = false
	s.showSectionForm = false
	s.showResourcesList = !s.showResourcesList
}

func (s *Sections) toggleShowResourceForm(ctx app.Context, e app.Event) {
	s.resource.Title = ""
	s.errMessage = ""
	s.showSectionForm = false
	s.showResourcesForm = !s.showResourcesForm
	if !s.showResourcesForm {
		s.showResourcesList = !s.showResourcesList
	} else {
		s.showResourcesList = false
	}
}

func (s *Sections) deleteSection(ctx app.Context, e app.Event) {
	app.Log("Delete section", s.section, s.sectionIdx)
	client := http.Client{}
	req, err := http.NewRequest(
		http.MethodDelete, fmt.Sprintf("%v/section/%v", app.Getenv("BACK_URL"), s.section.Id), nil,
	)
	if err != nil {
		s.errMessage = "Failed creating request to delete section"
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	app.Log(res.Body)
	if err != nil {
		app.Log(err)
		s.errMessage = "Failed to delete section"
		return
	}
	if res.StatusCode == http.StatusOK {
		s.sectionsList = slices.Delete(s.sectionsList, s.sectionIdx, s.sectionIdx)
		ctx.Reload()
	} else {
		app.Log(err)
		s.errMessage = "Failed to delete section"
	}
}

func (s *Sections) toggleUpdateSectionForm(ctx app.Context, e app.Event) {
	s.showResourcesForm = false
	s.showResourcesList = false
	s.newSectionTitle = ""
	s.showSectionForm = !s.showSectionForm
}

func (s *Sections) modifySection(ctx app.Context, e app.Event) {
	app.Log(s.section)
	if s.newSectionTitle == "" {
		s.errMessage = "The field is empty"
		return
	}
	reqBody := fmt.Sprintf(`{"id":%v,"title":"%v"}`, s.section.Id, s.newSectionTitle)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%v/section", app.Getenv("BACK_URL")), strings.NewReader(reqBody))
	if err != nil {
		app.Log(err)
		s.errMessage = "Failed modifying section"
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		s.errMessage = "Failed modifying section"
		return
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		s.errMessage = "Failed reading the new section modifications"
		return
	}
	if res.StatusCode == http.StatusOK {
		var body sectionResponse
		if err = json.Unmarshal(b, &body); err != nil {
			app.Log(err)
			s.errMessage = "Could not parse the section modifications"
			return
		}
		app.Log("Section modified", body.Section)
		s.section = body.Section
		s.sectionsList[s.sectionIdx] = s.section
		s.errMessage = ""
		s.newSectionTitle = ""
		s.showSectionForm = false
	} else {
		var body errResponseBody
		if err = json.Unmarshal(b, &body); err != nil {
			app.Log(err)
			s.errMessage = "Could not parse the section modifications"
			return
		}
		app.Log(body.Err)
		s.errMessage = body.Message
	}
}
