package cmd

import (
	"fmt"

	"github.com/ashimrai123/harbor-satellite-fleet/groundctl/pkg/client"
	"github.com/ashimrai123/harbor-satellite-fleet/groundctl/pkg/output"
	"github.com/spf13/cobra"
)

var satelliteCmd = &cobra.Command{
	Use:     "satellite",
	Aliases: []string{"sat"},
	Short:   "Manage satellites",
}

var (
	listActive bool
	listStale  bool
)

var satelliteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all satellites",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		if listActive && listStale {
			return fmt.Errorf("cannot specify both --active and --stale")
		}

		var sats []client.Satellite
		if listActive {
			sats, err = c.GetActiveSatellites()
		} else if listStale {
			sats, err = c.GetStaleSatellites()
		} else {
			sats, err = c.ListSatellites()
		}

		if err != nil {
			return fmt.Errorf("failed to list satellites: %w", err)
		}

		p := output.NewPrinter(outputFormat)
		headers := []string{"NAME", "ID", "LAST SEEN", "HEARTBEAT", "CREATED"}
		var rows [][]string
		for _, s := range sats {
			rows = append(rows, []string{
				s.Name,
				fmt.Sprintf("%d", s.ID),
				s.LastSeen.String(),
				s.HeartbeatInterval.Val(),
				s.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
		p.PrintTable(headers, rows)
		return nil
	},
}

var satelliteGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get satellite details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		sat, err := c.GetSatellite(args[0])
		if err != nil {
			return fmt.Errorf("failed to get satellite: %w", err)
		}

		p := output.NewPrinter(outputFormat)
		p.PrintJSON(sat)
		return nil
	},
}

var (
	registerName   string
	registerConfig string
	registerGroups []string
)

var satelliteRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new satellite",
	Example: `  groundctl satellite register --name my-sat --config default-config
  groundctl satellite register --name edge-1 --config prod --groups us-west,eu-central`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		req := client.RegisterSatelliteRequest{
			Name:       registerName,
			ConfigName: registerConfig,
			Groups:     registerGroups,
		}

		resp, err := c.RegisterSatellite(req)
		if err != nil {
			return fmt.Errorf("failed to register satellite: %w", err)
		}

		output.PrintSuccess("satellite %q registered", registerName)
		fmt.Printf("ZTR Token: %s\n", resp.Token)
		return nil
	},
}

var satelliteDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a satellite",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		if err := c.DeleteSatellite(args[0]); err != nil {
			return fmt.Errorf("failed to delete satellite: %w", err)
		}

		output.PrintSuccess("satellite %q deleted", args[0])
		return nil
	},
}

var satelliteStatusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Get satellite status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		status, err := c.GetSatelliteStatus(args[0])
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		p := output.NewPrinter(outputFormat)
		if p.Format == output.FormatJSON {
			p.PrintJSON(status)
		} else {
			headers := []string{"FIELD", "VALUE"}
			rows := [][]string{
				{"Activity", status.Activity},
				{"Image Count", status.ImageCount.String()},
				{"CPU", status.CPUPercent.Val()},
				{"Memory", status.MemoryUsedBytes.Bytes()},
				{"Storage", status.StorageUsedBytes.Bytes()},
				{"Last Sync", status.LastSyncDurationMs.Ms()},
				{"Reported At", status.ReportedAt.Format("2006-01-02 15:04:05")},
			}
			p.PrintTable(headers, rows)
		}
		return nil
	},
}

var satelliteCachedImagesCmd = &cobra.Command{
	Use:   "cached-images [name]",
	Short: "Get cached images on a satellite",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		images, err := c.GetCachedImages(args[0])
		if err != nil {
			return fmt.Errorf("failed to get cached images: %w", err)
		}

		p := output.NewPrinter(outputFormat)
		headers := []string{"ID", "REFERENCE", "SIZE", "CREATED"}
		var rows [][]string
		for _, img := range images {
			rows = append(rows, []string{
				fmt.Sprintf("%d", img.ID),
				img.Reference,
				fmt.Sprintf("%.1f MB", float64(img.SizeBytes)/(1024*1024)),
				img.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
		p.PrintTable(headers, rows)
		return nil
	},
}

func init() {
	satelliteListCmd.Flags().BoolVar(&listActive, "active", false, "List only recently active satellites")
	satelliteListCmd.Flags().BoolVar(&listStale, "stale", false, "List only stale satellites")

	satelliteRegisterCmd.Flags().StringVar(&registerName, "name", "", "Satellite name (required)")
	satelliteRegisterCmd.Flags().StringVar(&registerConfig, "config", "", "Config name (required)")
	satelliteRegisterCmd.Flags().StringSliceVar(&registerGroups, "groups", nil, "Comma-separated group names")
	_ = satelliteRegisterCmd.MarkFlagRequired("name")
	_ = satelliteRegisterCmd.MarkFlagRequired("config")

	satelliteCmd.AddCommand(satelliteListCmd)
	satelliteCmd.AddCommand(satelliteGetCmd)
	satelliteCmd.AddCommand(satelliteRegisterCmd)
	satelliteCmd.AddCommand(satelliteDeleteCmd)
	satelliteCmd.AddCommand(satelliteStatusCmd)
	satelliteCmd.AddCommand(satelliteCachedImagesCmd)
	rootCmd.AddCommand(satelliteCmd)
}
