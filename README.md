# FillPDF

FillPDF is a golang library to easily fill PDF forms. This library uses the pdftk utility to fill the PDF forms with fdf data.
Currently this library only supports PDF text field values. Feel free to add support to more form types.


## Documentation 

Check the Documentation at [GoDoc.org](https://godoc.org/github.com/desertbit/fillpdf).

## Requirements

FillPDF, under the hood, leverages the toolchain provided by PDFtk. Windows and Mac users need to install this dependency separately, the pdftk-sever executable is available [here](https://www.pdflabs.com/tools/pdftk-server/). After the installation is complete ensure that the install directory has been added to the system PATH (should be added automatically during the installation process).


## Sample

There is an example in the sample directory:

```go
package main

import (
	"log"

	"github.com/desertbit/fillpdf"
)

func main() {
	// Create the form values.
	form := fillpdf.Form{
		"field_1": "Hello",
		"field_2": "World",
	}

	// Fill the form PDF with our values.
	err := fillpdf.Fill(form, "form.pdf", "filled.pdf", true)
	if err != nil {
		log.Fatal(err)
	}
}
```

Run the example as following:

```
cd sample
go build
./sample
```
