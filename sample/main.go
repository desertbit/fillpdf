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
	// Fill the form PDF with our values.
	err = fillpdf.Stamp("stamp.pdf", "form.pdf", "filled.pdf", true)
	if err != nil {
		log.Fatal(err)
	}
}
