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
func sendWhatsAppTemplate(toPhone string, templateName string, parameters []map[string]interface{}) error {
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
				"components": []map[string]interface{}{
					{
						"type":       "body",
						"parameters": parameters,
					},
				},
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
	fmt.Printf("[WhatsApp Trigger] SendVendorFoodOrderWhatsApp invoked for order: %s, vendorPhone: '%s'\n", orderNumber, vendorPhone)

	if vendorPhone == "" {
		fmt.Println("[WhatsApp Warning] Vendor phone number is empty. Cannot send food order alert.")
		return nil
	}

	amountStr := fmt.Sprintf("N%.2f", totalAmount)
	link := "https://nedzl.com/dashboard?tab=orders"

	parameters := []map[string]interface{}{
		{"type": "text", "text": orderNumber},
		{"type": "text", "text": productName},
		{"type": "text", "text": customerName},
		{"type": "text", "text": customerPhone},
		{"type": "text", "text": deliveryAddress},
		{"type": "text", "text": amountStr},
		{"type": "text", "text": link},
	}

	return sendWhatsAppTemplate(vendorPhone, "vendor_order_alert", parameters)
}

// SendServiceBookingWhatsApp sends a booking alert to an artisan
func SendServiceBookingWhatsApp(artisanPhone, bookingNumber, serviceType, customerName, customerPhone, address, appointmentDate string) error {
	fmt.Printf("[WhatsApp Trigger] SendServiceBookingWhatsApp invoked for booking: %s, artisanPhone: '%s'\n", bookingNumber, artisanPhone)

	if artisanPhone == "" {
		fmt.Println("[WhatsApp Warning] Artisan phone number is empty. Cannot send booking alert.")
		return nil
	}

	link := "https://nedzl.com/dashboard?tab=bookings"

	parameters := []map[string]interface{}{
		{"type": "text", "text": bookingNumber},
		{"type": "text", "text": serviceType},
		{"type": "text", "text": customerName},
		{"type": "text", "text": customerPhone},
		{"type": "text", "text": address},
		{"type": "text", "text": appointmentDate},
		{"type": "text", "text": link},
	}

	return sendWhatsAppTemplate(artisanPhone, "artisan_booking_alert", parameters)
}

// SendCustomerFoodDeliveredWhatsApp sends a notification to the customer when a vendor marks their food order as delivered
func SendCustomerFoodDeliveredWhatsApp(customerPhone, customerName, orderNumber, productName, vendorName string) error {
	fmt.Printf("[WhatsApp Trigger] SendCustomerFoodDeliveredWhatsApp invoked for order: %s, customerPhone: '%s'\n", orderNumber, customerPhone)

	if customerPhone == "" {
		fmt.Println("[WhatsApp Warning] Customer phone number is empty. Cannot send food delivery alert.")
		return nil
	}

	link := "https://nedzl.com/dashboard?tab=my_orders"

	parameters := []map[string]interface{}{
		{"type": "text", "text": customerName},
		{"type": "text", "text": orderNumber},
		{"type": "text", "text": productName},
		{"type": "text", "text": vendorName},
		{"type": "text", "text": link},
	}

	return sendWhatsAppTemplate(customerPhone, "customer_food_delivered", parameters)
}

// SendCustomerServiceCompletedWhatsApp sends a notification to the customer when an artisan marks their service as completed
func SendCustomerServiceCompletedWhatsApp(customerPhone, customerName, bookingNumber, serviceName, artisanName string) error {
	fmt.Printf("[WhatsApp Trigger] SendCustomerServiceCompletedWhatsApp invoked for booking: %s, customerPhone: '%s'\n", bookingNumber, customerPhone)

	if customerPhone == "" {
		fmt.Println("[WhatsApp Warning] Customer phone number is empty. Cannot send service completion alert.")
		return nil
	}

	link := "https://nedzl.com/dashboard?tab=service_bookings"

	parameters := []map[string]interface{}{
		{"type": "text", "text": customerName},
		{"type": "text", "text": bookingNumber},
		{"type": "text", "text": serviceName},
		{"type": "text", "text": artisanName},
		{"type": "text", "text": link},
	}

	return sendWhatsAppTemplate(customerPhone, "customer_service_completed", parameters)
}

