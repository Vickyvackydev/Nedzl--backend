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
		fmt.Println("WhatsApp credentials not configured. Skipping WhatsApp notification.")
		return nil // Skip sending if credentials are not configured, so it doesn't break the app
	}

	formattedPhone := FormatPhoneNumber(toPhone)
	if formattedPhone == "" {
		return fmt.Errorf("invalid phone number")
	}

	url := fmt.Sprintf("https://graph.facebook.com/v19.0/%s/messages", phoneNumberID)

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                formattedPhone,
		"type":              "template",
		"template": map[string]interface{}{
			"name": templateName,
			"language": map[string]interface{}{
				"code": "en",
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
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("whatsapp API error: %d, %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// SendVendorFoodOrderWhatsApp sends an order alert to the food vendor
// Expected Template Name: vendor_order_alert
// Expected Variables: {{1}} Order Number, {{2}} Item Name, {{3}} Customer Name, {{4}} Customer Phone, {{5}} Delivery Address, {{6}} Total Amount, {{7}} Link
func SendVendorFoodOrderWhatsApp(vendorPhone, orderNumber, productName, customerName, customerPhone string, totalAmount float64, deliveryAddress string) error {
	if vendorPhone == "" {
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
// Expected Template Name: artisan_booking_alert
// Expected Variables: {{1}} Booking Number, {{2}} Service Type, {{3}} Customer Name, {{4}} Customer Phone, {{5}} Address, {{6}} Appointment Date, {{7}} Link
func SendServiceBookingWhatsApp(artisanPhone, bookingNumber, serviceType, customerName, customerPhone, address, appointmentDate string) error {
	if artisanPhone == "" {
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
