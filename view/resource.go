package view

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

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
	user       model.User
	resource   model.Resource
	project    model.Project
	imagesList []string

	accessToken string
}

func (r *Resource) OnMount(ctx app.Context) {
	if err := ctx.SessionStorage().Get("user", &r.user); err != nil {
		app.Log("Could not get user data from session storage")
	}
	atCookie := app.Window().Call("getAccessTokenCookie")
	if atCookie.IsUndefined() {
		ctx.Navigate("/")
	} else {
		r.accessToken = atCookie.String()
	}
	if err := ctx.SessionStorage().Get("project", &r.project); err != nil {
		app.Log(fmt.Sprintf("Could not get resource data: %v", err))
	}

	if err := ctx.SessionStorage().Get("resource", &r.resource); err != nil {
		app.Log(fmt.Sprintf("Could not get resource data: %v", err))
	}
	r.getImages(ctx)
}

func (r *Resource) Render() app.UI {
	return app.Div().Body(
		app.If(!r.editMode, func() app.UI {
			return app.Div().Body(
				app.Button().Text("Dashboard").OnClick(func(ctx app.Context, e app.Event) {
					ctx.Navigate("/dashboard")
				}),
				app.P().Text(r.resource.Title),
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
				app.Form().EncType("multipart/form-data").Body(
					app.Input().Type("file").Accept("image/*").Name("image").ID("imageFile"),
					app.Button().Type("submit").Text("Upload").OnClick(func(ctx app.Context, e app.Event) {
						e.PreventDefault()
						resIdStr := strconv.Itoa(int(r.resource.Id))
						app.Window().Call("uploadImage", app.Getenv("BACK_URL"), r.accessToken, resIdStr)
					}),
				),
				app.Range(r.imagesList).Slice(func(i int) app.UI {
					imgName := path.Base(r.imagesList[i])
					return app.Div().Body(
						app.A().Href(r.imagesList[i]).Body(
							app.Img().Src(r.imagesList[i]).Alt(imgName).Width(200),
						),
						app.Button().Text("Delete").OnClick(func(ctx app.Context, e app.Event) {
							r.deleteImage(ctx, imgName)
						}),
					)
				}),
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
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.accessToken))
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	app.Log(res.Body)
	if err != nil {
		app.Log(err)
		r.errMessage = "Failed to delete resource"
		return
	}
	if res.StatusCode == http.StatusOK {
		app.Log("Resource deleted successfully")
		ctx.Navigate("dashboard/project")
	} else {
		app.Log(err)
		r.errMessage = "Failed to delete resource"
	}
}

type resourceResponse struct {
	Message  string         `json:"message"`
	Resource model.Resource `json:"resource"`
}

func (r *Resource) editResource(ctx app.Context, e app.Event) {
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
	app.Log(r.titleChanges, r.contentChanges, r.urlChanges, time.Now(), r.user.Username)
	reqBody = fmt.Sprintf(`{"id":%v, "title":"%v", "content":"%v", "url":"%v", "last_edition_at":"%v", "last_edition_by":"%v"}`,
		r.resource.Id, r.titleChanges, r.contentChanges, r.urlChanges, time.Now(), r.user.Username,
	)
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%v/resource", app.Getenv("BACK_URL")), strings.NewReader(reqBody))
	if err != nil {
		app.Log(err)
		r.errMessage = "Failed modifying resource"
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.accessToken))
	req.Header.Add("Content-Type", "application/json")
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
		r.resource = body.Resource
		app.Log("Resource modified", r.resource)
		ctx.SessionStorage().Set("resource", r.resource)
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

type imagesListRes struct {
	Message string   `json:"message"`
	Images  []string `json:"images"`
}

func (r *Resource) getImages(ctx app.Context) {
	ctx.Async(func() {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/image/%v", app.Getenv("BACK_URL"), r.resource.Id), nil)
		if err != nil {
			r.errMessage = "Failed to get images of resource"
			app.Log(err)
			return
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.accessToken))
		req.Header.Add("Content-Type", "application/json")
		client := http.Client{}
		res, err := client.Do(req)
		if err != nil {
			r.errMessage = "Failed to send the request to get images"
			app.Log(err)
			return
		}
		if res.StatusCode == http.StatusOK {
			var resBody imagesListRes
			json.NewDecoder(res.Body).Decode(&resBody)
			app.Log(resBody.Message)
			r.imagesList = resBody.Images
			app.Log(r.imagesList)
			ctx.Dispatch(func(ctx app.Context) {})
		} else {
			var errResBody errResponseBody
			json.NewDecoder(res.Body).Decode(&errResBody)
			r.errMessage = errResBody.Message
			app.Log(errResBody.Err)
		}
	})
}

func (r *Resource) deleteImage(ctx app.Context, imgName string) {
	req, err := http.NewRequest(
		http.MethodDelete, fmt.Sprintf("%v/image", app.Getenv("BACK_URL")),
		strings.NewReader(fmt.Sprintf(`{"image":"%v", "resource_id": %d}`, imgName, r.resource.Id)),
	)
	if err != nil {
		app.Log(err)
		r.errMessage = fmt.Sprintf("Failed to delete %s", imgName)
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		r.errMessage = fmt.Sprintf("Failed to delete %s", imgName)
		return
	}
	if res.StatusCode == http.StatusOK {
		app.Log("Image deleted successfully")
		ctx.Dispatch(func(ctx app.Context) {})
	} else {
		var errResBody errResponseBody
		err := json.NewDecoder(res.Body).Decode(&errResBody)
		if err != nil {
			r.errMessage = "Failed to read error response"
			app.Log(err)
			return
		}
		r.errMessage = errResBody.Message
		app.Log(err)
	}
}
