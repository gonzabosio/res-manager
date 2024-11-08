package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/gonzabosio/res-manager/model"
)

func exitErrorf(w http.ResponseWriter, msg string, args ...interface{}) {
	WriteJSON(w, map[string]interface{}{
		"message": fmt.Sprintf(msg+"\n", args...),
	}, http.StatusInternalServerError)
}

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	file, fileHeader, err := r.FormFile("image")
	resourceId := r.FormValue("resourceId")
	if err != nil {
		WriteJSON(w, map[string]string{
			"message": "Could not get image from request",
			"error":   err.Error(),
		}, http.StatusInternalServerError)
		return
	}
	filename := fileHeader.Filename
	// S3 ↓
	svc := s3.New(h.S3.Session)

	result, err := svc.ListBuckets(nil)
	if err != nil {
		exitErrorf(w, "Unable to list buckets, %v", err)
		return
	}

	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}

	uploader := s3manager.NewUploader(h.S3.Session)

	res, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(h.S3.Bucket),
		Key:    aws.String(filename),
		Body:   file,
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		exitErrorf(w, "Unable to upload %q to %q, %v", filename, h.S3.Bucket, err)
		return
	}
	resId, err := strconv.ParseInt(resourceId, 10, 64)
	if err != nil {
		WriteJSON(w, map[string]string{
			"message": "Failed to parse resource id to int64",
			"error":   err.Error(),
		}, http.StatusInternalServerError)
		return
	}
	err = h.Service.SaveImageURL(res.Location, resId)
	if err != nil {
		WriteJSON(w, map[string]string{
			"message": "Failed to save url image",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message":        fmt.Sprintf("Successfully uploaded %q to %q", filename, h.S3.Bucket),
		"image_location": res.Location,
	}, http.StatusOK)
}

func (h *Handler) GetImages(w http.ResponseWriter, r *http.Request) {
	resIdStr := chi.URLParam(r, "resource-id")
	resId, err := strconv.ParseInt(resIdStr, 10, 64)
	if err != nil {
		WriteJSON(w, map[string]string{
			"message": "Failed to parse resource id to int64",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	images, err := h.Service.GetImagesByResourceID(resId)
	if err != nil {
		if err.Error() == "no images found" {
			WriteJSON(w, map[string]string{
				"message": "No images found in the resource",
				"error":   err.Error(),
			}, http.StatusNoContent)
			return
		} else {
			WriteJSON(w, map[string]string{
				"message": "Failed to get images",
				"error":   err.Error(),
			}, http.StatusBadRequest)
			return
		}
	}
	WriteJSON(w, map[string]interface{}{
		"message": "Images retrieved successfully",
		"images":  images,
	}, http.StatusOK)
}

func (h *Handler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	var reqBody model.DeleteImageReq
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		WriteJSON(w, map[string]string{
			"message": "Failed to decode request body to delete image",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err = validate.Struct(reqBody); err != nil {
		errors := err.(validator.ValidationErrors)
		WriteJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	obj := reqBody.ImageName
	svc := s3.New(h.S3.Session)
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(h.S3.Bucket), Key: aws.String(obj)})
	if err != nil {
		WriteJSON(w, map[string]string{
			"message": "Failed to delete object in S3",
			"error":   err.Error(),
		}, http.StatusInternalServerError)
		exitErrorf(w, "Unable to delete object %q from bucket %q, %v", obj, h.S3.Bucket, err)
		return
	}
	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(h.S3.Bucket),
		Key:    aws.String(obj),
	})
	if err != nil {
		WriteJSON(w, map[string]string{
			"message": "Failed to wait object elimination",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	err = h.Service.DeleteImageByResourceID(reqBody.ImageName, reqBody.ResourceId)
	if err != nil {
		WriteJSON(w, map[string]string{
			"message": "Failed to delete image from database",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, map[string]string{
		"message": "Image deleted successfully",
	}, http.StatusOK)
}
