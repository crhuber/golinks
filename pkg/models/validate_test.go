package models

import (
	"strings"
	"testing"
)

func TestLinkInputValidate(t *testing.T) {
	ok := LinkInput{
		Keyword:     "gh/{*}",
		Destination: "https://github.com/{*}",
		Tags:        []Tag{{Name: "x"}},
	}
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid link: %v", err)
	}

	cases := []LinkInput{
		{Destination: "https://example.com"},
		{Keyword: "docs", Destination: "ftp://example.com"},
		{Keyword: "docs", Destination: "example.com"},
		{Keyword: "api", Destination: "https://example.com"},
		{Keyword: "static", Destination: "https://example.com"},
		{Keyword: "directory", Destination: "https://example.com"},
		{Keyword: "healthz", Destination: "https://example.com"},
		{Keyword: "favicon.ico", Destination: "https://example.com"},
		{Keyword: strings.Repeat("a", 101), Destination: "https://example.com"},
		{Keyword: " bad", Destination: "https://example.com"},
	}
	for _, in := range cases {
		if err := in.Validate(); err == nil {
			t.Fatalf("expected error for %#v", in)
		}
	}
}

func TestQueryStringValidate(t *testing.T) {
	if err := (QueryString{Sort: "updated_at", Order: "desc"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (QueryString{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (QueryString{Sort: "123", Order: "asc"}).Validate(); err == nil {
		t.Fatal("expected sort error")
	}
	if err := (QueryString{Sort: "keyword", Order: "123"}).Validate(); err == nil {
		t.Fatal("expected order error")
	}
}

func TestTagValidate(t *testing.T) {
	if err := (Tag{Name: "code"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Tag{Name: "ab"}).Validate(); err == nil {
		t.Fatal("expected short name error")
	}
	if err := (Tag{Name: ""}).Validate(); err == nil {
		t.Fatal("expected empty name error")
	}
	if err := (Tag{Name: strings.Repeat("a", 21)}).Validate(); err == nil {
		t.Fatal("expected long name error")
	}
}
