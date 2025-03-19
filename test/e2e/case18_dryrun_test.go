// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"open-cluster-management.io/governance-policy-propagator/controllers/dryrun"
)

var _ = FDescribe("Test policy webhook", Ordered, func() {
	const ()
	var cancel context.CancelFunc

	BeforeAll(func() {
		// ctx, cancelFunc := context.WithCancel(context.Background())
		// cancel = cancelFunc

		// //
		// cmd = exec.CommandContext(ctx, "kubectl", "port-forward", "svc/dryrun-service", "8090:8090", "-n", "open")
		// var stderr bytes.Buffer
		// cmd.Stderr = &stderr

		// err := cmd.Start()
		// Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Failed to start port-forward: %s", stderr.String()))
	})
	AfterAll(func() {
		if cancel != nil {
			cancel() // Stops the port-forward process
		}
	})
	Describe("Test dryrun", func() {
		It("Should the error message is presented", func() {
			input, err := os.ReadFile(path.Join("../resources/case18_dryrun", "case18_dryrun_input.yaml"))
			Expect(err).ShouldNot(HaveOccurred())

			policy, err := os.ReadFile(path.Join("../resources/case18_dryrun", "case18_dryrun_policy.yaml"))
			Expect(err).ShouldNot(HaveOccurred())

			requestData := dryrun.RequestData{
				InputResources: string(input),
				Policy:         string(policy),
			}

			s, err := json.Marshal(requestData)
			Expect(err).ShouldNot(HaveOccurred())

			getRes, err := http.Get("http://localhost:8090/dryrun/health")
			Expect(err).NotTo(HaveOccurred())
			defer getRes.Body.Close()

			body, err := io.ReadAll(getRes.Body)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(string(body)).Should(Equal("Hello"))

			postRes, err := http.Post("http://localhost:8090/dryrun", "application/json", bytes.NewReader(s))
			Expect(err).NotTo(HaveOccurred())
			defer postRes.Body.Close()

			Expect(postRes.StatusCode).To(Equal(http.StatusOK))

			response := dryrun.ResponseData{}

			err = json.NewDecoder(postRes.Body).Decode(&response)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(response.Result).Should(Equal(`# Diffs:
v1 Pod default/pod-case-18:

# Compliance messages:
Compliant; notification - pods [pod-case-18] found as specified in namespace default
`))
		})
	})
})
