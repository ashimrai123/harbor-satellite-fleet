package cmd

import (
	"fmt"

	"github.com/ashimrai123/harbor-satellite-fleet/groundctl/pkg/client"
	"github.com/ashimrai123/harbor-satellite-fleet/groundctl/pkg/output"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage satellite configurations",
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configs",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		configs, err := c.ListConfigs()
		if err != nil {
			return fmt.Errorf("failed to list configs: %w", err)
		}

		p := output.NewPrinter(outputFormat)
		headers := []string{"NAME", "ID", "REGISTRY", "CREATED"}
		var rows [][]string
		for _, cfg := range configs {
			rows = append(rows, []string{
				cfg.ConfigName,
				fmt.Sprintf("%d", cfg.ID),
				cfg.RegistryURL,
				cfg.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
		p.PrintTable(headers, rows)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get config details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		cfg, err := c.GetConfig(args[0])
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}

		p := output.NewPrinter(outputFormat)
		p.PrintJSON(cfg)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
	rootCmd.AddCommand(configCmd)
}
