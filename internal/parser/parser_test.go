package parser

import (
	"os"
	"testing"
)

func TestParseArticles(t *testing.T) {
	f, err := os.Open("testdata/small.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	articles, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	// Should skip ns=4 page (Wikipedia namespace)
	if len(articles) != 3 {
		t.Fatalf("expected 3 articles, got %d", len(articles))
	}

	cat := findByTitle(articles, "Cat")
	if cat == nil {
		t.Fatal("expected to find article 'Cat'")
	}
	if cat.ID != 100 {
		t.Fatalf("expected Cat ID=100, got %d", cat.ID)
	}

	// Cat should link to: Domestication, Species, Carnivore, Mammal, Dog
	expectedLinks := []string{"Domestication", "Species", "Carnivore", "Mammal", "Dog"}
	if len(cat.Links) != len(expectedLinks) {
		t.Fatalf("expected %d links from Cat, got %d: %v", len(expectedLinks), len(cat.Links), cat.Links)
	}
	for _, want := range expectedLinks {
		found := false
		for _, got := range cat.Links {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected Cat to link to %q, links: %v", want, cat.Links)
		}
	}
}

func findByTitle(articles []Article, title string) *Article {
	for i := range articles {
		if articles[i].Title == title {
			return &articles[i]
		}
	}
	return nil
}
