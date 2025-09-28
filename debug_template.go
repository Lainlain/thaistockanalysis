package main

import (
	"html/template"
	"log"
)

func main() {
	_, err := template.ParseFiles("web/templates/base.gohtml")
	if err != nil {
		log.Fatal("Template error:", err)
	}
	log.Println("Template OK")
}
