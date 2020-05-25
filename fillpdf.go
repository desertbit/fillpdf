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

// Fill a PDF form with the specified form values and create a final filled PDF file.
// One variadic boolean specifies, whenever to overwrite the destination file if it exists.
func Fill(form Form, formPDFFile, destPDFFile string, overwrite ...bool) (err error) {
	// Get the absolute paths.
	formPDFFile, err = filepath.Abs(formPDFFile)
	if err != nil {
		return fmt.Errorf("failed to create the absolute path: %+v", err)
	}
	destPDFFile, err = filepath.Abs(destPDFFile)
	if err != nil {
		return fmt.Errorf("failed to create the absolute path: %+v", err)
	}

	// Check if the form file exists.
	e, err := exists(formPDFFile)
	if err != nil {
		return fmt.Errorf("failed to check if form PDF file exists: %+v", err)
	} else if !e {
		return fmt.Errorf("form PDF file does not exists: '%s'", formPDFFile)
	}

	// Check if the pdftk utility exists.
	_, err = exec.LookPath("pdftk")
	if err != nil {
		return fmt.Errorf("pdftk utility is not installed")
	}

	// Create a temporary directory.
	tmpDir, err := ioutil.TempDir("", "fillpdf-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %+v", err)
	}

	// Remove the temporary directory on defer again.
	defer func() {
		errD := os.RemoveAll(tmpDir)
		// Log the error only.
		if errD != nil {
			log.Printf("fillpdf: failed to remove temporary directory '%s' again: %+v", tmpDir, errD)
		}
	}()

	// 1. Generate an FDF file.
	// create the temporary intermediate output file path.
	intermediateOutputFile := filepath.Clean(tmpDir + "/intermediate-output.pdf")

	// create an FDF file from the input form.
	inputFDFFile := filepath.Clean(tmpDir + "/input.fdf")
	err = createFDFFile(form, inputFDFFile)
	if err != nil {
		return fmt.Errorf("failed to create FDF file from the input form: %+v", err)
	}

	// 2. Generate a PDF file with need_appearances.
	// create the pdftk command-line arguments.
	args := []string{
		formPDFFile,
		"fill_form", inputFDFFile,
		"output", intermediateOutputFile,
		"need_appearances",
	}

	// run the pdftk utility.
	err = runCommandInPath(tmpDir, "pdftk", args...)
	if err != nil {
		return fmt.Errorf("pdftk error: %+v\n%s\n%+v", err, tmpDir, args)
	}

	// 3. Export an FDF file out of it.
	// create the pdftk command-line arguments.
	exportedFDF := filepath.Clean(tmpDir + "/exported.fdf")
	args = []string{
		intermediateOutputFile,
		"generate_FDF",
		"output", exportedFDF,
	}
	// run the pdftk utility.
	err = runCommandInPath(tmpDir, "pdftk", args...)
	if err != nil {
		return fmt.Errorf("pdftk error: %+v\n%s\n%+v", err, tmpDir, args)
	}

	// 4. Generate a flattened PDF file out of the exported FDF file.
	// create the pdftk command-line arguments.
	finalOutputFile := filepath.Clean(tmpDir + "/final-output.pdf")
	args = []string{
		intermediateOutputFile,
		"fill_form", exportedFDF,
		"output", finalOutputFile,
		"flatten",
	}
	// run the pdftk utility.
	err = runCommandInPath(tmpDir, "pdftk", args...)
	if err != nil {
		return fmt.Errorf("pdftk error: %+v\n%s\n%+v", err, tmpDir, args)
	}

	// Check if the destination file exists.
	e, err = exists(destPDFFile)
	if err != nil {
		return fmt.Errorf("failed to check if destination PDF file exists: %+v", err)
	} else if e {
		if len(overwrite) == 0 || !overwrite[0] {
			return fmt.Errorf("destination PDF file already exists: '%s'", destPDFFile)
		}

		err = os.Remove(destPDFFile)
		if err != nil {
			return fmt.Errorf("failed to remove destination PDF file: %+v", err)
		}
	}

	// On success, copy the output file to the final destination.
	err = copyFile(finalOutputFile, destPDFFile)
	if err != nil {
		return fmt.Errorf("failed to copy created output PDF to final destination: %+v", err)
	}

	return nil
}

func createFDFFile(form Form, path string) error {
	// Create the file.
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new writer.
	w := bufio.NewWriter(file)

	// Write the FDF header.
	fmt.Fprintln(w, FDFHeader)

	// Write the form data.
	for key, value := range form {
		fmt.Fprintf(w, "<< /T (%s) /V (%v)>>\n", key, value)
	}

	// Write the FDF footer.
	fmt.Fprintln(w, FDFFooter)

	// Flush everything.
	return w.Flush()
}

// FDFHeader marks the begining of the FDF file.
const FDFHeader = `%FDF-1.2
%,,oe"
1 0 obj
<<
/FDF << /Fields [`

// FDFFooter marks the end of the FDF file.
const FDFFooter = `]
>>
>>
endobj
trailer
<<
/Root 1 0 R
>>
%%EOF`
