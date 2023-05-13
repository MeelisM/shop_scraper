package helpers

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type Product struct {
	Name         string
	Price        string
	SavingsPrice string
}

func RimiScraper() ([]Product, error) {
	c := colly.NewCollector()

	var productIds = []int{497705, 7006629, 200280}
	var products []Product

	c.OnHTML(".js-product-container", func(e *colly.HTMLElement) {
		memberPrice := e.ChildText(".price-badge__price span")
		if len(memberPrice) != 0 {
			memberPrice = "0," + memberPrice[1:]
		}

		singleProduct := Product{
			Name:         e.ChildText(".card__name"),
			Price:        e.ChildText(".price-tag.card__price span") + "," + e.ChildText(".price-tag.card__price div sup") + e.ChildText(".price-tag.card__price div sub"),
			SavingsPrice: memberPrice,
		}

		products = append(products, singleProduct)
	})

	queryString := ""
	for i, id := range productIds {
		if i > 0 {
			queryString += " "
		}
		queryString += strconv.Itoa(id)
	}
	url := fmt.Sprintf("https://www.rimi.ee/epood/ee/otsing?query=%s", strings.ReplaceAll(queryString, " ", "%20"))
	err := c.Visit(url)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape Rimi data: %v", err)
	}

	// Get current timestamp
	t := time.Now()
	date := t.Format("2006-01-02")
	time := t.Format("15:04:05")

	fileExists := false
	if _, err := os.Stat("products.csv"); err == nil {
		fileExists = true
	}

	// Open CSV file in append mode
	file, err := os.OpenFile("products.csv", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV headers

	if !fileExists {
		err = writer.Write([]string{"Date", "Time", "Product Name", "Price", "Member Price"})
		if err != nil {
			return nil, fmt.Errorf("failed to write CSV headers: %v", err)
		}
	}

	// Write product data to CSV with separate date and time columns

	for _, product := range products {
		var savingsPrice string
		if product.SavingsPrice != "" {
			savingsPrice = product.SavingsPrice + "/tk"
		} else {
			savingsPrice = product.Price
		}
		err = writer.Write([]string{date, time, product.Name, product.Price, savingsPrice})
		if err != nil {
			return nil, fmt.Errorf("failed to write product data to CSV: %v", err)
		}
	}

	return products, nil
}
