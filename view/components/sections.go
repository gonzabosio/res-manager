package components

import (
	"bytes"
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

type LockResourceResponse struct {
	Message    string `json:"message"`
	LockStatus bool   `json:"lock_status"`
}

func (s *Sections) OnMount(ctx app.Context) {
	if err := ctx.SessionStorage().Get("user", &s.user); err != nil {
		app.Log("Could not get user from local storage")
	}
	if err := ctx.LocalStorage().Get("access-token", &s.accessToken); err != nil {
		app.Log(err)
		ctx.Navigate("/")
		return
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
		app.Log(err)
		s.errMessage = "Failed getting sections"
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
	} else if res.StatusCode == http.StatusUnauthorized {
		ctx.LocalStorage().Del("access-token")
		ctx.Navigate("/")
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
				app.Button().Text("Create resource").Class("global-btn").OnClick(s.toggleShowResourceForm),
				app.Button().Text("Back to sections").Class("global-btn").OnClick(s.toggleResourcesListView),
				app.Range(s.resourcesList).Slice(func(i int) app.UI {
					it := s.resourcesList[i]
					return app.If(it.LockStatus && s.user.Id != it.LockedBy, func() app.UI {
						// locked
						return app.P().Text(fmt.Sprintf("-%s- edition locked by: %d 🔒", it.Title, it.LockedBy)).
							Class("small-card", "resource-card")
					}).ElseIf(s.user.Id == it.LockedBy, func() app.UI {
						// locked by the current user
						return app.A().Text(fmt.Sprintf("-%s- edition locked by you", it.Title)).Href(fmt.Sprintf("/dashboard/project/res?sid=%d&stitle=%s", s.section.Id, url.QueryEscape(s.section.Title))).
							Class("small-card", "resource-card").
							OnClick(func(ctx app.Context, e app.Event) {
								s.lockResource(ctx, s.user.Id, it.Id)
								ctx.SessionStorage().Set("resource", it)
							})
					}).Else(func() app.UI {
						// unlocked
						return app.A().Text(it.Title).Href(fmt.Sprintf("/dashboard/project/res?sid=%d&stitle=%s", s.section.Id, url.QueryEscape(s.section.Title))).
							Class("small-card", "resource-card").
							OnClick(func(ctx app.Context, e app.Event) {
								e.PreventDefault()
								ctx.Async(func() {
									locked := s.verifyLockStatus(it.Id)
									if locked {
										ctx.Navigate("/dashboard/project")
										app.Window().Call("alert", "Resource locked 🔒")
									} else {
										s.lockResource(ctx, s.user.Id, it.Id)
										ctx.SessionStorage().Set("resource", it)
									}
								})
							})
					})
				}),
			)
		}).ElseIf(s.showResourcesForm, func() app.UI {
			return app.Div().Body(
				app.Input().Type("text").Placeholder("Title").Value(s.resource.Title).OnChange(s.ValueTo(&s.resource.Title)),
				app.Button().Text("Create").Class("global-btn").OnClick(s.addResource),
				app.Button().Text("Cancel").Class("global-btn").OnClick(s.toggleShowResourceForm),
				app.Form().EncType("multipart/form-data").Body(
					app.Input().Type("file").ID("fileInput").Name("file").Accept(".csv"),
					app.Button().Type("submit").Text("Upload CSV").Class("global-btn").OnClick(s.loadResourceDataFromCSV),
				),
			)
		}).ElseIf(s.showSectionForm, func() app.UI {
			return app.Div().Body(
				app.Input().Type("text").Placeholder("Title").Value(s.newSectionTitle).OnChange(s.ValueTo(&s.newSectionTitle)),
				app.Button().Text("Edit").Class("global-btn").OnClick(s.modifySection),
				app.Button().Text("Cancel").Class("global-btn").OnClick(s.toggleUpdateSectionForm),
			)
		}).Else(func() app.UI {
			return app.Div().Body(
				app.Range(s.sectionsList).Slice(func(i int) app.UI {
					return app.Div().Class("small-card").OnClick(func(ctx app.Context, e app.Event) {
						s.toggleResourcesListView(ctx, e)
						s.section.Id = s.sectionsList[i].Id
						s.section.Title = s.sectionsList[i].Title
						s.loadResources(ctx)
					}).Body(
						app.P().Text(s.sectionsList[i].Title),
						app.Button().Text("Edit").Class("global-btn").OnClick(func(ctx app.Context, e app.Event) {
							e.StopImmediatePropagation()
							s.section = s.sectionsList[i]
							s.sectionIdx = i
							s.toggleUpdateSectionForm(ctx, e)
						}),
						app.Button().Text("Delete").Class("global-btn").OnClick(func(ctx app.Context, e app.Event) {
							e.StopImmediatePropagation()
							s.section = s.sectionsList[i]
							s.sectionIdx = i
							s.deleteSection(ctx)
						}),
					)
				}),
			)
		}),
		app.P().Text(s.errMessage).Class("err-message"),
	)
}

func (s *Sections) loadResources(ctx app.Context) {
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
		app.Log(err)
		s.errMessage = "Failed getting resources"
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
	} else if res.StatusCode == http.StatusUnauthorized {
		ctx.LocalStorage().Del("access-token")
		ctx.Navigate("/")
	} else {
		var errRes errResponseBody
		s.errMessage = errRes.Message
		app.Log(errRes.Err)
	}
}

func (s *Sections) addResource(ctx app.Context, e app.Event) {
	reqUrl := fmt.Sprintf("%v/resource/%v", app.Getenv("BACK_URL"), s.user.Id)
	reqBody := fmt.Sprintf(`{"title":"%v","last_edition_by":"%v","section_id":%d}`, s.resource.Title, s.user.Username, s.section.Id)
	req, err := http.NewRequest(http.MethodPost, reqUrl, strings.NewReader(reqBody))
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
	} else if res.StatusCode == http.StatusUnauthorized {
		ctx.LocalStorage().Del("access-token")
		ctx.Navigate("/")
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

func (s *Sections) deleteSection(ctx app.Context) {
	// app.Log("Delete section", s.section, s.sectionIdx)
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
	} else if res.StatusCode == http.StatusUnauthorized {
		ctx.LocalStorage().Del("access-token")
		ctx.Navigate("/")
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
	// app.Log(s.section)
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
		s.section = body.Section
		s.sectionsList[s.sectionIdx] = s.section
		s.errMessage = ""
		s.newSectionTitle = ""
		s.showSectionForm = false
	} else if res.StatusCode == http.StatusUnauthorized {
		ctx.LocalStorage().Del("access-token")
		ctx.Navigate("/")
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

func (s *Sections) loadResourceDataFromCSV(ctx app.Context, e app.Event) {
	type resourceObj struct {
		ID json.Number `json:"resource_id"`
	}
	e.PreventDefault()
	sectionIdStr := strconv.Itoa(int(s.section.Id))
	userIdStr := strconv.Itoa(int(s.user.Id))
	app.Window().Call("uploadCSV", app.Getenv("BACK_URL"), s.accessToken, s.user.Username, sectionIdStr, userIdStr)
	var resObj resourceObj
	time.Sleep(3 * time.Second)
	if err := ctx.SessionStorage().Get("resource-id", &resObj.ID); err != nil {
		app.Log(err)
		s.errMessage = "Failed to create resource"
		return
	}
	resourceID, err := strconv.ParseInt(resObj.ID.String(), 10, 64)
	if err != nil {
		app.Log("Failed to convert resource ID to int64:", err)
		s.errMessage = "Failed to create resource"
		return
	}
	ctx.SessionStorage().Get("resource", &s.resource)
	s.resource.Id = resourceID
	ctx.SessionStorage().Set("resource", &s.resource)
	ctx.SessionStorage().Del("resource-id")
	ctx.Navigate(fmt.Sprintf("/dashboard/project/res?sid=%d&stitle=%s", s.section.Id, url.QueryEscape(s.section.Title)))
}

func (s *Sections) lockResource(ctx app.Context, userId, resourceId int64) {
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/resource/lock", app.Getenv("BACK_URL")), bytes.NewReader([]byte(fmt.Sprintf(`
		{
			"user_id": %d,
			"resource_id": %d
		}`, userId, resourceId,
	))))
	if err != nil {
		app.Logf("Failed to create lock-resource request: %v\n", err)
		s.errMessage = "Failed to lock resource"
		return
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		s.errMessage = "Failed to lock resource"
		return
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		s.errMessage = "Failed to read lock resource process"
		return
	}
	if res.StatusCode == http.StatusOK {
		var body LockResourceResponse
		if err = json.Unmarshal(b, &body); err != nil {
			app.Log(err)
			s.errMessage = "Failed to parse lock resource response"
			return
		}
		app.Logf("%s - lock status: %t", body.Message, body.LockStatus)
		s.errMessage = ""
	} else if res.StatusCode == http.StatusUnauthorized {
		ctx.LocalStorage().Del("access-token")
		ctx.Navigate("/")
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

func (s *Sections) verifyLockStatus(resourceId int64) bool {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/resource/%d/lock", app.Getenv("BACK_URL"), resourceId), nil)
	if err != nil {
		s.errMessage = "Failed to check lock status"
		app.Log(err)
		return false
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		s.errMessage = "Failed to check lock status"
		app.Log(err)
		return false
	}
	resBody := make(map[string]interface{})
	b, err := io.ReadAll(res.Body)
	if err != nil {
		app.Log(err)
		return false
	}
	if err := json.Unmarshal(b, &resBody); err != nil {
		app.Log(err)
		return false
	}
	lockStatus := resBody["lock_status"].(bool)
	return lockStatus
}
