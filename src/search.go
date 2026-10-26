package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

const (
	numProducts    = 100000 // memory footprint
	checksPerQuery = 100    // fixed work per request: check exactly 100, matched or not
	maxResults     = 20     // spec: total_found maximum is 20
	maxQueryLen    = 100    // spec: q maxLength
)

var (
	brands = []string{"Alpha", "Beta", "Gamma", "Delta", "Nova", "Orion", "Vertex", "Zenith"}
	// category_id N (1-based, spec minimum 1) maps to categories[N-1]
	categories = []string{"Electronics", "Books", "Home", "Garden", "Toys", "Sports", "Clothing", "Beauty", "Grocery", "Automotive"}

	searchStore sync.Map // productID (int) -> Product

	// Fixed CPU cost per checked product, set via SEARCH_WORK_ROUNDS (0 = no extra work).
	workRounds = envInt("SEARCH_WORK_ROUNDS", 0)
	workSink   atomic.Uint32 // keeps the compiler from discarding the hashing
)

func envInt(key string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil && v >= 0 {
		return v
	}
	return def
}

// simulateWork burns a fixed, deterministic amount of CPU per product, standing in for the
// "run a model / process a frame" work the assignment describes. Same cost on every call.
func simulateWork(name string) byte {
	if workRounds == 0 {
		return 0
	}
	h := sha256.Sum256([]byte(name))
	for i := 1; i < workRounds; i++ {
		h = sha256.Sum256(h[:])
	}
	return h[0]
}

// seedSearchProducts runs once at startup, before r.Run().
func seedSearchProducts() {
	for i := 1; i <= numProducts; i++ {
		brand := brands[i%len(brands)]
		searchStore.Store(i, Product{
			ProductID:    i,
			SKU:          fmt.Sprintf("SKU-%06d", i),
			Manufacturer: brand,
			CategoryID:   i%len(categories) + 1,
			Weight:       100 + i%900,
			SomeOtherID:  i,
			Name:         fmt.Sprintf("Product %s %d", brand, i),
		})
	}
}

// searchCatalog checks exactly checksPerQuery products and returns up to maxResults matches,
// plus the number matched and the number checked (both used by the test and the header).
func searchCatalog(query string) (results []Product, matched, checked int) {
	needle := strings.ToLower(query)
	results = make([]Product, 0, maxResults)

	var acc byte
	searchStore.Range(func(_, v any) bool {
		p := v.(Product)
		checked++
		acc ^= simulateWork(p.Name)
		if strings.Contains(strings.ToLower(p.Name), needle) ||
			strings.Contains(strings.ToLower(categories[p.CategoryID-1]), needle) {
			matched++
			if len(results) < maxResults {
				results = append(results, p)
			}
		}
		return checked < checksPerQuery // stop after exactly 100 checks
	})
	workSink.Store(uint32(acc))
	return results, matched, checked
}

// searchProducts handles GET /products/search?q=...
func searchProducts(c *gin.Context) {
	q, ok := c.GetQuery("q")
	if n := utf8.RuneCountInString(q); !ok || n < 1 || n > maxQueryLen {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "INVALID_INPUT", Message: "query parameter q is required", Details: "q must be 1 to 100 characters",
		})
		return
	}

	start := time.Now()
	results, matched, checked := searchCatalog(q)
	elapsed := time.Since(start)

	// Not part of the response schema: headers that make the fixed-work bound visible with curl -i.
	c.Header("X-Products-Checked", strconv.Itoa(checked))
	c.Header("X-Products-Matched", strconv.Itoa(matched))

	c.JSON(http.StatusOK, gin.H{
		"products":    results,
		"total_found": len(results), // spec caps this at 20
		"search_time": fmt.Sprintf("%.6fs", elapsed.Seconds()),
	})
}

// health handles GET /health for the ALB target group.
func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
