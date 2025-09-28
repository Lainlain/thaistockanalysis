package main
package main

import (
	"html/template"
	"log"
)

func main() {
	_, err := template.ParseFiles("web/templates/base.gohtml", "web/templates/index.gohtml")
	if err != nil {
		log.Fatal("Template parsing error:", err)
	}
	log.Println("Templates parsed successfully")
}
