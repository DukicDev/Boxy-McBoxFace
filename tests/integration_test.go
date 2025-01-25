package tests

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestHostname(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "testBoxy", "../")
	err := buildCmd.Run()
	if err != nil {
		t.Fatalf("Failed to build the testbinary: %v", err)
	}
	defer os.Remove("testBoxy")

	runCmd := exec.Command("./testBoxy", "run", "hostname")

	var testStdout bytes.Buffer
	var testStderr bytes.Buffer
	runCmd.Stdout = &testStdout
	runCmd.Stderr = &testStderr

	err = runCmd.Run()
	if err != nil {
		t.Fatalf("Execution failed: %v, stderr: %s", err, testStderr.String())
	}

	output := strings.TrimSpace(testStdout.String())
	expected := "Boxy-McBoxFace"
	if output != expected {
		t.Errorf("Expected Output: %v, got %v", expected, output)
	}

	if testStderr.Len() != 0 {
		t.Errorf("Expected no stderr output, got: %v", testStderr.String())
	}
}

func TestUser(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "testBoxy", "../")
	err := buildCmd.Run()
	if err != nil {
		t.Fatalf("Failed to build the testbinary: %v", err)
	}
	defer os.Remove("testBoxy")

	runCmd := exec.Command("./testBoxy", "run", "whoami")

	var testStdout bytes.Buffer
	var testStderr bytes.Buffer
	runCmd.Stdout = &testStdout
	runCmd.Stderr = &testStderr

	err = runCmd.Run()
	if err != nil {
		t.Fatalf("Execution failed: %v, stderr: %s", err, testStderr.String())
	}

	output := strings.TrimSpace(testStdout.String())
	expected := "root"
	if output != expected {
		t.Errorf("Expected Output: %v, got %v", expected, output)
	}

	if testStderr.Len() != 0 {
		t.Errorf("Expected no stderr output, got: %v", testStderr.String())
	}
}
