/*
 *  FillPDF - Fill PDF forms
 *  Copyright DesertBit
 *  Author: Roland Singer
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package fillpdf

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/gdamore/encoding"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	// pdftk does not support UTF-8. To support at least some special characters,
	// let's use the Latin-1 encoding.
	latin1Encoder = encoding.ISO8859_1.NewEncoder()
)

// Form represents the PDF form.
// This is a key value map.
type Form map[string]interface{}

// Options represents the options to alter the PDF filling process
type Options struct {
	// Flatten will flatten the document making the form fields no longer editable
	Flatten *wrapperspb.BoolValue
	// Remove metadata
	RemoveMetadata *wrapperspb.BoolValue
}

func (o *Options) Override(opt Options) {
	if opt.Flatten != nil {
		o.Flatten = opt.Flatten
	}
	if opt.RemoveMetadata != nil {
		o.RemoveMetadata = opt.RemoveMetadata
	}
}

func defaultOptions() Options {
	return Options{
		Flatten:        wrapperspb.Bool(true),
		RemoveMetadata: wrapperspb.Bool(false),
	}
}

// Fill a PDF form with the specified form values and create a final filled PDF file.
// The options parameter alters few aspects of the generation.
func Fill(form Form, formPDFFile string, options ...Options) (out []byte, err error) {
	// If the user provided the options we overwrite the defaults with the given struct.
	opts := defaultOptions()
	for _, opt := range options {
		opts.Override(opt)
	}

	// Get the absolute paths.
	formPDFFile, err = filepath.Abs(formPDFFile)
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
		return nil, errors.New("pdftk utility is not installed!")
	}

	// Create the fdf content.
	fdfContent, err := createFdfFile(form)
	if err != nil {
		return nil, fmt.Errorf("failed to create fdf form data file: %v", err)
	}

	// Create the pdftk command line arguments.
	args := []string{
		formPDFFile,
		"fill_form", "-",
		"output", "-",
	}

	// If the user specified to flatten the output PDF we append the related parameter.
	if opts.Flatten.GetValue() {
		args = append(args, "flatten")
	}

	// Run the pdftk utility.
	output, err := runCommand("pdftk", bytes.NewBuffer([]byte(fdfContent)), args...)
	if err != nil {
		return nil, fmt.Errorf("pdftk error: %v", err)
	}

	if opts.RemoveMetadata.GetValue() {
		// Check if the exiftool utility exists.
		_, err = exec.LookPath("exiftool")
		if err != nil {
			return nil, errors.New("exiftool utility is not installed!")
		}
		// exiftool -all:all= - -o -
		output, err = runCommand("exiftool", output, "-all:all=", "-", "-o", "-")
		if err != nil {
			return nil, fmt.Errorf("exiftool error: %v", err)
		}
	}

	return output.Bytes(), nil
}

func createFdfFile(form Form) (output string, err error) {
	// Write the fdf header.
	output = fdfHeader

	// Write the form data.
	var valueStr string
	for key, value := range form {
		// Convert to Latin-1.
		valueStr, err = latin1Encoder.String(fmt.Sprintf("%v", value))
		if err != nil {
			return "", fmt.Errorf("failed to convert string to Latin-1")
		}
		output += fmt.Sprintf("<< /T (%s) /V (%s)>>\n", key, valueStr)
	}

	// Write the fdf footer.
	output += fdfFooter
	return output, nil
}

const fdfHeader = `%FDF-1.2
%,,oe"
1 0 obj
<<
/FDF << /Fields [`

const fdfFooter = `]
>>
>>
endobj
trailer
<<
/Root 1 0 R
>>
%%EOF`
