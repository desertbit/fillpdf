package fillpdf

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Field contains information about the fields exported from a pdf via pdftk
type Field struct {
	Type    string
	Name    string
	AltName string
	Flags   string
}

func GetFields(formPDFFile string) ([]Field, error) {
	formPDFFile, err := filepath.Abs(formPDFFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create the absolute path: %v", err)
	}

	// Check if the form file exists.
	e, err := exists(formPDFFile)
	if err != nil {
		return nil, fmt.Errorf("failed to check if form PDF file exists: %v", err)
	} else if !e {
		return nil, fmt.Errorf("form PDF file does not exists: '%s'", formPDFFile)
	}

	// Check if the pdftk utility exists.
	_, err = exec.LookPath("pdftk")
	if err != nil {
		return nil, fmt.Errorf("pdftk utility is not installed")
	}

	// Create the pdftk command line arguments.
	args := []string{
		formPDFFile,
		"dump_data_fields",
	}

	output, err := runCommandWithResults("pdftk", args...)
	if err != nil {
		return nil, fmt.Errorf("pdftk error: %v", err)
	}

	fieldsData := strings.Split(output, "---\n")

	fields := []Field{}
	for _, f := range fieldsData {
		lines := strings.Split(f, "\n")
		if len(lines) <= 2 {
			continue
		}

		field := Field{}

		for _, line := range lines {
			props := strings.Split(line, ": ")

			if len(props) != 2 {
				continue
			}

			switch props[0] {
			case "FieldType":
				field.Type = strings.ToLower(props[1])
			case "FieldName":
				field.Name = props[1]
			case "FieldNameAlt":
				field.AltName = props[1]
			case "FieldFlags":
				field.Flags = props[1]
			}
		}

		fields = append(fields, field)
	}

	return fields, nil
}
