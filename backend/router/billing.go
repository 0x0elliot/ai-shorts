package router

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	auth "go-authentication-boilerplate/auth"
	"go-authentication-boilerplate/models"
	"go-authentication-boilerplate/util"
)

func SetupBillingRoutes() {
	BILLING.Post("/lemon", HandleLemonSqueezyWebhook)

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

func HandleLemonSqueezyWebhook(c *fiber.Ctx) error {
	webhookSecret := os.Getenv("LEMONSQUEEZY_WEBHOOK_SECRET")
	signature := c.Get("X-Signature")

	if !util.VerifyWebhookSignature(c.Body(), signature, webhookSecret) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": true, "message": "Invalid signature"})
	}

	// print the webhook payload
	log.Printf("[INFO] Webhook payload: %s", c.Body())

	var webhook util.LemonSqueezyWebhook
	if err := json.Unmarshal(c.Body(), &webhook); err != nil {
		log.Printf("[ERROR] Failed to unmarshal webhook payload: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": true, "message": "Invalid payload"})
	}

	switch webhook.Meta.EventName {
	case "subscription_created":
		return handleSubscriptionCreated(c, &webhook)
	case "subscription_cancelled":
		return handleSubscriptionCancelled(c, &webhook)
	default:
		log.Printf("[INFO] Unhandled webhook event: %s", webhook.Meta.EventName)
		return c.SendStatus(fiber.StatusOK)
	}
}

func handleSubscriptionCreated(c *fiber.Ctx, webhook *util.LemonSqueezyWebhook) error {
	userID := webhook.Meta.CustomData.UserID

	user, err := util.GetUserById(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to find user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "message": "Failed to process subscription"})
	}

	log.Printf("[INFO] Product ID: %s", webhook.Data.Attributes.ProductId)

	// convert productID int to string
	// productID := string(webhook.Data.Attributes.ProductId)
	productID := fmt.Sprintf("%d", webhook.Data.Attributes.ProductId)

	// variantID := string(webhook.Data.Attributes.VariantId)

	// second word of webhook.Data.Attributes.ProductName is subscription type
	subscriptionTypeUpper := strings.Split(webhook.Data.Attributes.ProductName, " ")[1]
	subscriptionType := strings.ToLower(subscriptionTypeUpper)

	// get price 
	productPlan := models.GetPlanByLemonSqueezyID(productID)
	if productPlan == nil {
		log.Printf("[ERROR] Failed to find product: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "message": "Failed to process subscription"})
	}

	subscription := &models.Subscription{
		UserID:               user.ID,
		LemonSqueezyID:       productID,
		Status:               webhook.Data.Attributes.Status,
		PlanName:             webhook.Data.Attributes.ProductName,
		PlanSubscriptionType: subscriptionType,
		PlanCharge:           productPlan.Charge,
		CurrentPeriodEnd:     webhook.Data.Attributes.RenewsAt,
	}

	subscription.ID = webhook.Data.ID
	subscription.CreatedAt = time.Now().UTC().Format("2006-01-02T15:04:05.999Z07:00")

	if _, err := util.SetSubscription(subscription); err != nil {
		log.Printf("[ERROR] Failed to create subscription: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "message": "Failed to process subscription"})
	}

	return c.SendStatus(fiber.StatusOK)
}

func handleSubscriptionCancelled(c *fiber.Ctx, webhook *util.LemonSqueezyWebhook) error {
	subscription, err := util.GetSubscriptionByLemonSqueezyID(webhook.Data.ID)
	if err != nil {
		log.Printf("[ERROR] Failed to find subscription: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "message": "Failed to process subscription cancellation"})
	}

	subscription.Status = "cancelled"
	subscription.CancelAtPeriodEnd = true

	if _, err := util.SetSubscription(subscription); err != nil {
		log.Printf("[ERROR] Failed to update subscription: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "message": "Failed to process subscription cancellation"})
	}

	return c.SendStatus(fiber.StatusOK)
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
	variantID, err := util.GetLemonSqueezyVariants(plan.LemonSqueezyID)
	if err != nil {
		log.Printf("[ERROR] Error getting variants for product %s: %v\n", plan.LemonSqueezyID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "message": "Failed to get variants"})
	}

	checkoutSession, err := util.CreateLemonSqueezyCheckout(user.Email, variantID[0], userID)
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