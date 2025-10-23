package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

// Product represents a product in our catalog
type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Brand       string `json:"brand"`
}

// SearchResponse represents the API response
type SearchResponse struct {
	Products   []Product `json:"products"`
	TotalFound int       `json:"total_found"`
	SearchTime string    `json:"search_time"`
}

// ProductStore manages our products with thread-safe operations
type ProductStore struct {
	products sync.Map
	count    atomic.Int64
}

// Sample data for variety
var (
	brands      = []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta", "Theta", "Iota", "Kappa"}
	categories  = []string{"Electronics", "Books", "Home", "Sports", "Toys", "Fashion", "Garden", "Automotive", "Health", "Office"}
	adjectives  = []string{"Premium", "Professional", "Essential", "Ultimate", "Advanced", "Basic", "Deluxe", "Standard", "Plus", "Pro"}
	productType = []string{"Device", "Tool", "Kit", "System", "Solution", "Package", "Bundle", "Set", "Collection", "Series"}
)

// NewProductStore creates and initializes a product store
func NewProductStore() *ProductStore {
	ps := &ProductStore{}
	ps.generateProducts(100000)
	return ps
}

// generateProducts creates 100,000 products with variety
func (ps *ProductStore) generateProducts(count int) {
	rand.Seed(time.Now().UnixNano())

	for i := 1; i <= count; i++ {
		// Create varied product names
		brand := brands[i%len(brands)]
		adj := adjectives[rand.Intn(len(adjectives))]
		pType := productType[rand.Intn(len(productType))]

		product := Product{
			ID:          i,
			Name:        fmt.Sprintf("Product %s %d", brand, i),
			Category:    categories[i%len(categories)],
			Description: fmt.Sprintf("%s %s %s - High quality product for your needs", adj, brand, pType),
			Brand:       brand,
		}

		ps.products.Store(i, product)
		ps.count.Add(1)
	}

	log.Printf("Generated %d products", count)
}

// boundedSearch performs a search that checks exactly 100 products
func (ps *ProductStore) boundedSearch(query string) ([]Product, int, time.Duration) {
	start := time.Now()
	query = strings.ToLower(query)

	var results []Product
	totalFound := 0
	checked := 0

	// Start from a random position for variety in results
	startID := rand.Intn(int(ps.count.Load())-100) + 1

	// Check exactly 100 products
	for i := startID; i < startID+100 && checked < 100; i++ {
		checked++

		if value, ok := ps.products.Load(i); ok {
			product := value.(Product)

			// Case-insensitive search in name and category
			if strings.Contains(strings.ToLower(product.Name), query) ||
				strings.Contains(strings.ToLower(product.Category), query) {
				totalFound++

				// Only keep first 20 results
				if len(results) < 20 {
					results = append(results, product)
				}
			}
		}
	}

	return results, totalFound, time.Since(start)
}

// Server handles HTTP requests
type Server struct {
	store *ProductStore
}

// NewServer creates a new server instance
func NewServer() *Server {
	return &Server{
		store: NewProductStore(),
	}
}

// searchHandler handles search requests
func (s *Server) searchHandler(ctx *fasthttp.RequestCtx) {
	// Set response headers
	ctx.SetContentType("application/json")
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")

	// Get query parameter
	query := string(ctx.QueryArgs().Peek("q"))
	if query == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		json.NewEncoder(ctx).Encode(map[string]string{"error": "query parameter 'q' is required"})
		return
	}

	// Perform bounded search
	products, totalFound, duration := s.store.boundedSearch(query)

	// Create response
	response := SearchResponse{
		Products:   products,
		TotalFound: totalFound,
		SearchTime: fmt.Sprintf("%.3fs", duration.Seconds()),
	}

	// Send response
	if err := json.NewEncoder(ctx).Encode(response); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		json.NewEncoder(ctx).Encode(map[string]string{"error": "failed to encode response"})
	}
}

// statsHandler returns store statistics
func (s *Server) statsHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	stats := map[string]interface{}{
		"total_products":      s.store.count.Load(),
		"products_per_search": 100,
		"max_results":         20,
		"categories":          categories,
		"brands":              brands,
	}

	json.NewEncoder(ctx).Encode(stats)
}

// requestHandler routes requests
func (s *Server) requestHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/search":
		if string(ctx.Method()) == "GET" {
			s.searchHandler(ctx)
		} else {
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		}
	case "/stats":
		if string(ctx.Method()) == "GET" {
			s.statsHandler(ctx)
		} else {
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		}
	default:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.WriteString("Not Found")
	}
}

func main() {
	server := NewServer()

	log.Println("Server starting on :8080")
	log.Println("Endpoints:")
	log.Println("  GET /search?q=<query> - Search products (checks 100 products)")
	log.Println("  GET /stats - Get server statistics")

	if err := fasthttp.ListenAndServe(":8080", server.requestHandler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
