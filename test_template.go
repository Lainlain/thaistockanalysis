package main

import (
	"html/template"
	"testing"
)

func TestTemplateParsing(t *testing.T) {
	_, err := template.ParseFiles("web/templates/base.gohtml", "web/templates/index.gohtml")
	if err != nil {
		t.Fatalf("Template parsing error: %v", err)
	}
	t.Log("Templates parsed successfully")
}
