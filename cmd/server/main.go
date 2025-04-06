package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"o-techcollaborative.org/website/internal/iconify"
	"o-techcollaborative.org/website/internal/portabletext"
	"o-techcollaborative.org/website/internal/sanity"
	"o-techcollaborative.org/website/internal/utils"
)

type PageData struct {
	Homepage sanity.Homepage
}

func main() {
	sanityClient := sanity.NewClient("5uh7zmgn", "production")

	templateFunctions := utils.TemplateFunctions()
	templateFunctions["RenderPortableText"] = portabletext.Render
	templateFunctions["GetIconSVG"] = iconify.GetSVG

	tmpl, err := template.New("").Funcs(templateFunctions).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=86400") // 24 hours

		homepage, err := sanityClient.FetchHomepage()
		if err != nil {
			http.Error(w, "Failed to fetch homepage data", http.StatusInternalServerError)
			log.Printf("Error fetching homepage data: %v", err)
			return
		}

		data := PageData{
			Homepage: homepage,
		}

		if err := tmpl.ExecuteTemplate(w, "index", data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Error rendering template: %v", err)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
