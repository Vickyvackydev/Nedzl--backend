package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"api/db"
	"api/models"
	"api/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type SendCommunityMessageRequest struct {
	SenderName  string     `json:"sender_name"`
	SenderEmail string     `json:"sender_email"`
	Message     string     `json:"message"`
	ReplyToID   *uuid.UUID `json:"reply_to_id"`
}

type EmojiReactionItem struct {
	Emoji string   `json:"emoji"`
	Count int      `json:"count"`
	Users []string `json:"users"`
}

// Regex to detect links, URLs, and web domains
var linkRegex = regexp.MustCompile(`(?i)(https?://|www\.|[a-zA-Z0-9-]+\.(com|ng|org|net|io|edu|gov|co|me|xyz|app|dev))`)

// GetCommunityMessages retrieves the latest community messages
func GetCommunityMessages(c echo.Context) error {
	var messages []models.CommunityMessage
	if err := db.DB.Preload("ReplyTo").Order("created_at ASC").Limit(200).Find(&messages).Error; err != nil {
		return utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch community messages", err)
	}

	return utils.ResponseSucess(c, http.StatusOK, "Community messages retrieved successfully", messages)
}

// SendCommunityMessage validates and stores a new message in Nedzl Community
func SendCommunityMessage(c echo.Context) error {
	var req SendCommunityMessageRequest
	if err := c.Bind(&req); err != nil {
		return utils.ResponseError(c, http.StatusBadRequest, "Invalid request payload", err)
	}

	req.SenderName = strings.TrimSpace(req.SenderName)
	req.Message = strings.TrimSpace(req.Message)

	if req.SenderName == "" {
		req.SenderName = "Nedzl Community Member"
	}

	if req.Message == "" {
		return utils.ResponseError(c, http.StatusBadRequest, "Message body cannot be empty", nil)
	}

	// Link restriction validation: Strictly block any links or URLs
	if linkRegex.MatchString(req.Message) {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  false,
			"message": "Posting links or web URLs is strictly prohibited in Nedzl Community.",
		})
	}

	var userIDPtr *uuid.UUID
	if val := c.Get("user_id"); val != nil {
		if uid, ok := val.(uuid.UUID); ok {
			userIDPtr = &uid
		}
	}

	msg := models.CommunityMessage{
		SenderName:  req.SenderName,
		SenderEmail: req.SenderEmail,
		UserID:      userIDPtr,
		Message:     req.Message,
		ReplyToID:   req.ReplyToID,
	}

	if err := db.DB.Create(&msg).Error; err != nil {
		return utils.ResponseError(c, http.StatusInternalServerError, "Failed to post message", err)
	}

	// Preload ReplyTo if it exists
	if msg.ReplyToID != nil {
		db.DB.Preload("ReplyTo").First(&msg, msg.ID)
	}

	return utils.ResponseSucess(c, http.StatusCreated, "Message posted successfully", msg)
}

type ReactRequest struct {
	Emoji      string `json:"emoji"`
	SenderName string `json:"sender_name"`
}

// ReactToCommunityMessage toggles an emoji reaction on a message
func ReactToCommunityMessage(c echo.Context) error {
	msgIDStr := c.Param("id")
	msgID, err := uuid.Parse(msgIDStr)
	if err != nil {
		return utils.ResponseError(c, http.StatusBadRequest, "Invalid message ID", err)
	}

	var req ReactRequest
	if err := c.Bind(&req); err != nil || req.Emoji == "" {
		return utils.ResponseError(c, http.StatusBadRequest, "Emoji is required", err)
	}

	if req.SenderName == "" {
		req.SenderName = "Anonymous"
	}

	var msg models.CommunityMessage
	if err := db.DB.First(&msg, "id = ?", msgID).Error; err != nil {
		return utils.ResponseError(c, http.StatusNotFound, "Message not found", err)
	}

	var reactions []EmojiReactionItem
	if len(msg.Reactions) > 0 {
		_ = json.Unmarshal(msg.Reactions, &reactions)
	}

	// Toggle user reaction for the given emoji
	foundEmojiIndex := -1
	for i, r := range reactions {
		if r.Emoji == req.Emoji {
			foundEmojiIndex = i
			break
		}
	}

	if foundEmojiIndex >= 0 {
		// Emoji exists in list
		userIndex := -1
		for uIdx, uName := range reactions[foundEmojiIndex].Users {
			if uName == req.SenderName {
				userIndex = uIdx
				break
			}
		}

		if userIndex >= 0 {
			// Remove reaction if user already reacted
			reactions[foundEmojiIndex].Users = append(reactions[foundEmojiIndex].Users[:userIndex], reactions[foundEmojiIndex].Users[userIndex+1:]...)
			reactions[foundEmojiIndex].Count--
		} else {
			// Add user reaction
			reactions[foundEmojiIndex].Users = append(reactions[foundEmojiIndex].Users, req.SenderName)
			reactions[foundEmojiIndex].Count++
		}

		// Remove reaction entry if count reaches 0
		if reactions[foundEmojiIndex].Count <= 0 {
			reactions = append(reactions[:foundEmojiIndex], reactions[foundEmojiIndex+1:]...)
		}
	} else {
		// New emoji reaction
		reactions = append(reactions, EmojiReactionItem{
			Emoji: req.Emoji,
			Count: 1,
			Users: []string{req.SenderName},
		})
	}

	jsonBytes, _ := json.Marshal(reactions)
	msg.Reactions = jsonBytes

	if err := db.DB.Model(&msg).Update("reactions", jsonBytes).Error; err != nil {
		return utils.ResponseError(c, http.StatusInternalServerError, "Failed to update reaction", err)
	}

	return utils.ResponseSucess(c, http.StatusOK, "Reaction updated", msg)
}
