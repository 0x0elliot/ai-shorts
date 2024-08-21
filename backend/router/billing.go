package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"

	auth "go-authentication-boilerplate/auth"
	"go-authentication-boilerplate/models"
	"go-authentication-boilerplate/util"
)

func SetupBillingRoutes() {
	privBilling := BILLING.Group("/private")
	privBilling.Use(auth.SecureAuth())

	privBilling.Post("/create-checkout", HandleCreateCheckout)
	privBilling.Get("/plans", HandleGetPlans)
	privBilling.Get("/current-plan", HandleGetCurrentPlan)
}

type CheckoutInput struct {
	PlanID string `json:"plan_id"`
	Email  string `json:"email"`
}

func HandleCreateCheckout(c *fiber.Ctx) error {
	input := new(CheckoutInput)

	userId := c.Locals("id").(string)

	user, err := util.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Error getting user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "message": "Failed to get user"})
	}

	if err := c.BodyParser(input); err != nil {
		log.Printf("[ERROR] Couldn't parse the input: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": true, "message": "Please review your input"})
	}

	userID := c.Locals("id").(string)

	plan := models.GetPlanByLemonSqueezyID(input.PlanID)
	if plan == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": true, "message": "Plan not found"})
		
	}
	variantID, err := getLemonSqueezyVariants(plan.LemonSqueezyID)
	if err != nil {
		log.Printf("[ERROR] Error getting variants for product %s: %v\n", plan.LemonSqueezyID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "message": "Failed to get variants"})
	}

	checkoutSession, err := createLemonSqueezyCheckout(user.Email, variantID[0], userID)
	if err != nil {
		log.Printf("[ERROR] Failed to create LemonSqueezy checkout: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "message": "Failed to create checkout session"})
	}

	dbCheckoutSession := models.CheckoutSession{
		UserID:         userID,
		LemonSqueezyID: checkoutSession.Data.ID,
		URL:            checkoutSession.Data.Attributes.URL,
		Status:         "pending",
		ExpiresAt:      checkoutSession.Data.Attributes.ExpiresAt,
	}

	if _, err := util.SetCheckoutSession(&dbCheckoutSession); err != nil {
		log.Printf("[ERROR] Failed to save checkout session: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "message": "Failed to save checkout session"})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "Checkout session created successfully",
		"data": fiber.Map{
			"checkout_url": checkoutSession.Data.Attributes.URL,
			"expires_at":   checkoutSession.Data.Attributes.ExpiresAt,
		},
	})
}

func HandleGetCurrentPlan(c *fiber.Ctx) error {
	userID := c.Locals("id").(string)
	subscription, err := util.GetActiveSubscriptionByUserID(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get active subscription: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "message": "Failed to get active subscription"})
	}

	// If no active subscription found
	if subscription == nil {
		return c.JSON(fiber.Map{
			"error":        false,
			"subscription": nil,
		})
	}

	// Prepare invoice information
	invoices := make([]fiber.Map, len(subscription.Invoices))
	for i, invoice := range subscription.Invoices {
		invoices[i] = fiber.Map{
			"id":           invoice.ID,
			"amount":       invoice.Amount,
			"currency":     invoice.Currency,
			"status":       invoice.Status,
			"paid_at":      invoice.PaidAt,
			"refunded_at":  invoice.RefundedAt,
			"download_url": invoice.DownloadURL,
		}
	}

	return c.JSON(fiber.Map{
		"error": false,
		"subscription": fiber.Map{
			"id":                     subscription.ID,
			"lemon_squeezy_id":       subscription.LemonSqueezyID,
			"status":                 subscription.Status,
			"plan_name":              subscription.PlanName,
			"plan_subscription_type": subscription.PlanSubscriptionType,
			"plan_charge":            subscription.PlanCharge,
			"current_period_end":     subscription.CurrentPeriodEnd,
			"cancel_at_period_end":   subscription.CancelAtPeriodEnd,
			"invoices":               invoices,
		},
	})
}

func HandleGetPlans(c *fiber.Ctx) error {
	allPlans := models.GetAllPlans()
	return c.JSON(fiber.Map{
		"error": false,
		"plans": allPlans,
	})
}

type LemonSqueezyCheckoutResponse struct {
	Data struct {
		ID         string `json:"id"`
		Attributes struct {
			StoreID       int       `json:"store_id"`
			CustomerEmail string    `json:"customer_email"`
			Currency      string    `json:"currency"`
			Total         int       `json:"total"`
			ExpiresAt     time.Time `json:"expires_at"`
			URL           string    `json:"url"`
		} `json:"attributes"`
	} `json:"data"`
}

func getLemonSqueezyVariants(productID string) ([]string, error) {
	url := fmt.Sprintf("https://api.lemonsqueezy.com/v1/variants?filter[product_id]=%s", productID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("ACIDRAIN_LEMONSQUEEZY_KEYS")))
	req.Header.Set("Accept", "application/vnd.api+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	variantIDs := make([]string, len(response.Data))
	for i, variant := range response.Data {
		variantIDs[i] = variant.ID
	}

	return variantIDs, nil
}

func createLemonSqueezyCheckout(email string, planID string, userID string) (*LemonSqueezyCheckoutResponse, error) {

	log.Printf("[INFO] Creating LemonSqueezy checkout for user: %s, plan: %s", userID, planID)

	url := fmt.Sprintf("%s/v1/checkouts", "https://api.lemonsqueezy.com")
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "checkouts",
			"attributes": map[string]interface{}{
				"checkout_data": map[string]interface{}{
					"email": email,
					"custom": map[string]interface{}{
						"user_id": userID,
					},
				},
			},
			"relationships": map[string]interface{}{
				"store": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "stores",
						"id":   "117377",
					},
				},
				"variant": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "variants",
						"id":   planID,
					},
				},
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("ACIDRAIN_LEMONSQUEEZY_KEYS")))
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Accept", "application/vnd.api+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var checkoutResponse LemonSqueezyCheckoutResponse
	if err := json.Unmarshal(body, &checkoutResponse); err != nil {
		return nil, err
	}

	return &checkoutResponse, nil
}
