package models

import (
	"time"
)

// PlanDetails represents the structure of our plan information
type PlanDetails struct {
	LemonSqueezyID string
	Name           string
	SubscriptionType string // "monthly" or "yearly"
	Charge         float64
}

// Plans holds all our plan details
var Plans = []PlanDetails{
	{LemonSqueezyID: "336427", Name: "Basic Monthly", SubscriptionType: "monthly", Charge: 10.00},
	{LemonSqueezyID: "336436", Name: "Basic Yearly", SubscriptionType: "yearly", Charge: 102.00},
	{LemonSqueezyID: "336421", Name: "Standard Monthly", SubscriptionType: "monthly", Charge: 19.00},
	{LemonSqueezyID: "336437", Name: "Standard Yearly", SubscriptionType: "yearly", Charge: 193.80},
	{LemonSqueezyID: "336428", Name: "Pro Monthly", SubscriptionType: "monthly", Charge: 39.00},
	{LemonSqueezyID: "336438", Name: "Pro Yearly", SubscriptionType: "yearly", Charge: 397.80},
	{LemonSqueezyID: "336432", Name: "Premium Monthly", SubscriptionType: "monthly", Charge: 69.00},
	{LemonSqueezyID: "336439", Name: "Premium Yearly", SubscriptionType: "yearly", Charge: 703.80},
}

type Subscription struct {
	Base
	UserID            string    `json:"user_id" gorm:"not null"`
	User              User      `json:"user" gorm:"foreignKey:UserID"`
	LemonSqueezyID    string    `json:"lemon_squeezy_id" gorm:"unique;not null"`
	Status            string    `json:"status" gorm:"not null"`
	PlanName          string    `json:"plan_name" gorm:"not null"`
	PlanSubscriptionType string `json:"plan_subscription_type" gorm:"not null"`
	PlanCharge        float64   `json:"plan_charge" gorm:"not null"`
	CurrentPeriodEnd  time.Time `json:"current_period_end"`
	CancelAtPeriodEnd bool      `json:"cancel_at_period_end"`
	Invoices          []Invoice `json:"invoices" gorm:"foreignKey:SubscriptionID"`
}

type Invoice struct {
	Base
	SubscriptionID    uint      `json:"subscription_id" gorm:"not null"`
	LemonSqueezyID    string    `json:"lemon_squeezy_id" gorm:"unique;not null"`
	Amount            float64   `json:"amount" gorm:"not null"`
	Currency          string    `json:"currency" gorm:"not null"`
	Status            string    `json:"status" gorm:"not null"`
	PaidAt            time.Time `json:"paid_at"`
	RefundedAt        *time.Time `json:"refunded_at"`
	DownloadURL       string    `json:"download_url"`
}

// CheckoutSession represents a LemonSqueezy checkout session
type CheckoutSession struct {
	Base
	UserID         string      `json:"user_id" gorm:"not null"`
	User           User `json:"user" gorm:"foreignKey:UserID"`
	LemonSqueezyID string    `json:"lemon_squeezy_id" gorm:"unique;not null"`
	URL            string    `json:"url" gorm:"not null"`
	Status         string    `json:"status" gorm:"not null"`
	ExpiresAt      time.Time `json:"expires_at"`
}


// GetPlanByLemonSqueezyID retrieves plan details by LemonSqueezy ID
func GetPlanByLemonSqueezyID(id string) *PlanDetails {
	for _, plan := range Plans {
		if plan.LemonSqueezyID == id {
			return &plan
		}
	}
	return nil
}

// GetAllPlans returns all available plans
func GetAllPlans() []PlanDetails {
	return Plans
}