package handlers

import (
	"api/models"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// FacebookRSS represents the top-level RSS structure for Meta catalogs
type FacebookRSS struct {
	XMLName xml.Name        `xml:"rss"`
	Version string          `xml:"version,attr"`
	NS      string          `xml:"xmlns:g,attr"`
	Channel FacebookChannel `xml:"channel"`
}

type FacebookChannel struct {
	Title       string         `xml:"title"`
	Link        string         `xml:"link"`
	Description string         `xml:"description"`
	Items       []FacebookItem `xml:"item"`
}

type FacebookItem struct {
	ID             string `xml:"g:id"`
	Title          string `xml:"g:title"`
	Description    string `xml:"g:description"`
	Link           string `xml:"g:link"`
	ImageLink      string `xml:"g:image_link"`
	Brand          string `xml:"g:brand"`
	Condition      string `xml:"g:condition"`
	Availability   string `xml:"g:availability"`
	Price          string `xml:"g:price"`
	GoogleCategory string `xml:"g:google_product_category"`
}

// GetFacebookProductFeed generates an XML feed for Facebook/Meta catalogs
func GetFacebookProductFeed(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var products []models.Products

		// Fetch all products that are currently "ONGOING" (active)
		if err := db.Where("status = ?", models.StatusOngoing).Find(&products).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch products"})
		}

		// determine base URLs from environment or use defaults
		baseURL := os.Getenv("FRONTEND_URL")
		if baseURL == "" {
			baseURL = "https://nedzl.com" // Default fallback
		}

		feed := FacebookRSS{
			Version: "2.0",
			NS:      "http://base.google.com/ns/1.0",
			Channel: FacebookChannel{
				Title:       "Nedzl Market Product Feed",
				Link:        baseURL,
				Description: "Latest products from Nedzl Market for Facebook/Meta Catalog",
				Items:       make([]FacebookItem, 0, len(products)),
			},
		}

		for _, p := range products {
			// Extract the first image URL from JSON
			var images []string
			imageLink := ""
			if err := json.Unmarshal(p.ImageUrls, &images); err == nil && len(images) > 0 {
				imageLink = images[0]
			}

			// Format condition
			condition := "new"
			if p.Condition != "" {
				condition = p.Condition // Meta expects: new, refurbished, used
			}

			item := FacebookItem{
				ID:           p.ID.String(),
				Title:        p.Name,
				Description:  p.Description,
				Link:         fmt.Sprintf("%s/products/%s", baseURL, p.ID.String()),
				ImageLink:    imageLink,
				Brand:        p.BrandName,
				Condition:    condition,
				Availability: "in stock",
				Price:        fmt.Sprintf("%.2f NGN", p.ProductPrice), // Adjust currency if needed
			}

			// Add more mapping logic if you have specific categories
			item.GoogleCategory = p.CategoryName

			feed.Channel.Items = append(feed.Channel.Items, item)
		}

		return c.XML(http.StatusOK, feed)
	}
}
