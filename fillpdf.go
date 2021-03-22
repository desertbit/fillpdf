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
)

// Form represents the PDF form.
// This is a key value map.
type Form map[string]interface{}

// Options represents the options to alter the PDF filling process
// Overwrite will overwrite any pre existing filled PDF
// Flatten will flatten the document making the form field no longer editable
type Options struct {
	Overwrite bool
	Flatten   bool
}

// Fill a PDF form with the specified form values and create a final filled PDF file.
// The options parameter alters few aspects of the generation.
func Fill(form Form, formPDFFile, destPDFFile string, options Options) (err error) {
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

	if options.Flatten {
		args = append(args, "flatten")
	}

	// Run the pdftk utility.
	err = runCommandInPath(tmpDir, "pdftk", args...)
	if err != nil {
		return fmt.Errorf("pdftk error: %v", err)
	}

	// Check if the destination file exists.
	e, err = exists(destPDFFile)
	if err != nil {
		return fmt.Errorf("failed to check if destination PDF file exists: %v", err)
	} else if e {
		if !options.Overwrite {
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
	for key, value := range form {
		fmt.Fprintf(w, "<< /T (%s) /V (%v)>>\n", key, value)
	}

	// Write the fdf footer.
	fmt.Fprintln(w, fdfFooter)

	// Flush everything.
	return w.Flush()
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
