package main

import (
	"log"

	"github.com/desertbit/fillpdf"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func main() {
	// Create the form values.
	form := fillpdf.Form{
		"field_1": "Hello",
		"field_2": "Wörld",
	}

	// Fill the form PDF with our values.
	err := fillpdf.Fill(form, "form.pdf", "filled.pdf", fillpdf.Options{
		RemoveMetadata: wrapperspb.Bool(true),
	})
	if err != nil {
		log.Fatal(err)
	}
}
