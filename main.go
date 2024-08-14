package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/vertexai/genai"
)

type AnimalForm struct {
	Animal string
}

func main() {
	ctx := context.Background()
	var projectId string
	var err error
	projectId = os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectId == "" {
		projectId, err = metadata.ProjectIDWithContext(ctx)
		if err != nil {
			return
		}
	}
	var client *genai.Client
	client, err = genai.NewClient(ctx, projectId, "us-central1")
	if err != nil {
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash-001")

	// Load the HTML template
	tmpl, err := template.ParseFiles("templates/form.html")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Handle form submission
			r.ParseForm()
			animal := r.FormValue("animal")

			resp, err := model.GenerateContent(
				ctx,
				genai.Text(
					fmt.Sprintf("Give me 10 fun facts about %s. Return the results as HTML without markdown backticks.", animal)),
			)

			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
				htmlContent := resp.Candidates[0].Content.Parts[0]
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				fmt.Fprint(w, htmlContent)
			}
		} else {
			// Display the form
			form := AnimalForm{}
			err := tmpl.Execute(w, form)
			if err != nil {
				panic(err)
			}
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)
}
