package dryrun

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"open-cluster-management.io/config-policy-controller/pkg/dryrun"
	ctrl "sigs.k8s.io/controller-runtime"
)

// RequestData defines the request structure.
type RequestData struct {
	Mapping        string `json:"mapping"`
	Policy         string `json:"policy"`
	DesiredStatus  string `json:"desiredStatus"`
	InputResources string `json:"inputResources"`
}

type ResponseData struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Result  string `json:"result"`
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	log := ctrl.Log.WithName("dryrun")

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)

		return
	}

	var data RequestData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)

		return
	}

	// Required fields validation
	if err := validateRequiredFields(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	// Initialize dry-run command
	d := dryrun.DryRunner{}
	cmd := d.GetCmd()
	out := bytes.Buffer{}
	cmd.SetOut(&out)

	// Temporary files cleanup stack
	var tempFiles []*os.File
	defer cleanupTempFiles(tempFiles, log)

	//
	// Process required  input resources files
	inputFile, err := createAndRegisterTempFile("input.yaml", data.InputResources, &tempFiles)
	if err != nil {
		handleFileError(w, log, err, "Failed to create input resources file")

		return
	}

	cmd.SetArgs([]string{inputFile.Name()})

	// Policy
	policyFile, err := createAndRegisterTempFile("policy.yaml", data.Policy, &tempFiles)
	if err != nil {
		handleFileError(w, log, err, "Failed to create policy file")

		return
	}

	if err := cmd.Flags().Set("policy", policyFile.Name()); err != nil {
		handleFileError(w, log, err, "Failed to set policy file")

		return
	}

	// Process optional input files
	if err := processOptionalFile(cmd, "desired-status", data.DesiredStatus, &tempFiles, log, w); err != nil {
		return
	}

	if err := processOptionalFile(cmd, "mappings-file", data.Mapping, &tempFiles, log, w); err != nil {
		return
	}

	err = cmd.Execute()
	if err != nil && !errors.Is(err, dryrun.ErrNonCompliant) {
		handleFileError(w, log, err, err.Error())

		return
	}

	response := ResponseData{
		Message: "Post request successful",
		Status:  http.StatusOK,
		Result:  out.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Set status code

	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleFileError(w, log, err, "Failed to encode response")

		return
	}
}

func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Hello")
}
