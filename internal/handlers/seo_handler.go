package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"spot-solution/internal/config"
)

type SEOHandler struct {
	cfg *config.Config
}

func NewSEOHandler(cfg *config.Config) *SEOHandler {
	return &SEOHandler{cfg: cfg}
}

func (h *SEOHandler) RobotsTxt(c *gin.Context) {
	content := fmt.Sprintf(`User-agent: *
Allow: /
Disallow: /admin/
Disallow: /api/

Sitemap: %s/sitemap.xml
`, h.cfg.BaseURL)

	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(content))
}

func (h *SEOHandler) SitemapXML(c *gin.Context) {
	now := time.Now().Format("2006-01-02")
	var xml strings.Builder

	xml.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	xml.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	routes := []struct {
		path     string
		priority string
		change   string
	}{
		{"/", "1.0", "weekly"},
		{"/services", "0.9", "weekly"},
		{"/about", "0.8", "monthly"},
		{"/industries", "0.8", "monthly"},
		{"/why-us", "0.8", "monthly"},
		{"/contact", "0.8", "monthly"},
		{"/privacy", "0.3", "yearly"},
		{"/terms", "0.3", "yearly"},
	}

	for _, r := range routes {
		xml.WriteString(fmt.Sprintf(`  <url>
    <loc>%s%s</loc>
    <lastmod>%s</lastmod>
    <changefreq>%s</changefreq>
    <priority>%s</priority>
  </url>
`, h.cfg.BaseURL, r.path, now, r.change, r.priority))
	}

	// Add individual services
	for slug := range ServicesCatalog {
		xml.WriteString(fmt.Sprintf(`  <url>
    <loc>%s/services/%s</loc>
    <lastmod>%s</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.85</priority>
  </url>
`, h.cfg.BaseURL, slug, now))
	}

	xml.WriteString(`</urlset>`)

	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(xml.String()))
}
