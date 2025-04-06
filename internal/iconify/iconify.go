package iconify

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type IconCache struct {
	SVG       template.HTML
	FetchedAt time.Time
}

var iconCache = make(map[string]IconCache)

func GetSVG(iconName string, size int) template.HTML {
	if iconName == "" {
		return ""
	}

	cacheKey := fmt.Sprintf("%s-%d", iconName, size)

	if cached, ok := iconCache[cacheKey]; ok {
		if time.Since(cached.FetchedAt) < 7*24*time.Hour {
			return cached.SVG
		}
	}

	parts := strings.Split(iconName, ":")
	if len(parts) != 2 {
		log.Printf("Invalid icon format: %s", iconName)
	}

	prefix, name := parts[0], parts[1]

	iconURL := fmt.Sprintf("https://api.iconify.design/%s/%s.svg?height=%d", prefix, name, size)

	req, err := http.NewRequest("GET", iconURL, nil)
	if err != nil {
		log.Printf("Error creating request for icon %s: %v", iconName, err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching icon %s: %v", iconName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error fetching icon %s: status code %d", iconName, resp.StatusCode)
	}

	svgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading icon %s: %v", iconName, err)
	}

	svgHTML := template.HTML(svgBytes)
	iconCache[cacheKey] = IconCache{
		SVG:       svgHTML,
		FetchedAt: time.Now(),
	}

	return svgHTML
}
