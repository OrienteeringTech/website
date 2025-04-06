package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

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

	const cacheControlHeader = "public, max-age=86400" // 24 hours

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", addCacheControlHeader(fs, cacheControlHeader)))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Cache-Control", cacheControlHeader)

		homepage, err := sanityClient.FetchHomepage()
		if err != nil {
			http.Error(w, "Failed to fetch homepage data", http.StatusInternalServerError)
			log.Printf("Error fetching homepage data: %v", err)
			return
		}

		data := PageData{
			Homepage: homepage,
		}

		buf := &bytes.Buffer{}
		if err := tmpl.ExecuteTemplate(buf, "index", data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Error rendering template: %v", err)
			return
		}

		minifiedHTML := minifyHTML(buf.String())

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write([]byte(minifiedHTML)); err != nil {
			log.Printf("Error writing response: %v", err)
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

func minifyHTML(html string) string {
	commentRegex := regexp.MustCompile(`<!--[\s\S]*?-->`)
	html = commentRegex.ReplaceAllString(html, "")

	tagWhitespaceRegex := regexp.MustCompile(`>\s+<`)
	html = tagWhitespaceRegex.ReplaceAllString(html, "><")

	html = strings.TrimSpace(html)

	whitespaceRegex := regexp.MustCompile(`\s{2,}`)
	html = whitespaceRegex.ReplaceAllString(html, " ")

	return html
}

func addCacheControlHeader(h http.Handler, cacheValue string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", cacheValue)
		h.ServeHTTP(w, r)
	})
}
