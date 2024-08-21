package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"os"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	models "go-authentication-boilerplate/models"
)


type LemonSqueezyCheckoutResponse struct {
	Data struct {
		ID         string `json:"id"`
		Attributes struct {
			StoreID       int       `json:"store_id"`
			UserEmail	 string    `json:"user_email"`
			Currency      string    `json:"currency"`
			Total         int       `json:"total"`
			ExpiresAt     time.Time `json:"expires_at"`
			URL           string    `json:"url"`
		} `json:"attributes"`
	} `json:"data"`
}

type LemonSqueezyWebhook struct {
	Meta struct {
		EventName string `json:"event_name"`
		CustomData 		 struct {
			UserID string `json:"user_id"`
		} `json:"custom_data"`
	} `json:"meta"`
	Data struct {
		Relationships struct {
			SubscriptionInvoice struct {
				Links struct {
					Related string `json:"related"`
				} `json:"links"`
			} `json:"subscription_invoice"`
		} `json:"relationships"`

		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			ProductName    string    `json:"product_name"`
			RenewsAt       time.Time `json:"renews_at"`
			Status           string    `json:"status"`
			UserEmail    string    `json:"user_email"`
			ProductId        int       `json:"product_id"`
			VariantId        int       `json:"variant_id"`
			Total            float64       `json:"total"`
			CurrentPeriodEnd time.Time `json:"current_period_end"`
		} `json:"attributes"`
	} `json:"data"`
}

func VerifyWebhookSignature(payload []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func GetPlanNameFromProductID(productID string) string {
	for _, plan := range models.Plans {
		planID := plan.LemonSqueezyID

		if planID == productID {
			return plan.Name
		}
	}
	return "Unknown Plan"
}

func GetPlanTypeFromVariantID(variantID string) string {
	for _, plan := range models.Plans {
		planID := plan.LemonSqueezyID
		
		plansVariants, err := GetLemonSqueezyVariants(planID)
		if err != nil {
			continue
		}
	
		if plansVariants[0] == variantID {
			return plan.SubscriptionType
		}
	}

	return ""
}

func GetLemonSqueezyVariants(productID string) ([]string, error) {
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

func CreateLemonSqueezyCheckout(email string, planID string, userID string) (*LemonSqueezyCheckoutResponse, error) {

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
