package utils

import (
	"api/emails"
	"api/models"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

func StartJobs(db *gorm.DB) {
	go func() {
		// Run every 24 hours
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			<-ticker.C
			CheckAndSendBulkEmails(db)
		}
	}()
}

func CheckAndSendBulkEmails(db *gorm.DB) {
	var unnotifiedProducts []models.Products
	if err := db.Where("is_notified = ?", false).Order("created_at desc").Find(&unnotifiedProducts).Error; err != nil {
		log.Println("Error fetching unnotified products:", err)
		return
	}

	if len(unnotifiedProducts) >= 5 {

		var productNames []string
		count := 0
		for _, p := range unnotifiedProducts {
			if count < 3 {
				productNames = append(productNames, p.Name)
				count++
			}
		}

		// Fetch all active users' emails
		var users []models.User
		if err := db.Where("status = ?", "ACTIVE").Select("email").Find(&users).Error; err != nil {
			log.Println("Error fetching active users:", err)
			return
		}

		var emailsList []string
		for _, u := range users {
			emailsList = append(emailsList, u.Email)
		}

		if len(emailsList) > 0 {
			err := emails.SendNewProductsBulkMail(emailsList, productNames)
			if err != nil {
				log.Println("Error sending bulk email:", err)
				return
			}
			fmt.Printf("Bulk email sent to %d users for %d new products\n", len(emailsList), len(unnotifiedProducts))

			for _, p := range unnotifiedProducts {
				db.Model(&p).Update("is_notified", true)
			}
		}
	}
}
