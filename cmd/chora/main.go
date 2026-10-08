// Package main provides the `chora` CLI — a developer tool for Chora platform management.
//
// Commands:
//
//	chora auth login --api-key <key>  — authenticate and store credentials
//	chora auth logout                 — clear local credentials
//	chora atoms list                 — list LearningAtoms
//	chora atoms get <id>              — fetch a single atom by ID
//	chora atoms create --file <yaml>  — create atom from YAML/JSON definition
//	chora tenants info                — display current tenant details
//	chora familiars status            — display familiar status
//	chora health                      — check service health
//	chora events tail                 — tail event bus in real time
//	chora flags list                  — list feature flag overrides
//	chora flags set <code>            — toggle a feature flag override
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/apollo-chora/chora-cli/cmd/chora/internal/client"
	"github.com/apollo-chora/chora-cli/cmd/chora/internal/config"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// version is set at build time via ldflags.
var version = "dev"

const defaultGatewayURL = "http://localhost:8093"

func main() {
	rootCmd := &cobra.Command{
		Use:     "chora",
		Short:   "Chora CLI — developer tool for platform management",
		Long:    "The chora CLI provides developer and admin access to the Chora Bimodal Atomic Learning Ecosystem.",
		Version: version,
	}

	// -------------------------------------------------------------------------
	// auth subcommand
	// -------------------------------------------------------------------------
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
	}

	var apiKeyFlag string
	var gatewayURLFlag string

	authLoginCmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate against the Chora platform",
		RunE: func(cmd *cobra.Command, args []string) error {
			if apiKeyFlag == "" {
				return fmt.Errorf("--api-key is required")
			}
			gwURL := gatewayURLFlag
			if gwURL == "" {
				gwURL = defaultGatewayURL
			}

			store := config.NewFileStore(config.DefaultDir())
			cred := &config.Credentials{
				APIKey:     apiKeyFlag,
				GatewayURL: gwURL,
			}
			if err := store.Save(cred); err != nil {
				return fmt.Errorf("save credentials: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Authenticated. Credentials stored securely.\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Gateway: %s\n", gwURL)
			return nil
		},
	}
	authLoginCmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "Platform API key (sk_live_... or sk_test_...)")
	authLoginCmd.Flags().StringVar(&gatewayURLFlag, "gateway", "", "Gateway URL (default: "+defaultGatewayURL+")")

	authLogoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Clear local credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := config.NewFileStore(config.DefaultDir())
			if err := store.Clear(); err != nil {
				return fmt.Errorf("clear credentials: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Logged out. Credentials cleared.")
			return nil
		},
	}

	authCmd.AddCommand(authLoginCmd, authLogoutCmd)

	// -------------------------------------------------------------------------
	// atoms subcommand
	// -------------------------------------------------------------------------
	atomsCmd := &cobra.Command{
		Use:   "atoms",
		Short: "LearningAtom management commands",
	}

	atomsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List LearningAtoms for the current tenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := authenticatedClient()
			if err != nil {
				return err
			}
			resp, err := c.Get(context.Background(), "/api/v1/atoms")
			if err != nil {
				return fmt.Errorf("API call: %w", err)
			}
			defer resp.Body.Close()
			return printJSON(cmd.OutOrStdout(), resp)
		},
	}

	atomsGetCmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Fetch a single LearningAtom by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := authenticatedClient()
			if err != nil {
				return err
			}
			resp, err := c.Get(context.Background(), "/api/v1/atoms/"+args[0])
			if err != nil {
				return fmt.Errorf("API call: %w", err)
			}
			defer resp.Body.Close()
			return printJSON(cmd.OutOrStdout(), resp)
		},
	}

	var atomFileFlag string
	atomsCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a LearningAtom from a YAML/JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if atomFileFlag == "" {
				return fmt.Errorf("--file is required")
			}
			c, err := authenticatedClient()
			if err != nil {
				return err
			}
			body, err := os.ReadFile(atomFileFlag)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}
			jsonBody, err := yaml.YAMLToJSON(body)
			if err != nil {
				return fmt.Errorf("parse atom definition: %w", err)
			}
			resp, err := c.Post(context.Background(), "/api/atoms", jsonBody)
			if err != nil {
				return fmt.Errorf("API call: %w", err)
			}
			defer resp.Body.Close()

			if loc := resp.Header.Get("Location"); loc != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", loc)
			}
			return printJSON(cmd.OutOrStdout(), resp)
		},
	}
	atomsCreateCmd.Flags().StringVar(&atomFileFlag, "file", "", "Path to atom definition (YAML/JSON)")

	atomsCmd.AddCommand(atomsListCmd, atomsGetCmd, atomsCreateCmd)

	// -------------------------------------------------------------------------
	// tenants subcommand
	// -------------------------------------------------------------------------
	tenantsCmd := &cobra.Command{
		Use:   "tenants",
		Short: "Tenant management commands",
	}

	tenantsInfoCmd := &cobra.Command{
		Use:   "info",
		Short: "Display current tenant details",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := authenticatedClient()
			if err != nil {
				return err
			}
			resp, err := c.Get(context.Background(), "/api/tenants/me")
			if err != nil {
				return fmt.Errorf("API call: %w", err)
			}
			defer resp.Body.Close()
			return printJSON(cmd.OutOrStdout(), resp)
		},
	}

	tenantsCmd.AddCommand(tenantsInfoCmd)

	// -------------------------------------------------------------------------
	// familiars subcommand
	// -------------------------------------------------------------------------
	familiarsCmd := &cobra.Command{
		Use:   "familiars",
		Short: "Familiar companion commands",
	}

	familiarsStatusCmd := &cobra.Command{
		Use:   "status",
		Short: "Display Familiar status and stats",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := authenticatedClient()
			if err != nil {
				return err
			}
			resp, err := c.Get(context.Background(), "/api/v1/familiars/me/stats")
			if err != nil {
				return fmt.Errorf("API call: %w", err)
			}
			defer resp.Body.Close()
			return printJSON(cmd.OutOrStdout(), resp)
		},
	}

	familiarsCmd.AddCommand(familiarsStatusCmd)

	// -------------------------------------------------------------------------
	// events subcommand
	// -------------------------------------------------------------------------
	eventsCmd := &cobra.Command{
		Use:   "events",
		Short: "Event bus commands",
	}

	eventsTailCmd := &cobra.Command{
		Use:   "tail",
		Short: "Tail the event bus in real time",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "chora events tail — streaming not yet implemented (requires WebSocket)")
			return nil
		},
	}

	eventsCmd.AddCommand(eventsTailCmd)

	// -------------------------------------------------------------------------
	// flags subcommand
	// -------------------------------------------------------------------------
	flagsCmd := &cobra.Command{
		Use:   "flags",
		Short: "Feature flag management commands",
	}

	flagsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List feature flag overrides",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := authenticatedClient()
			if err != nil {
				return err
			}
			resp, err := c.Get(context.Background(), "/api/feature-flags")
			if err != nil {
				return fmt.Errorf("API call: %w", err)
			}
			defer resp.Body.Close()
			return printJSON(cmd.OutOrStdout(), resp)
		},
	}

	flagsSetCmd := &cobra.Command{
		Use:   "set [code]",
		Short: "Toggle a feature flag override",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := authenticatedClient()
			if err != nil {
				return err
			}
			body := []byte(fmt.Sprintf(`{"code":"%s","enabled":true}`, args[0]))
			resp, err := c.Put(context.Background(), "/api/v1/admin/feature-flags/"+args[0], body)
			if err != nil {
				return fmt.Errorf("API call: %w", err)
			}
			defer resp.Body.Close()
			return printJSON(cmd.OutOrStdout(), resp)
		},
	}

	flagsCmd.AddCommand(flagsListCmd, flagsSetCmd)

	// -------------------------------------------------------------------------
	// health subcommand
	// -------------------------------------------------------------------------
	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "Check service health status",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := authenticatedClient()
			if err != nil {
				return err
			}
			resp, err := c.Get(context.Background(), "/health")
			if err != nil {
				return fmt.Errorf("API call: %w", err)
			}
			defer resp.Body.Close()
			return printJSON(cmd.OutOrStdout(), resp)
		},
	}

	// -------------------------------------------------------------------------
	// Register top-level subcommands
	// -------------------------------------------------------------------------
	rootCmd.AddCommand(authCmd, atomsCmd, tenantsCmd, familiarsCmd, eventsCmd, flagsCmd, healthCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// authenticatedClient loads credentials and returns a configured API client.
func authenticatedClient() (*client.Client, error) {
	store := config.NewFileStore(config.DefaultDir())
	cred, err := store.Load()
	if err != nil {
		return nil, err
	}
	return client.New(cred.GatewayURL, cred.APIKey), nil
}

// printJSON reads the response body and pretty-prints it as JSON.
func printJSON(w io.Writer, resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		fmt.Fprintf(w, "Error (HTTP %d):\n", resp.StatusCode)
	}

	var pretty json.RawMessage
	if json.Unmarshal(body, &pretty) == nil {
		formatted, _ := json.MarshalIndent(pretty, "", "  ")
		fmt.Fprintln(w, string(formatted))
	} else {
		fmt.Fprintln(w, string(body))
	}
	return nil
}
