package main

import "testing"

// Every generated name contains "product", so every checked item matches:
// matched == 100 proves exactly 100 were checked; results must be capped at 20.
func TestSearchChecksExactly100(t *testing.T) {
	seedSearchProducts()

	results, matched, checked := searchCatalog("PrOdUcT")
	if checked != checksPerQuery {
		t.Fatalf("checked %d products, want %d", checked, checksPerQuery)
	}
	if matched != checksPerQuery {
		t.Fatalf("matched %d, want %d", matched, checksPerQuery)
	}
	if len(results) != maxResults {
		t.Fatalf("returned %d results, want %d", len(results), maxResults)
	}

	if _, matched, checked := searchCatalog("no-such-term"); checked != checksPerQuery || matched != 0 {
		t.Fatalf("miss: checked=%d matched=%d, want 100 and 0", checked, matched)
	}
}
