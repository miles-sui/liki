// Package product defines the product types and pricing shared between agent and payment layers.
package product

import "github.com/google/uuid"

// Product identifies a paid report product.
type Product string

const ProductNaming Product = "naming"

func (p Product) EmailSubject() string {
	switch p {
	case ProductNaming:
		return "您的起名报告"
	default:
		return "您的命理报告"
	}
}

// Currency identifies a payment currency.
type Currency string

const (
	CNY Currency = "CNY"
	USD Currency = "USD"
)

// NewOrderID generates a new UUID-based order ID.
func NewOrderID() string {
	return uuid.New().String()
}

// NamingAmountCents returns the price in cents for the naming product.
func NamingAmountCents(c Currency) int {
	switch c {
	case USD:
		return 2990
	case CNY:
		return 9900
	default:
		return 0
	}
}
