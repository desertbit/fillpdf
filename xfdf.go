package fillpdf

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os"
)

type Fields struct {
	XMLName xml.Name `xml:"fields"`
	Field   []Field
}

type Field struct {
	XMLName xml.Name `xml:"field"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value"`
}
type XFDF struct {
	XMLName  xml.Name `xml:"xfdf"`
	XMLNS    string   `xml:"xmldn,attr"`
	XMLSpace string   `xml:"xml:space,attr"`
	Fields   Fields   `xml:"fields"`
}

const (
	xmlHeader    = `<?xml version="1.0" encoding="UTF-8"?>`
	xfdfNS       = "http://ns.adobe.com/xfdf/"
	xfdfXMLSpace = "preserve"
)

func createXFDFFile(form Form, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)

	fmt.Fprintln(w, xmlHeader)
	xfdfStruct := XFDF{
		XMLNS:    xfdfNS,
		XMLSpace: xfdfXMLSpace,
		Fields: Fields{
			Field: []Field{},
		},
	}
	for key, value := range form {
		xfdfStruct.Fields.Field = append(xfdfStruct.Fields.Field, Field{
			Name:  key,
			Value: fmt.Sprintf("%v", value),
		})
	}

	output, err := xml.Marshal(xfdfStruct)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, string(output))
	return w.Flush()
}
