package view

import (
	"bytes"
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
	"github.com/gonzabosio/res-manager/view/components"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Resource struct {
	app.Compo

	editMode       bool
	titleChanges   string
	contentChanges string
	urlChanges     string

	errMessage  string
	infoMessage string

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
	if err := ctx.LocalStorage().Get("access-token", &r.accessToken); err != nil {
		app.Log(err)
		ctx.Navigate("/")
		return
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
				app.Button().Text("Dashboard").Class("global-btn").OnClick(func(ctx app.Context, e app.Event) {
					r.unlockResource(ctx, r.user.Id, r.resource.Id)
					ctx.Navigate("/dashboard")
				}),
				app.P().Text("Pressing dashboard button you will unlock the edition of this resource").Class("advice-txt"),
				app.P().Text(r.resource.Title).ID("res-title"),
				app.Button().Text("Delete").Class("global-btn").OnClick(r.deleteResource),
				app.Button().Text("Edit").Class("global-btn").OnClick(func(ctx app.Context, e app.Event) {
					r.titleChanges = r.resource.Title
					r.contentChanges = r.resource.Content
					r.urlChanges = r.resource.URL
					r.editMode = true
				}),
				app.P().Text(r.errMessage).Class("err-message"),
				app.If(r.resource.URL != "", func() app.UI {
					return app.A().Text(r.resource.URL).Href(r.resource.URL)
				}),
				app.P().Text(r.resource.Content).Class("content-display"),
				app.P().Text(r.infoMessage).ID("info-message"),
				app.Form().EncType("multipart/form-data").Body(
					app.Input().Type("file").Accept("image/*").Name("image").ID("imageFile"),
					app.Button().Type("submit").Text("Upload").Class("global-btn").OnClick(func(ctx app.Context, e app.Event) {
						e.PreventDefault()
						resIdStr := strconv.Itoa(int(r.resource.Id))
						app.Window().Call("uploadImage", app.Getenv("BACK_URL"), r.accessToken, resIdStr)
					}),
				),
				app.Range(r.imagesList).Slice(func(i int) app.UI {
					imgName := path.Base(r.imagesList[i])
					return app.Div().Body(
						app.A().Href(r.imagesList[i]).ID("img").Body(
							app.Img().Src(r.imagesList[i]).Alt(imgName).Width(200),
						),
						app.Div().ID("below-img").Body(
							app.P().Text(imgName),
							app.Button().Text("Delete").Class("global-btn").OnClick(func(ctx app.Context, e app.Event) {
								r.deleteImage(ctx, imgName)
							}),
						),
					)
				}),
			)
		}).Else(func() app.UI {
			return app.Div().Body(
				app.Button().Text("Save").Class("global-btn").OnClick(r.editResource),
				app.Button().Text("Cancel").Class("global-btn").OnClick(func(ctx app.Context, e app.Event) {
					r.editMode = false
				}),
				app.P().Text(r.errMessage).Class("err-message"),
				app.Input().Type("text").Placeholder("Title").Value(r.titleChanges).OnChange(r.ValueTo(&r.titleChanges)),
				app.Input().Type("text").Placeholder("URL").Value(r.urlChanges).OnChange(r.ValueTo(&r.urlChanges)),
				app.P().Text("Content"),
				app.Div().ID("content-editor-container").Body(
					app.Textarea().
						Text(r.contentChanges).ID("content-editor").
						OnChange(r.ValueTo(&r.contentChanges)),
				),
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
	if err != nil {
		app.Log(err)
		r.errMessage = "Failed to delete resource"
		return
	}
	if res.StatusCode == http.StatusOK {
		ctx.Navigate("dashboard/project")
	} else if res.StatusCode == http.StatusUnauthorized {
		ctx.LocalStorage().Del("access-token")
		ctx.Navigate("/")
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
	reqBody := &model.PatchResource{
		Id: r.resource.Id, Title: r.titleChanges, Content: r.contentChanges, URL: r.urlChanges, LastEditionAt: time.Now(), LastEditionBy: r.user.Username, SectionId: 0,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		r.errMessage = "Failed modifying resource"
		app.Log(err)
		return
	}
	if r.urlChanges != "" {
		if err := validator.New().Var(r.urlChanges, "url"); err != nil {
			r.errMessage = "Invalid URL"
			return
		}
	}
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%v/resource", app.Getenv("BACK_URL")), bytes.NewReader(reqBytes))
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
		// app.Log("Resource modified", r.resource)
		ctx.SessionStorage().Set("resource", r.resource)
		r.errMessage = ""
		r.resource.Title = r.titleChanges
		r.resource.Content = r.contentChanges
		if r.urlChanges != "" {
			r.resource.URL = r.urlChanges
		}
		r.editMode = false

	} else if res.StatusCode == http.StatusUnauthorized {
		ctx.LocalStorage().Del("access-token")
		ctx.Navigate("/")
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
			// app.Log(resBody.Message)
			r.imagesList = resBody.Images
			// app.Log(r.imagesList)
			ctx.Dispatch(func(ctx app.Context) {})
		} else if res.StatusCode == http.StatusUnauthorized {
			ctx.LocalStorage().Del("access-token")
			ctx.Navigate("/")
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
		r.infoMessage = "Image deleted successfully"
	} else if res.StatusCode == http.StatusUnauthorized {
		ctx.LocalStorage().Del("access-token")
		ctx.Navigate("/")
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

func (r *Resource) unlockResource(ctx app.Context, userId, resourceId int64) {
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/resource/unlock", app.Getenv("BACK_URL")), bytes.NewReader([]byte(fmt.Sprintf(`
	{
		"user_id": %d,
		"resource_id": %d
	}`, userId, resourceId,
	))))
	if err != nil {
		app.Logf("Failed to create unlock-resource request: %v\n", err)
		r.errMessage = "Failed to unlock resource"
		return
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.accessToken))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		app.Log(err)
		r.errMessage = "Failed to unlock resource"
		return
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		r.errMessage = "Failed to read unlock resource process"
		return
	}
	if res.StatusCode == http.StatusOK {
		var body components.LockResourceResponse
		if err = json.Unmarshal(b, &body); err != nil {
			app.Log(err)
			r.errMessage = "Failed to parse unlock resource response"
			return
		}
		app.Logf("%s - unlock status: %t", body.Message, body.LockStatus)
		r.errMessage = ""
	} else if res.StatusCode == http.StatusUnauthorized {
		ctx.LocalStorage().Del("access-token")
		ctx.Navigate("/")
	} else {
		var body errResponseBody
		if err = json.Unmarshal(b, &body); err != nil {
			app.Log(err)
			r.errMessage = "Could not parse the section modifications"
			return
		}
		app.Log(body.Err)
		r.errMessage = body.Message
	}
}
