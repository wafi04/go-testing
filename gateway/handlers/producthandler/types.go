package producthandler

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/exp/rand"
)


type ProductRequest struct {
	Id            string      `json:"id"`
	Name          string       `json:"name"`
	SubTitle      *string       `json:"sub_title"`
	Description   string           `json:"description"`
	Sku           string            `json:"sku"` 
	Price         float64             `json:"price"`
	CategoryId    string                `json:"category_id"`
}



func GenerateSku(name string) string {
	// Initialize random seed

	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	cleanName := reg.ReplaceAllString(name, "")
	cleanName = strings.ToUpper(cleanName)

	// Get the first 3 letters, pad with X if name is too short
	namePrefix := cleanName
	if len(namePrefix) > 3 {
		namePrefix = namePrefix[:3]
	} else {
		for len(namePrefix) < 3 {
			namePrefix += "X"
		}
	}

	// Get current year
	year := time.Now().Year()

	// Generate random 4-digit number
	randomNum := rand.Intn(9000) + 1000 // Ensures 4 digits (1000-9999)

	// Combine all parts to create the SKU
	sku := fmt.Sprintf("%s-%d-%04d", namePrefix, year, randomNum)

	return sku
}

func IsSkuValid(sku string) bool {
	pattern := regexp.MustCompile(`^[A-Z]{3}-\d{4}-\d{4}$`)
	return pattern.MatchString(sku)
}