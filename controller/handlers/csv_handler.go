package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gonzabosio/res-manager/model"
)

func (h *Handler) UploadCSV(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // max upload size: 10 MB
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	// retrieve the file from form data
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// read the file content
	csvReader := csv.NewReader(file)
	newResource := new(model.Resource)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "Error reading CSV file", http.StatusInternalServerError)
			return
		}
		fmt.Printf("Record: %v\n", record)
		for i, field := range record {
			if i == 0 {
				switch strings.ToLower(field) {
				case "title":
					i++
					newResource.Title = record[i]
				case "content":
					i++
					newResource.Content = record[i]
				case "url":
					i++
					newResource.URL = record[i]
				case "images":
					var images []string
					for j := 1; j < len(record); j++ {
						images = append(images, record[j])
					}
					log.Println("All images:", images)
					newResource.Images = images
				}
			}
		}
	}
	lastEditionBy := r.FormValue("lastEditionBy")
	newResource.LastEditionBy = lastEditionBy
	strId := r.FormValue("sectionId")
	sectionId, err := strconv.ParseInt(strId, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not parse section id: %v", err), http.StatusInternalServerError)
		return
	}
	newResource.SectionId = sectionId
	if err := h.Service.CreateResource(newResource); err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("could not save resource: %v", err), http.StatusBadRequest)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message":  "Resources created successfully by csv file",
		"resource": newResource,
	}, http.StatusOK)
}
