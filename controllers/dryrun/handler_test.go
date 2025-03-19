package dryrun

import (
	"bytes"
	"embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

//go:embed testdata/*
var testfiles embed.FS

// TestPostHandler tests the PostHandler function.
func TestCompliantPostHandler(t *testing.T) {
	t.Run("Test compliant dryrun test", postHandlerTest("post_test"))
}

func TestNonCompliantPostHandler(t *testing.T) {
	t.Run("Test noncompliant dryrun test", postHandlerTest("post_noncompliant_test"))
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/dryrun/health", nil)

	rr := httptest.NewRecorder()

	HealthHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}
}

func postHandlerTest(testDir string) func(t *testing.T) {
	return func(t *testing.T) {
		r := readFiles(t, "Test dryrun post handler", testDir)

		s, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}

		// Create a mock POST request
		req := httptest.NewRequest(http.MethodPost, "/dryrun", bytes.NewReader(s))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		// Call the handler function
		PostHandler(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}

		// Check the response body
		var response ResponseData

		err = json.NewDecoder(rr.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to parse response JSON: %v", err)
		}

		out, err := os.ReadFile(path.Join("testdata", testDir, "output.txt"))
		if err != nil {
			t.Fatal(err)
		}

		expectedOutput := string(out)

		if response.Result != expectedOutput {
			t.Errorf("Expected message '%s', got '%s'", expectedOutput, response.Result)
		}
	}
}

// readFiles reads test scenario files from the "testdata" directory and populates a RequestData struct.
// It searches for files with specific prefixes (input, policy, mapping, desired) and assigns their content
// to the corresponding struct fields.
func readFiles(t *testing.T, testName, dir string) RequestData {
	var requestData RequestData

	entries, err := testfiles.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.Name() != dir {
			continue
		}

		scenarioFiles, err := testfiles.ReadDir(path.Join("testdata", dir))
		if err != nil {
			t.Fatal(err)
		}

		// Define a mapping of filename prefixes to struct fields
		fieldMap := map[string]*string{
			"input":   &requestData.InputResources,
			"policy":  &requestData.Policy,
			"mapping": &requestData.Mapping,
			"desired": &requestData.DesiredStatus,
		}

		for _, f := range scenarioFiles {
			for prefix, field := range fieldMap {
				if strings.HasPrefix(f.Name(), prefix) {
					data, err := os.ReadFile(path.Join("testdata", dir, f.Name()))
					if err != nil {
						t.Fatal(err)
					}
					*field = string(data)
					break
				}
			}
		}

		break // Exit once the directory is processed
	}

	return requestData
}
