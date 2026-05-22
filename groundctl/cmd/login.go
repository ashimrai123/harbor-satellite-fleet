package cmd

import (
	"fmt"

	"github.com/ashimrai123/harbor-satellite-fleet/groundctl/pkg/client"
	"github.com/ashimrai123/harbor-satellite-fleet/groundctl/pkg/output"
	"github.com/spf13/cobra"
)

var (
	loginURL      string
	loginUser     string
	loginPassword string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Ground Control",
	Long:  "Login to a Ground Control instance and store the auth token for subsequent commands.",
	Example: `  groundctl login --url http://localhost:8080 --user admin --password secret
  groundctl login -u http://192.168.1.100:8080 -U admin -P mypassword`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if loginURL == "" {
			return fmt.Errorf("--url is required")
		}
		if loginUser == "" || loginPassword == "" {
			return fmt.Errorf("--user and --password are required")
		}

		c := client.NewClient(loginURL, "")
		if err := c.Login(loginUser, loginPassword); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		output.PrintSuccess("logged in to %s as %s", loginURL, loginUser)
		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from Ground Control",
	Long:  "Invalidate the current auth token on the server and remove the local credentials.",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			// If there's no config at all, just clean up silently
			_ = client.DeleteConfig()
			output.PrintSuccess("already logged out")
			return nil
		}

		if err := c.Logout(); err != nil {
			// Even if the server call fails (e.g. expired token), remove local creds
			_ = client.DeleteConfig()
			output.PrintSuccess("local credentials removed (server returned: %v)", err)
			return nil
		}

		output.PrintSuccess("logged out")
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVarP(&loginURL, "url", "u", "", "Ground Control URL (e.g. http://localhost:8080)")
	loginCmd.Flags().StringVarP(&loginUser, "user", "U", "", "Username")
	loginCmd.Flags().StringVarP(&loginPassword, "password", "P", "", "Password")
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
}
