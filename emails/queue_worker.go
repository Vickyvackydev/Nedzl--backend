package emails

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"api/models"
	"gorm.io/gorm"
)

var (
	queueWorkerRunning bool
	queueWorkerMu      sync.Mutex
)

func StartBulkEmailQueueWorker(db *gorm.DB) {
	queueWorkerMu.Lock()
	if queueWorkerRunning {
		queueWorkerMu.Unlock()
		return
	}
	queueWorkerRunning = true
	queueWorkerMu.Unlock()

	fmt.Println("🚀 Starting Bulk Email Queue Worker...")

	// Process queue once immediately on startup
	go processQueue(db)

	// Run queue worker every 2 minutes
	ticker := time.NewTicker(2 * time.Minute)
	go func() {
		for range ticker.C {
			processQueue(db)
		}
	}()
}

func processQueue(db *gorm.DB) {
	if db == nil {
		return
	}

	maxDailyStr := os.Getenv("MAX_DAILY_BULK_EMAILS")
	maxDaily := 250 // Default daily bulk cap (leaving 50 reserved for transactional emails on 300 free plan)
	if maxDailyStr != "" {
		if parsed, err := strconv.Atoi(maxDailyStr); err == nil {
			maxDaily = parsed
		}
	}

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var sentToday int64
	db.Model(&models.BulkEmailQueueItem{}).
		Where("status = ? AND sent_at >= ?", "SENT", startOfDay).
		Count(&sentToday)

	if maxDaily > 0 && sentToday >= int64(maxDaily) {
		fmt.Printf("⏸️ Bulk Email Queue: Daily quota reached today (%d/%d sent). Pending emails will resume tomorrow at 00:00.\n", sentToday, maxDaily)
		return
	}

	allowance := 500
	if maxDaily > 0 {
		allowance = maxDaily - int(sentToday)
		if allowance <= 0 {
			return
		}
	}

	var pendingItems []models.BulkEmailQueueItem
	err := db.Where("status = ?", "PENDING").
		Order("id ASC").
		Limit(allowance).
		Find(&pendingItems).Error

	if err != nil || len(pendingItems) == 0 {
		return
	}

	fmt.Printf("📬 Bulk Email Queue: Processing %d pending email(s)... (Sent today: %d/%d)\n", len(pendingItems), sentToday, maxDaily)

	for _, item := range pendingItems {
		sendErr := sendSingleMail(item.RecipientEmail, item.RecipientName, item.Subject, item.HTMLContent)
		if sendErr == nil {
			sentAt := time.Now()
			item.Status = "SENT"
			item.SentAt = &sentAt
			item.ErrorMessage = ""
			db.Save(&item)
			fmt.Printf("✅ Bulk email sent to %s (%s)\n", item.RecipientEmail, item.Subject)
		} else {
			fmt.Printf("❌ Failed to send bulk email to %s: %v\n", item.RecipientEmail, sendErr)
			errText := sendErr.Error()
			if strings.Contains(strings.ToLower(errText), "quota") || strings.Contains(strings.ToLower(errText), "rate limit") {
				// Daily quota hit on provider side - stop processing for today
				item.ErrorMessage = errText
				db.Save(&item)
				fmt.Println("⚠️ Provider daily rate limit hit. Pausing queue worker until next cycle/day.")
				break
			} else {
				item.Status = "FAILED"
				item.ErrorMessage = errText
				db.Save(&item)
			}
		}

		time.Sleep(200 * time.Millisecond) // Gentle rate limiting between email dispatches
	}
}

func TriggerQueueProcessing(db *gorm.DB) {
	go processQueue(db)
}
