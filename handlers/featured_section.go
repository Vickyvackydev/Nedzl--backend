package handlers

import (
	"api/models"
	"api/utils"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func UpdateFeaturedSection(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		boxNumber, _ := strconv.Atoi(c.Param("box_number"))

		var req struct {
			CategoryName string   `json:"category_name"`
			Description  string   `json:"description"`
			ProductIDS   []string `json:"product_ids"`
		}

		if err := c.Bind(&req); err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid Request body", err)
		}

		// validation rules

		if (boxNumber == 1 || boxNumber == 2) && len(req.ProductIDS) < 3 {
			return utils.ResponseError(c, http.StatusBadRequest, "This category requires at least 3 products", nil)
		}

		if (boxNumber == 3 || boxNumber == 4) && len(req.ProductIDS) < 2 {
			return utils.ResponseError(c, http.StatusBadRequest, "This category requires at least 2 products", nil)

		}

		var section models.FeaturedSection

		if err := db.Where("box_number = ?", boxNumber).First(&section).Error; err != nil {
			section = models.FeaturedSection{

				BoxNumber:    boxNumber,
				CategoryName: req.CategoryName,
				Description:  req.Description,
			}
			db.Create(&section)
		} else {
			db.Model(&section).Updates(&models.FeaturedSection{
				CategoryName: req.CategoryName,
				Description:  req.Description,
			})
		}

		// remove previous product

		db.Where("feature_section_id = ?", section.ID).Delete(&models.FeaturedSectionProduct{})

		// insert new product

		for _, pid := range req.ProductIDS {
			db.Create(&models.FeaturedSectionProduct{
				FeaturedSectionID: section.ID,
				ProductID:         uuid.MustParse(pid),
			})

		}
		// RELOAD SECTION WITH PRODUCTS
		db.Preload("Products.Product").
			First(&section, "id = ?", section.ID)

		return utils.ResponseSucess(c, http.StatusOK, "Feature section updated", section)
	}
}

func GetFeaturedSections(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {

		var sections []models.FeaturedSection
		// Get all existing sections (1–4)
		if err := db.Order("box_number ASC").Find(&sections).Error; err != nil {
			return utils.ResponseError(c, 500, "Failed to fetch sections", err)
		}

		// Prepare response structure for all 4 boxes
		response := make([]map[string]interface{}, 0)

		for box := 1; box <= 4; box++ {
			var sec models.FeaturedSection
			found := false

			// Find section matching box number
			for _, s := range sections {
				if s.BoxNumber == box {
					sec = s
					found = true
					break
				}
			}

			// If section does not exist yet → return empty placeholder
			if !found {
				response = append(response, map[string]interface{}{
					"box_number":    box,
					"category_name": nil,
					"description":   nil,
					"products":      []interface{}{},
				})
				continue
			}

			// Get products for this section
			var secProducts []models.FeaturedSectionProduct
			db.Where("featured_section_id = ?", sec.ID).
				Order("created_at ASC").
				Find(&secProducts)

			// Fetch actual product details for each productID
			var productDetails []models.Products
			productIDs := make([]uuid.UUID, 0)

			for _, sp := range secProducts {
				productIDs = append(productIDs, sp.ProductID)
			}

			if len(productIDs) > 0 {
				db.Where("id IN ?", productIDs).Find(&productDetails)
			}

			// Assemble section response
			response = append(response, map[string]interface{}{
				"box_number":    sec.BoxNumber,
				"category_name": sec.CategoryName,
				"description":   sec.Description,
				"products":      productDetails,
			})
		}

		return utils.ResponseSucess(c, 200, "Featured sections retrieved", response)
	}
}
