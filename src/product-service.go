package main

import (
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

type Product struct {
	ProductID    int    `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   int    `json:"category_id"`
	Weight       int    `json:"weight"`
	SomeOtherID  int    `json:"some_other_id"`
}

// Pointers let us tell "field missing" apart from "field is 0".
type productRequest struct {
	ProductID    *int    `json:"product_id"`
	SKU          *string `json:"sku"`
	Manufacturer *string `json:"manufacturer"`
	CategoryID   *int    `json:"category_id"`
	Weight       *int    `json:"weight"`
	SomeOtherID  *int    `json:"some_other_id"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type store struct {
	mu       sync.RWMutex
	products map[int]Product
}

var db = &store{products: make(map[int]Product)}

func (s *store) get(id int) (Product, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.products[id]
	return p, ok
}

func (s *store) update(p Product) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.products[p.ProductID]; !ok {
		return false
	}
	s.products[p.ProductID] = p
	return true
}

// seed runs before the server starts, so no lock is needed.
func seed(n int) {
	for i := 1; i <= n; i++ {
		db.products[i] = Product{i, "SKU-" + strconv.Itoa(i), "Acme Corporation", 1, 1000, 1}
	}
}

func parseProductID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("productId"))
	return id, err == nil && id >= 1
}

func validate(r productRequest) (Product, string) {
	switch {
	case r.ProductID == nil || *r.ProductID < 1:
		return Product{}, "product_id is required and must be >= 1"
	case r.SKU == nil || len(*r.SKU) < 1 || len(*r.SKU) > 100:
		return Product{}, "sku is required (1-100 chars)"
	case r.Manufacturer == nil || len(*r.Manufacturer) < 1 || len(*r.Manufacturer) > 200:
		return Product{}, "manufacturer is required (1-200 chars)"
	case r.CategoryID == nil || *r.CategoryID < 1:
		return Product{}, "category_id is required and must be >= 1"
	case r.Weight == nil || *r.Weight < 0:
		return Product{}, "weight is required and must be >= 0"
	case r.SomeOtherID == nil || *r.SomeOtherID < 1:
		return Product{}, "some_other_id is required and must be >= 1"
	}
	return Product{*r.ProductID, *r.SKU, *r.Manufacturer, *r.CategoryID, *r.Weight, *r.SomeOtherID}, ""
}

func getProduct(c *gin.Context) {
	id, ok := parseProductID(c)
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "NOT_FOUND", Message: "product not found"})
		return
	}
	p, found := db.get(id)
	if !found {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "NOT_FOUND", Message: "product not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func addProductDetails(c *gin.Context) {
	id, ok := parseProductID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "INVALID_INPUT", Message: "productId must be a positive integer"})
		return
	}
	var req productRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "INVALID_INPUT", Message: "malformed JSON body", Details: err.Error()})
		return
	}
	p, msg := validate(req)
	if msg != "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "INVALID_INPUT", Message: "invalid product", Details: msg})
		return
	}
	if p.ProductID != id {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "INVALID_INPUT", Message: "product_id in body must match path"})
		return
	}
	if !db.update(p) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "NOT_FOUND", Message: "product not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func main() {
	seed(1000)
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.CustomRecovery(func(c *gin.Context, _ any) {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			ErrorResponse{Error: "INTERNAL_ERROR", Message: "internal server error"})
	}))

	r.GET("/products/:productId", getProduct)
	r.POST("/products/:productId/details", addProductDetails)

	if os.Getenv("ENABLE_DEBUG") == "1" { // the only way to demo a 500
		r.GET("/debug/panic", func(c *gin.Context) { panic("forced") })
	}
	r.Run(":8080")
}
