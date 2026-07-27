package utils

import (
	"api/emails"
	"api/models"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

func StartJobs(db *gorm.DB) {
	go func() {
		log.Println("Jobs: Running initial bulk email check on startup...")
		CheckAndSendBulkEmails(db)

		// Run every 24 hours
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			<-ticker.C
			log.Println("Jobs: Running periodic bulk email check...")
			CheckAndSendBulkEmails(db)
		}
	}()

	go func() {
		log.Println("Jobs: Starting Escrow Auto-Release background worker...")
		AutoReleaseEscrowBookings(db)

		// Check every 15 minutes
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		for {
			<-ticker.C
			AutoReleaseEscrowBookings(db)
		}
	}()
}

func AutoReleaseEscrowBookings(db *gorm.DB) {
	cutoff := time.Now().Add(-24 * time.Hour)
	var pendingBookings []models.ServiceBooking

	err := db.Where("status = ? AND artisan_completed_at <= ?", "ARTISAN_COMPLETED", cutoff).Find(&pendingBookings).Error
	if err != nil {
		log.Println("Jobs: Error fetching pending escrow bookings for auto-release:", err)
		return
	}

	if len(pendingBookings) == 0 {
		return
	}

	now := time.Now()
	for _, b := range pendingBookings {
		b.Status = "COMPLETED"
		b.CompletedAt = &now
		b.PaymentStatus = "RELEASED_TO_ARTISAN"

		if err := db.Save(&b).Error; err != nil {
			log.Printf("Jobs: Error auto-releasing booking #%s: %v\n", b.BookingNumber, err)
		} else {
			log.Printf("Jobs: Auto-completed booking #%s and released 90%% payout to artisan after 24h\n", b.BookingNumber)
		}
	}
}

func CheckAndSendBulkEmails(db *gorm.DB) {
	log.Println("Jobs: Checking database for unnotified products...")
	var unnotifiedProducts []models.Products
	if err := db.Where("is_notified = ?", false).Order("created_at desc").Find(&unnotifiedProducts).Error; err != nil {
		log.Println("Error fetching unnotified products:", err)
		return
	}

	log.Printf("Jobs: Found %d unnotified products in the database", len(unnotifiedProducts))

	if len(unnotifiedProducts) >= 5 {

		var emailProducts []emails.EmailProduct
		count := 0
		for _, p := range unnotifiedProducts {
			if count < 3 {
				var images []string
				imageUrl := ""
				if len(p.ImageUrls) > 0 {
					_ = json.Unmarshal(p.ImageUrls, &images)
				}
				if len(images) > 0 {
					imageUrl = images[0]
				} else {
					imageUrl = "https://nedzl.com/placeholder.png"
				}

				desc := p.Description
				if len(desc) > 120 {
					desc = desc[:117] + "..."
				}

				emailProducts = append(emailProducts, emails.EmailProduct{
					ID:          p.ID.String(),
					Name:        p.Name,
					Price:       p.ProductPrice,
					Description: desc,
					ImageUrl:    imageUrl,
				})
				count++
			}
		}

		// Fetch all active users
		var users []models.User
		if err := db.Where("status = ?", "ACTIVE").Select("email", "user_name").Find(&users).Error; err != nil {
			log.Println("Error fetching active users:", err)
			return
		}

		var recipients []emails.BulkEmailRecipient
		for _, u := range users {
			recipients = append(recipients, emails.BulkEmailRecipient{
				Email:    u.Email,
				UserName: u.UserName,
			})
		}

		if len(recipients) > 0 {
			err := emails.SendNewProductsBulkMail(recipients, emailProducts)
			if err != nil {
				log.Println("Error sending bulk email:", err)
				return
			}
			fmt.Printf("Bulk email sent to %d users for %d new products\n", len(recipients), len(unnotifiedProducts))

			for _, p := range unnotifiedProducts {
				db.Model(&p).Update("is_notified", true)
			}
		}
	} else {
		log.Printf("Jobs: Threshold of 5 unnotified products not met (have %d). Skipping email sending.", len(unnotifiedProducts))
	}
}
