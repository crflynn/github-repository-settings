package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github-repository-settings/internal/github"
	"github-repository-settings/internal/settings"
)

func main() {
	if err := run(context.Background(), os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "github-repository-settings:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, getenv func(string) string, output io.Writer) error {
	dryRunInput := strings.ToLower(strings.TrimSpace(getenv("INPUT_DRY-RUN")))
	dryRun := false
	switch dryRunInput {
	case "", "false":
	case "true":
		dryRun = true
	default:
		return fmt.Errorf("invalid dry-run input %q: expected true or false", dryRunInput)
	}
	manifestPath := getenv("INPUT_SETTINGS-FILE")
	if manifestPath == "" {
		manifestPath = ".github/repository-settings.yml"
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("read settings manifest %q: %w", manifestPath, err)
	}
	desired, err := settings.Decode(data)
	if err != nil {
		return err
	}
	var client *github.Client
	if dryRun {
		client, err = github.NewDryRunClient(getenv("GITHUB_API_URL"), getenv("INPUT_TOKEN"), getenv("GITHUB_REPOSITORY"))
	} else {
		client, err = github.NewClient(getenv("GITHUB_API_URL"), getenv("INPUT_TOKEN"), getenv("GITHUB_REPOSITORY"), nil)
	}
	if err != nil {
		return err
	}
	if err := client.Apply(ctx, desired); err != nil {
		return err
	}
	if dryRun {
		if _, err := fmt.Fprintf(output, "Dry-run plan for %s (no write requests were sent):\n", getenv("GITHUB_REPOSITORY")); err != nil {
			return err
		}
		operations := client.PlannedOperations()
		if len(operations) == 0 {
			_, err := fmt.Fprintln(output, "No write requests would be made.")
			return err
		}
		for _, operation := range operations {
			if _, err := fmt.Fprintf(output, "Would %s %s\n", operation.Action(), operation.Path); err != nil {
				return err
			}
			if len(operation.Details) == 0 {
				continue
			}
			var details any
			if err := json.Unmarshal(operation.Details, &details); err != nil {
				continue
			}
			details = redactSensitiveValues(details)
			formatted, err := json.MarshalIndent(details, "", "  ")
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(output, "  Requested values:\n  %s\n", strings.ReplaceAll(string(formatted), "\n", "\n  ")); err != nil {
				return err
			}
		}
		return nil
	}
	_, err = fmt.Fprintf(output, "Reconciled repository settings for %s\n", getenv("GITHUB_REPOSITORY"))
	return err
}

func redactSensitiveValues(value any) any {
	switch value := value.(type) {
	case map[string]any:
		for key, nested := range value {
			normalized := strings.ToLower(strings.ReplaceAll(key, "-", "_"))
			if strings.Contains(normalized, "secret") || strings.Contains(normalized, "token") ||
				strings.Contains(normalized, "password") || strings.Contains(normalized, "authorization") ||
				strings.Contains(normalized, "private_key") {
				value[key] = "[REDACTED]"
				continue
			}
			value[key] = redactSensitiveValues(nested)
		}
		return value
	case []any:
		for index, nested := range value {
			value[index] = redactSensitiveValues(nested)
		}
		return value
	default:
		return value
	}
}
