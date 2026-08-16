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
		log.Println("Jobs: Starting Escrow Auto-Release & Payout background worker...")
		AutoReleaseEscrowBookings(db)
		AutoReleaseEscrowFoodOrders(db)
		ProcessPendingUnpaidEscrowTransfers(db)

		// Check every 15 minutes
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		for {
			<-ticker.C
			AutoReleaseEscrowBookings(db)
			AutoReleaseEscrowFoodOrders(db)
			ProcessPendingUnpaidEscrowTransfers(db)
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
			log.Printf("Jobs: Auto-completed booking #%s and released 90%% payout status to artisan after 24h\n", b.BookingNumber)
		}
	}
}

func AutoReleaseEscrowFoodOrders(db *gorm.DB) {
	cutoff := time.Now().Add(-24 * time.Hour)
	var pendingOrders []models.FoodOrder

	err := db.Where("(status = ? OR status = ?) AND vendor_delivered_at <= ?", "DELIVERED", "DELIVERED_BY_VENDOR", cutoff).Find(&pendingOrders).Error
	if err != nil {
		log.Println("Jobs: Error fetching pending food orders for auto-release:", err)
		return
	}

	if len(pendingOrders) == 0 {
		return
	}

	now := time.Now()
	for _, o := range pendingOrders {
		o.Status = "COMPLETED"
		o.PaymentStatus = "RELEASED_TO_VENDOR"
		if o.CustomerConfirmedAt == nil {
			o.CustomerConfirmedAt = &now
		}

		if err := db.Save(&o).Error; err != nil {
			log.Printf("Jobs: Error auto-releasing food order #%s: %v\n", o.OrderNumber, err)
		} else {
			log.Printf("Jobs: Auto-completed food order #%s and released payout status to vendor after 24h\n", o.OrderNumber)
		}
	}
}

// ProcessPendingUnpaidEscrowTransfers scans for completed orders/bookings whose Paystack transfers haven't been initiated yet and executes them automatically
func ProcessPendingUnpaidEscrowTransfers(db *gorm.DB) {
	// 1. Process Food Orders
	var pendingFoodOrders []models.FoodOrder
	if err := db.Where("(status = ? OR payment_status = ?) AND (payout_transfer_ref IS NULL OR payout_transfer_ref = '')", "COMPLETED", "RELEASED_TO_VENDOR").Find(&pendingFoodOrders).Error; err == nil {
		for _, ord := range pendingFoodOrders {
			var vendor models.User
			if err := db.First(&vendor, "id = ?", ord.VendorID).Error; err == nil {
				accNum := vendor.AccountNumber
				bankName := vendor.BankName
				accName := vendor.AccountName

				if (accNum == "" || bankName == "") && len(vendor.BankAccounts) > 0 {
					var accounts []models.BankAccountItem
					if err := json.Unmarshal(vendor.BankAccounts, &accounts); err == nil {
						for _, acc := range accounts {
							if acc.IsDefault || accNum == "" {
								accNum = acc.AccountNumber
								bankName = acc.BankName
								accName = acc.AccountName
								if acc.IsDefault {
									break
								}
							}
						}
					}
				}

				if accNum != "" && bankName != "" {
					if accName == "" {
						accName = vendor.UserName
					}
					bankCode := GetBankCodeByName(bankName)
					transferRef := fmt.Sprintf("TRF-%s-%d", ord.OrderNumber, time.Now().Unix())
					reason := fmt.Sprintf("Nedzl Food Order #%s Payout", ord.OrderNumber)

					code, err := InitiatePaystackTransfer(bankCode, accNum, accName, ord.VendorPayout, transferRef, reason)
					if err != nil {
						log.Printf("Jobs: Automated Paystack Transfer failed for food order #%s: %v\n", ord.OrderNumber, err)
					} else {
						log.Printf("Jobs: Automated Paystack Transfer initiated for food order #%s (%s): ₦%.2f to %s (%s)\n", ord.OrderNumber, code, ord.VendorPayout, accNum, bankName)
						payoutTime := time.Now()
						db.Model(&models.FoodOrder{}).Where("id = ?", ord.ID).Updates(map[string]interface{}{
							"payout_transfer_ref":   transferRef,
							"payout_transferred_at": payoutTime,
						})
					}
				}
			}
		}
	}

	// 2. Process Service Bookings
	var pendingBookings []models.ServiceBooking
	if err := db.Where("(status = ? OR payment_status = ?) AND (payout_transfer_ref IS NULL OR payout_transfer_ref = '')", "COMPLETED", "RELEASED_TO_ARTISAN").Find(&pendingBookings).Error; err == nil {
		for _, bk := range pendingBookings {
			var artisan models.User
			if err := db.First(&artisan, "id = ?", bk.ArtisanID).Error; err == nil {
				accNum := artisan.AccountNumber
				bankName := artisan.BankName
				accName := artisan.AccountName

				if (accNum == "" || bankName == "") && len(artisan.BankAccounts) > 0 {
					var accounts []models.BankAccountItem
					if err := json.Unmarshal(artisan.BankAccounts, &accounts); err == nil {
						for _, acc := range accounts {
							if acc.IsDefault || accNum == "" {
								accNum = acc.AccountNumber
								bankName = acc.BankName
								accName = acc.AccountName
								if acc.IsDefault {
									break
								}
							}
						}
					}
				}

				if accNum != "" && bankName != "" {
					if accName == "" {
						accName = artisan.UserName
					}
					bankCode := GetBankCodeByName(bankName)
					transferRef := fmt.Sprintf("TRF-%s-%d", bk.BookingNumber, time.Now().Unix())
					reason := fmt.Sprintf("Nedzl Service Booking #%s Payout", bk.BookingNumber)

					code, err := InitiatePaystackTransfer(bankCode, accNum, accName, bk.ArtisanPayout, transferRef, reason)
					if err != nil {
						log.Printf("Jobs: Automated Paystack Transfer failed for service booking #%s: %v\n", bk.BookingNumber, err)
					} else {
						log.Printf("Jobs: Automated Paystack Transfer initiated for service booking #%s (%s): ₦%.2f to %s (%s)\n", bk.BookingNumber, code, bk.ArtisanPayout, accNum, bankName)
						payoutTime := time.Now()
						db.Model(&models.ServiceBooking{}).Where("id = ?", bk.ID).Updates(map[string]interface{}{
							"payout_transfer_ref":   transferRef,
							"payout_transferred_at": payoutTime,
						})
					}
				}
			}
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
