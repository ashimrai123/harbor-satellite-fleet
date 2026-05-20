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

var configDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		if err := c.DeleteConfig(args[0]); err != nil {
			return fmt.Errorf("failed to delete config: %w", err)
		}

		output.PrintSuccess("config %q deleted", args[0])
		return nil
	},
}

var configAssignToSatelliteCmd = &cobra.Command{
	Use:   "assign-to-satellite [config] [satellite]",
	Short: "Assign a configuration to a satellite",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		configName := args[0]
		satellite := args[1]

		if err := c.SetSatelliteConfig(satellite, configName); err != nil {
			return fmt.Errorf("failed to assign config to satellite: %w", err)
		}

		output.PrintSuccess("config %q assigned to satellite %q", configName, satellite)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configAssignToSatelliteCmd)
	rootCmd.AddCommand(configCmd)
}
