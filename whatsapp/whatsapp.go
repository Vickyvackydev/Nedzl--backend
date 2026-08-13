package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// FormatPhoneNumber ensures the phone number is correctly formatted for WhatsApp API (E.164 without '+')
func FormatPhoneNumber(phone string) string {
	// Remove all non-numeric characters
	re := regexp.MustCompile(`[^0-9]`)
	cleanPhone := re.ReplaceAllString(phone, "")

	// Handle Nigerian numbers starting with 0 (e.g., 080...) -> 23480...
	if strings.HasPrefix(cleanPhone, "0") && len(cleanPhone) == 11 {
		cleanPhone = "234" + cleanPhone[1:]
	}

	return cleanPhone
}

// sendWhatsAppTemplate is a generic helper to send template messages via Meta Cloud API
func sendWhatsAppTemplate(toPhone string, templateName string, headerTitle string, parameters []map[string]interface{}) error {
	phoneNumberID := os.Getenv("WHATSAPP_PHONE_NUMBER_ID")
	token := os.Getenv("WHATSAPP_TOKEN")

	if phoneNumberID == "" || token == "" {
		fmt.Println("[WhatsApp] Credentials not configured (WHATSAPP_PHONE_NUMBER_ID / WHATSAPP_TOKEN). Skipping notification.")
		return nil
	}

	formattedPhone := FormatPhoneNumber(toPhone)
	if formattedPhone == "" {
		fmt.Printf("[WhatsApp Error] Cannot send notification: phone number '%s' is empty or invalid.\n", toPhone)
		return fmt.Errorf("invalid phone number")
	}

	url := fmt.Sprintf("https://graph.facebook.com/v19.0/%s/messages", phoneNumberID)

	components := []map[string]interface{}{}
	if headerTitle != "" {
		components = append(components, map[string]interface{}{
			"type": "header",
			"parameters": []map[string]interface{}{
				{
					"type": "text",
					"text": headerTitle,
				},
			},
		})
	}
	components = append(components, map[string]interface{}{
		"type":       "body",
		"parameters": parameters,
	})

	// Try default language "en", then fallback to "en_US" if needed
	langCodes := []string{"en", "en_US"}
	var lastErr error

	for _, lang := range langCodes {
		payload := map[string]interface{}{
			"messaging_product": "whatsapp",
			"to":                formattedPhone,
			"type":              "template",
			"template": map[string]interface{}{
				"name": templateName,
				"language": map[string]interface{}{
					"code": lang,
				},
				"components": components,
			},
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			fmt.Printf("[WhatsApp Error] HTTP Request failed for %s: %v\n", formattedPhone, err)
			continue
		}
		defer resp.Body.Close()

		bodyBytes, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			fmt.Printf("[WhatsApp Success] Message sent to %s via template '%s' (lang: %s)\n", formattedPhone, templateName, lang)
			return nil
		}

		lastErr = fmt.Errorf("status %d: %s", resp.StatusCode, string(bodyBytes))
		fmt.Printf("[WhatsApp Warning] Language '%s' failed for template '%s' (%s). Trying next fallback...\n", lang, templateName, string(bodyBytes))
	}

	fmt.Printf("[WhatsApp Error] All language attempts failed for %s: %v\n", formattedPhone, lastErr)
	return lastErr
}

// SendVendorFoodOrderWhatsApp sends an order alert to the food vendor
func SendVendorFoodOrderWhatsApp(vendorPhone, orderNumber, productName, customerName, customerPhone string, totalAmount float64, deliveryAddress string) error {
	if vendorPhone == "" {
		fmt.Println("[WhatsApp Warning] Vendor phone number is empty. Cannot send food order alert.")
		return nil
	}

	amountStr := fmt.Sprintf("N%.2f", totalAmount)
	link := "https://nedzl.com/dashboard?tab=orders"
	title := fmt.Sprintf("Nedzl Meals - Order #%s", orderNumber)

	parameters := []map[string]interface{}{
		{"type": "text", "text": title},
		{"type": "text", "text": productName},
		{"type": "text", "text": customerName},
		{"type": "text", "text": customerPhone},
		{"type": "text", "text": deliveryAddress},
		{"type": "text", "text": amountStr},
		{"type": "text", "text": link},
	}

	return sendWhatsAppTemplate(vendorPhone, "vendor_order_alert", "Nedzl Meals", parameters)
}

// SendServiceBookingWhatsApp sends a booking alert to an artisan
func SendServiceBookingWhatsApp(artisanPhone, bookingNumber, serviceType, customerName, customerPhone, address, appointmentDate string) error {
	if artisanPhone == "" {
		fmt.Println("[WhatsApp Warning] Artisan phone number is empty. Cannot send booking alert.")
		return nil
	}

	link := "https://nedzl.com/dashboard?tab=bookings"
	title := fmt.Sprintf("Nedzl Services - Booking #%s", bookingNumber)

	parameters := []map[string]interface{}{
		{"type": "text", "text": title},
		{"type": "text", "text": serviceType},
		{"type": "text", "text": customerName},
		{"type": "text", "text": customerPhone},
		{"type": "text", "text": address},
		{"type": "text", "text": appointmentDate},
		{"type": "text", "text": link},
	}

	return sendWhatsAppTemplate(artisanPhone, "artisan_booking_alert", "Nedzl Services", parameters)
}
