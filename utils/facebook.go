package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

// PostToFacebook sends a message and a link to a configured Facebook Page.
func PostToFacebook(message string, imageUrl string, link string) error {
	pageID := os.Getenv("FB_PAGE_ID")
	accessToken := os.Getenv("FB_PAGE_ACCESS_TOKEN")

	// If variables are missing, we log but don't crash since this is a side-effect
	if pageID == "" || accessToken == "" {
		log.Println("Facebook auto-post skipped: Credentials not found in environment variables.")
		return nil
	}

	// Graph API Endpoint for posting to the feed
	endpoint := fmt.Sprintf("https://graph.facebook.com/v21.0/%s/photos", pageID)

	caption := fmt.Sprintf("%s\n\n🛒 Buy now 👉 %s", message, link)

	form := url.Values{}
	form.Add("url", imageUrl)
	form.Add("caption", caption)
	form.Add("access_token", accessToken)
	// payload := map[string]string{
	// 	"message":      message,
	// 	"link":         link,
	// 	"access_token": accessToken,
	// }

	// jsonData, err := json.Marshal(payload)
	// if err != nil {
	// 	return fmt.Errorf("failed to marshal facebook payload: %w", err)
	// }

	resp, err := http.PostForm(endpoint, form)
	if err != nil {
		return fmt.Errorf("failed to send request to facebook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Facebook API error (Status %d): %s", resp.StatusCode, string(body))
		return fmt.Errorf("facebook api responded with status %d", resp.StatusCode)
	}

	log.Println("Successfully posted to Facebook Page.")
	return nil
}

// func Todo(db *gorm.DB) echo.HandlerFunc  {
// 	return func(c echo.Context) error {
// 		var todo models.Todo

// 		if err := c.Bind(&todo); err != nil {
// 			return utils.ResponseError(c, http.StatusBadRequest, "Invalid todo data", err)
// 		}

// 		if err := db.Create(&todo).Error; err != nil {
// 			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to create todo", err)
// 		}

// 		return utils.ResponseSuccess(c, http.StatusOK, "Todo created successfully", todo)

// 	}

// }
