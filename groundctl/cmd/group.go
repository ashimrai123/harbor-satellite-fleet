package cmd

import (
	"fmt"

	"github.com/ashimrai123/harbor-satellite-fleet/groundctl/pkg/client"
	"github.com/ashimrai123/harbor-satellite-fleet/groundctl/pkg/output"
	"github.com/spf13/cobra"
)

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage groups",
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		groups, err := c.ListGroups()
		if err != nil {
			return fmt.Errorf("failed to list groups: %w", err)
		}

		p := output.NewPrinter(outputFormat)
		headers := []string{"NAME", "ID", "REGISTRY", "PROJECTS", "CREATED"}
		var rows [][]string
		for _, g := range groups {
			projects := "-"
			if len(g.Projects) > 0 {
				projects = fmt.Sprintf("%v", g.Projects)
			}
			rows = append(rows, []string{
				g.GroupName,
				fmt.Sprintf("%d", g.ID),
				g.RegistryURL,
				projects,
				g.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
		p.PrintTable(headers, rows)
		return nil
	},
}

var groupGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get group details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		group, err := c.GetGroup(args[0])
		if err != nil {
			return fmt.Errorf("failed to get group: %w", err)
		}

		p := output.NewPrinter(outputFormat)
		p.PrintJSON(group)
		return nil
	},
}

var groupDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		if err := c.DeleteGroup(args[0]); err != nil {
			return fmt.Errorf("failed to delete group: %w", err)
		}

		output.PrintSuccess("group %q deleted", args[0])
		return nil
	},
}

var groupAddSatelliteCmd = &cobra.Command{
	Use:   "add-satellite [group] [satellite]",
	Short: "Add a satellite to a group",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		group := args[0]
		satellite := args[1]

		if err := c.AddSatelliteToGroup(satellite, group); err != nil {
			return fmt.Errorf("failed to add satellite to group: %w", err)
		}

		output.PrintSuccess("satellite %q added to group %q", satellite, group)
		return nil
	},
}

var groupRemoveSatelliteCmd = &cobra.Command{
	Use:   "remove-satellite [group] [satellite]",
	Short: "Remove a satellite from a group",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		group := args[0]
		satellite := args[1]

		if err := c.RemoveSatelliteFromGroup(satellite, group); err != nil {
			return fmt.Errorf("failed to remove satellite from group: %w", err)
		}

		output.PrintSuccess("satellite %q removed from group %q", satellite, group)
		return nil
	},
}

func init() {
	groupCmd.AddCommand(groupListCmd)
	groupCmd.AddCommand(groupGetCmd)
	groupCmd.AddCommand(groupDeleteCmd)
	groupCmd.AddCommand(groupAddSatelliteCmd)
	groupCmd.AddCommand(groupRemoveSatelliteCmd)
	rootCmd.AddCommand(groupCmd)
}
