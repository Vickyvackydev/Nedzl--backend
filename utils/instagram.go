package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"log"
	"net/http"
	"net/url"
	"os"
)

func PostToInstagram(message string, imageUrl string) error {
	IG_PAGE_ID := os.Getenv("IG_PAGE_ID")
	accessToken := os.Getenv("FB_PAGE_ACCESS_TOKEN")

	if IG_PAGE_ID == "" || accessToken == "" {
		log.Println("Instagram auto-post skipped: Credentials not found in environment variables.")
		return nil
	}

	containerID, err := createMediaContainer(IG_PAGE_ID, accessToken, imageUrl, message)

	if err != nil {
		return fmt.Errorf("failed to create media container: %w", err)
	}

	// Wait for the media container to be ready
	if err := waitForMediaReady(containerID, accessToken); err != nil {
		return fmt.Errorf("media not ready: %w", err)
	}

	if err := publishMediaContainer(IG_PAGE_ID, accessToken, containerID); err != nil {
		return fmt.Errorf("failed to publish media container: %w", err)
	}

	log.Println("Successfully posted to Instagram Page.")

	return nil
}

func createMediaContainer(pageID, token, imageUrl, caption string) (string, error) {

	endpoint := fmt.Sprintf("https://graph.facebook.com/v21.0/%s/media", pageID)

	form := url.Values{}
	form.Add("image_url", imageUrl)
	form.Add("caption", caption)
	form.Add("access_token", token)

	resp, err := http.PostForm(endpoint, form)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Instagram API error (Status %d): %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("instagram api responded with status %d", resp.StatusCode)
	}

	// we extract the id from the response body
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	return result["id"], nil

}

func waitForMediaReady(containerID, token string) error {
	endpoint := fmt.Sprintf("https://graph.facebook.com/v21.0/%s", containerID)
	params := url.Values{}
	params.Add("fields", "status_code")
	params.Add("access_token", token)

	// Poll up to 5 times with a delay
	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second) // Wait 2s between checks

		resp, err := http.Get(endpoint + "?" + params.Encode())
		if err != nil {
			log.Printf("Failed to check media status: %v", err)
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			continue
		}

		status, ok := result["status_code"].(string)
		if !ok {
			// If status_code isn't returned, maybe it's implicitly ready or error?
			// Let's assume waiting longer is better or proceed.
			continue
		}

		if status == "FINISHED" {
			return nil
		}
		if status == "ERROR" {
			return fmt.Errorf("media container status is ERROR")
		}
		// If "IN_PROGRESS", continue loop
		log.Printf("Media status: %s. Waiting...", status)
	}
	return fmt.Errorf("media validation timed out")
}

func publishMediaContainer(mediaID, token string, containerID string) error {
	endpoint := fmt.Sprintf("https://graph.facebook.com/v21.0/%s/media_publish", mediaID)

	form := url.Values{}
	form.Add("creation_id", containerID)
	form.Add("access_token", token)

	resp, err := http.PostForm(endpoint, form)

	if err != nil {
		return fmt.Errorf("failed to send request to facebook: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Instagram API error (Status %d): %s", resp.StatusCode, string(body))
		return fmt.Errorf("instagram api responded with status %d", resp.StatusCode)
	}

	return nil
}
