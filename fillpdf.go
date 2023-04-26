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
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gdamore/encoding"
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
	// Overwrite will overwrite any pre existing filled PDF
	Overwrite bool
	// Flatten will flatten the document making the form fields no longer editable
	Flatten bool
	// Remove metadata
	RemoveMetadata bool
}

func defaultOptions() Options {
	return Options{
		Overwrite:      true,
		Flatten:        true,
		RemoveMetadata: false,
	}
}

// Fill a PDF form with the specified form values and create a final filled PDF file.
// The options parameter alters few aspects of the generation.
func Fill(form Form, formPDFFile, destPDFFile string, options ...Options) (err error) {
	// If the user provided the options we overwrite the defaults with the given struct.
	opts := defaultOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	// Get the absolute paths.
	formPDFFile, err = filepath.Abs(formPDFFile)
	if err != nil {
		return fmt.Errorf("failed to create the absolute path: %v", err)
	}
	destPDFFile, err = filepath.Abs(destPDFFile)
	if err != nil {
		return fmt.Errorf("failed to create the absolute path: %v", err)
	}

	// Check if the form file exists.
	e, err := exists(formPDFFile)
	if err != nil {
		return fmt.Errorf("failed to check if form PDF file exists: %v", err)
	} else if !e {
		return fmt.Errorf("form PDF file does not exists: '%s'", formPDFFile)
	}

	// Check if the pdftk utility exists.
	_, err = exec.LookPath("pdftk")
	if err != nil {
		return fmt.Errorf("pdftk utility is not installed!")
	}

	// Create a temporary directory.
	tmpDir, err := ioutil.TempDir("", "fillpdf-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}

	var metadataFile string
	if opts.RemoveMetadata {
		metadataFile = filepath.Clean(tmpDir + "/metadata.tmp")
		errM := createMetadataFile(formPDFFile, metadataFile)
		if errM != nil {
			return fmt.Errorf("failed to create metadata file: %v", errM)
		}
	}

	// Remove the temporary directory on defer again.
	defer func() {
		errD := os.RemoveAll(tmpDir)
		// Log the error only.
		if errD != nil {
			log.Printf("fillpdf: failed to remove temporary directory '%s' again: %v", tmpDir, errD)
		}
	}()

	// Create the temporary output file path.
	outputFile := filepath.Clean(tmpDir + "/output.pdf")

	// Create the fdf data file.
	fdfFile := filepath.Clean(tmpDir + "/data.fdf")
	err = createFdfFile(form, fdfFile)
	if err != nil {
		return fmt.Errorf("failed to create fdf form data file: %v", err)
	}

	// Create the pdftk command line arguments.
	args := []string{
		formPDFFile,
		"fill_form", fdfFile,
		"output", outputFile,
	}

	// If the user specified to flatten the output PDF we append the related parameter.
	if opts.Flatten {
		args = append(args, "flatten")
	}

	// Run the pdftk utility.
	_, err = runCommandInPath(tmpDir, "pdftk", args...)
	if err != nil {
		return fmt.Errorf("pdftk error: %v", err)
	}

	if opts.RemoveMetadata {
		outputFile2 := filepath.Clean(tmpDir + "output2.pdf")
		args = []string{
			outputFile,
			"update_info", metadataFile,
			"output", outputFile2,
		}
		_, err = runCommandInPath(tmpDir, "pdftk", args...)
		if err != nil {
			return fmt.Errorf("pdftk error: %v", err)
		}
		outputFile = outputFile2
	}

	// Check if the destination file exists.
	e, err = exists(destPDFFile)
	if err != nil {
		return fmt.Errorf("failed to check if destination PDF file exists: %v", err)
	} else if e {
		if !opts.Overwrite {
			return fmt.Errorf("destination PDF file already exists: '%s'", destPDFFile)
		}

		err = os.Remove(destPDFFile)
		if err != nil {
			return fmt.Errorf("failed to remove destination PDF file: %v", err)
		}
	}

	// On success, copy the output file to the final destination.
	err = copyFile(outputFile, destPDFFile)
	if err != nil {
		return fmt.Errorf("failed to copy created output PDF to final destination: %v", err)
	}

	return nil
}

func createFdfFile(form Form, path string) error {
	// Create the file.
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new writer.
	w := bufio.NewWriter(file)

	// Write the fdf header.
	fmt.Fprintln(w, fdfHeader)

	// Write the form data.
	var valueStr string
	for key, value := range form {
		// Convert to Latin-1.
		valueStr, err = latin1Encoder.String(fmt.Sprintf("%v", value))
		if err != nil {
			return fmt.Errorf("failed to convert string to Latin-1")
		}
		fmt.Fprintf(w, "<< /T (%s) /V (%s)>>\n", key, valueStr)
	}

	// Write the fdf footer.
	fmt.Fprintln(w, fdfFooter)

	// Flush everything.
	return w.Flush()
}

func createMetadataFile(formPDFFile, path string) error {
	// Create the file.
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	out, err := runCommandInPath("", "pdftk", formPDFFile, "dump_data")
	if err != nil {
		return err
	}

	lines := strings.Split(string(out.Bytes()), "\n")
	for i := range lines {
		if strings.HasPrefix(lines[i], "InfoValue:") {
			lines[i] = "InfoValue:"
		}
	}

	_, err = file.Write([]byte(strings.Join(lines, "\n")))
	return err
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
