package main

import (
	"log"

	"github.com/desertbit/fillpdf"
)

func main() {
	// Create the form values.
	form := fillpdf.Form{
		"field_1": "Hello",
		"field_2": "Wörld",
	}

	// Fill the form PDF with our values.
	err := fillpdf.Fill(form, "form.pdf", "filled.pdf")
	if err != nil {
		log.Fatal(err)
	}

	fields, err := fillpdf.GetFields("form.pdf")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", fields)
}
