package auth

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"log"

	shopify "github.com/bold-commerce/go-shopify/v4"
)

var ShopifyAPIKey string = os.Getenv("ACIDRAIN_SHOPIFY_CLIENT_ID")
var ShopifyAPISecret string = os.Getenv("ACIDRAIN_SHOPIFY_CLIENT_SECRET")
var ShopifyScope string = "read_products,write_products,read_customers,read_orders"
var ShopifyRedirectURI string = "http://localhost:3000/shopify/callback"

func GenerateAuthURL(shopName string) string {
	log.Printf("[DEBUG] Crdentials: %s, %s, %s, %s", ShopifyAPIKey, ShopifyAPISecret, ShopifyScope, ShopifyRedirectURI)

	params := url.Values{}
	params.Add("client_id", ShopifyAPIKey)

	params.Add("response_type", "code")

	params.Add("scope", ShopifyScope)
	params.Add("redirect_uri", ShopifyRedirectURI)
	params.Add("state", "nonce") // Add state to prevent CSRF attacks
	params.Add("grant_options[]", "per-user") // Optional: for offline access

	return fmt.Sprintf("https://%s.myshopify.com/admin/oauth/authorize?%s", shopName, params.Encode())
}

func ExchangeToken(ctx context.Context,shopName, code string) (string, error) {
	app := shopify.App{
		ApiKey:      ShopifyAPIKey,
		ApiSecret:   ShopifyAPISecret,
		RedirectUrl: ShopifyRedirectURI,
	}

	token, erauth := app.GetAccessToken(ctx, shopName, code)
	if erauth != nil {
		log.Printf("[ERROR] Couldn't get access token: %v", erauth)
		return "", erauth
	}

	return token, nil
}