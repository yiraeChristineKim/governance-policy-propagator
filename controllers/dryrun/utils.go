package dryrun

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
)

// validateRequiredFields checks for empty required fields.
func validateRequiredFields(data RequestData) error {
	if data.Policy == "" {
		return errors.New("empty policy")
	}

	if data.InputResources == "" {
		return errors.New("empty resources")
	}

	return nil
}

// createAndRegisterTempFile creates a temporary file and registers it for cleanup.
func createAndRegisterTempFile(fileName, content string, tempFiles *[]*os.File) (*os.File, error) {
	tmpFile, err := os.CreateTemp("", fileName)
	if err != nil {
		return nil, err
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())

		return nil, err
	}

	*tempFiles = append(*tempFiles, tmpFile)

	return tmpFile, nil
}

// handleFileError logs an error and responds with an HTTP error.
func handleFileError(w http.ResponseWriter, log logr.Logger, err error, message string) {
	log.Error(err, message)
	http.Error(w, message, http.StatusInternalServerError)
}

// cleanupTempFiles ensures all temporary files are closed and deleted.
func cleanupTempFiles(files []*os.File, log logr.Logger) {
	for _, file := range files {
		file.Close()

		if err := os.Remove(file.Name()); err != nil {
			log.Error(err, "Failed to delete temp file", "file", file.Name())
		}
	}
}

// processOptionalFile creates and registers an optional temporary file.
func processOptionalFile(cmd *cobra.Command, flag, content string,
	tempFiles *[]*os.File, log logr.Logger, w http.ResponseWriter,
) error {
	if content == "" {
		return nil
	}

	file, err := createAndRegisterTempFile(flag+".yaml", content, tempFiles)
	if err != nil {
		handleFileError(w, log, err, fmt.Sprintf("Failed to create %s file", flag))

		return err
	}

	if err := cmd.Flags().Set(flag, file.Name()); err != nil {
		handleFileError(w, log, err, fmt.Sprintf("Failed to set %s file", flag))

		return err
	}

	return nil
}
