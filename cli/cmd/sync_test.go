package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/dabblebox/stash/component/dotenv"
)

const stash = "./stash_test"

func TestMain(m *testing.M) {

	cmd := exec.Command("go", "build", "-o", stash, "../main.go")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	defer os.Remove(stash)
	log.Println(os.Getenv("AWS_PROFILE"))
	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("AWS_PROFILE", "devops")

	os.Setenv("STASH_CONTEXT", "stash-test")
	os.Setenv("STASH_WARN", "false")

	os.Exit(m.Run())
}

func TestSecretsManagerSyncEnv(t *testing.T) {
	// Arrange
	var catalog = fmt.Sprintf("%s_catalog.yml", t.Name())
	var inputFile = fmt.Sprintf("%s_input.env", t.Name())
	var resultsFile = fmt.Sprintf("%s_results.env", t.Name())

	d := []byte(`API_KEY=7fec6e3b-01bc-4b28-acc9-21028fe812b7
DB_USER=user
DB_PASSWORD=123456
LOG=true`)

	if err := ioutil.WriteFile(inputFile, d, 0644); err != nil {
		t.Error(err)
	}
	defer func() {
		os.Remove(inputFile)
	}()

	input, err := dotenv.Parse(bytes.NewReader(d), false)
	if err != nil {
		t.Error(err)
	}

	output, err := os.Create(resultsFile)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		output.Close()
		os.Remove(resultsFile)
	}()

	// Act
	os.Setenv("STASH_SECRETS", "multiple")
	os.Setenv("STASH_KMS_KEY_ID", "aws/secretsmanager")

	cmd := exec.Command(stash, "sync", inputFile, "-s", "secrets-manager", "-f", catalog)
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		t.Error(err)
	}

	defer func() {
		cmd := exec.Command(stash, "purge", "-f", catalog)
		cmd.Stdout = os.Stdout

		if err := cmd.Run(); err != nil {
			t.Error(err)
		}
	}()

	cmd = exec.Command(stash, "get", "-f", catalog)
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		t.Error(err)
	}

	// Assert
	f, err := os.Open(resultsFile)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	results, err := dotenv.Parse(f, false)
	if err != nil {
		t.Error(err)
	}

	if len(input) != len(results) {
		t.Error("incorrect number of results")
	}

	for k, v := range input {
		if results[k] != v {
			t.Errorf("INVALID %s: value(%s) != result(%s)", k, v, results[k])
		}
	}
}

func TestParameterStoreSyncEnv(t *testing.T) {
	// Arrange
	var catalog = fmt.Sprintf("%s_catalog.yml", t.Name())
	var inputFile = fmt.Sprintf("%s_input.env", t.Name())
	var resultsFile = fmt.Sprintf("%s_results.env", t.Name())

	d := []byte(`API_KEY=7fec6e3b-01bc-4b28-acc9-21028fe812b7
DB_USER=user
DB_PASSWORD=123456
LOG=true`)

	if err := ioutil.WriteFile(inputFile, d, 0644); err != nil {
		t.Error(err)
	}
	defer func() {
		os.Remove(inputFile)
	}()

	input, err := dotenv.Parse(bytes.NewReader(d), false)
	if err != nil {
		t.Error(err)
	}

	output, err := os.Create(resultsFile)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		output.Close()
		os.Remove(resultsFile)
	}()

	// Act
	os.Setenv("STASH_KMS_KEY_ID", "aws/ssm")

	cmd := exec.Command(stash, "sync", inputFile, "-s", "parameter-store", "-f", catalog)
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		t.Error(err)
	}

	defer func() {
		cmd := exec.Command(stash, "purge", "-f", catalog)
		cmd.Stdout = os.Stdout

		if err := cmd.Run(); err != nil {
			t.Error(err)
		}
	}()

	cmd = exec.Command(stash, "get", "-f", catalog)
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		t.Error(err)
	}

	// Assert
	f, err := os.Open(resultsFile)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	results, err := dotenv.Parse(f, false)
	if err != nil {
		t.Error(err)
	}

	if len(input) != len(results) {
		t.Error("incorrect number of results")
	}

	for k, v := range input {
		if results[k] != v {
			t.Errorf("INVALID %s: value(%s) != result(%s)", k, v, results[k])
		}
	}
}

func TestS3SyncEnv(t *testing.T) {
	// Arrange
	var catalog = fmt.Sprintf("%s_catalog.yml", t.Name())
	var inputFile = fmt.Sprintf("%s_input.env", t.Name())
	var resultsFile = fmt.Sprintf("%s_results.env", t.Name())

	d := []byte(`API_KEY=7fec6e3b-01bc-4b28-acc9-21028fe812b7
DB_USER=user
DB_PASSWORD=123456
LOG=true`)

	if err := ioutil.WriteFile(inputFile, d, 0644); err != nil {
		t.Error(err)
	}
	defer func() {
		os.Remove(inputFile)
	}()

	input, err := dotenv.Parse(bytes.NewReader(d), false)
	if err != nil {
		t.Error(err)
	}

	output, err := os.Create(resultsFile)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		output.Close()
		os.Remove(resultsFile)
	}()

	// Act
	os.Setenv("STASH_S3_BUCKET", "stash-cli-test")
	os.Setenv("STASH_KMS_KEY_ID", "aws/s3")

	cmd := exec.Command(stash, "sync", inputFile, "-s", "s3", "-f", catalog)
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		t.Error(err)
	}

	defer func() {
		cmd := exec.Command(stash, "purge", "-f", catalog)
		cmd.Stdout = os.Stdout

		if err := cmd.Run(); err != nil {
			t.Error(err)
		}
	}()

	cmd = exec.Command(stash, "get", "-f", catalog)
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		t.Error(err)
	}

	// Assert
	f, err := os.Open(resultsFile)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	results, err := dotenv.Parse(f, false)
	if err != nil {
		t.Error(err)
	}

	if len(input) != len(results) {
		t.Error("incorrect number of results")
	}

	for k, v := range input {
		if results[k] != v {
			t.Errorf("INVALID %s: value(%s) != result(%s)", k, v, results[k])
		}
	}
}
