package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ashimrai123/harbor-satellite-fleet/groundctl/pkg/client"
	"github.com/ashimrai123/harbor-satellite-fleet/groundctl/pkg/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var applyFile string

// FleetSpec is the top-level declarative fleet configuration.
type FleetSpec struct {
	APIVersion string       `yaml:"apiVersion" json:"apiVersion"`
	Kind       string       `yaml:"kind" json:"kind"`
	Metadata   FleetMeta    `yaml:"metadata" json:"metadata"`
	Spec       FleetDetails `yaml:"spec" json:"spec"`
}

// FleetMeta contains fleet metadata.
type FleetMeta struct {
	Name string `yaml:"name" json:"name"`
}

// FleetDetails contains the desired state of the fleet.
type FleetDetails struct {
	Config     *FleetConfig      `yaml:"config,omitempty" json:"config,omitempty"`
	Groups     []FleetGroup      `yaml:"groups,omitempty" json:"groups,omitempty"`
	Satellites []FleetSatellite  `yaml:"satellites,omitempty" json:"satellites,omitempty"`
}

// FleetConfig describes a desired config.
type FleetConfig struct {
	Name       string          `yaml:"name" json:"name"`
	AppConfig  json.RawMessage `yaml:"app_config,omitempty" json:"app_config,omitempty"`
	ZotConfig  json.RawMessage `yaml:"zot_config,omitempty" json:"zot_config,omitempty"`
}

// FleetGroup describes a desired group with its artifacts.
type FleetGroup struct {
	Name      string                `yaml:"name" json:"name"`
	Registry  string                `yaml:"registry,omitempty" json:"registry,omitempty"`
	Artifacts []client.GroupArtifact `yaml:"artifacts,omitempty" json:"artifacts,omitempty"`
}

// FleetSatellite describes a desired satellite.
type FleetSatellite struct {
	Name       string   `yaml:"name" json:"name"`
	Groups     []string `yaml:"groups,omitempty" json:"groups,omitempty"`
	ConfigName string   `yaml:"config,omitempty" json:"config,omitempty"`
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a declarative fleet configuration",
	Long: `Apply a YAML file describing the desired state of your satellite fleet.
This reconciles configs, groups, and satellites against Ground Control,
similar to 'kubectl apply -f'.`,
	Example: `  groundctl apply -f fleet.yaml
  groundctl apply -f edge-production.yaml --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if applyFile == "" {
			return fmt.Errorf("--file (-f) is required")
		}

		c, err := client.NewClientFromConfig()
		if err != nil {
			return err
		}

		spec, err := loadFleetSpec(applyFile)
		if err != nil {
			return fmt.Errorf("failed to load spec: %w", err)
		}

		if err := validateFleetSpec(spec); err != nil {
			return fmt.Errorf("invalid spec: %w", err)
		}

		return reconcile(c, spec)
	},
}

func loadFleetSpec(path string) (*FleetSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var spec FleetSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &spec, nil
}

func validateFleetSpec(spec *FleetSpec) error {
	if spec.APIVersion != "groundcontrol/v1" {
		return fmt.Errorf("unsupported apiVersion %q (expected groundcontrol/v1)", spec.APIVersion)
	}
	if spec.Kind != "SatelliteFleet" {
		return fmt.Errorf("unsupported kind %q (expected SatelliteFleet)", spec.Kind)
	}
	if spec.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	return nil
}

func reconcile(c *client.Client, spec *FleetSpec) error {
	fmt.Printf("Applying fleet %q...\n\n", spec.Metadata.Name)

	// Step 1: Reconcile config
	if spec.Spec.Config != nil {
		if err := reconcileConfig(c, spec.Spec.Config); err != nil {
			return err
		}
	}

	// Step 2: Reconcile groups
	for _, g := range spec.Spec.Groups {
		if err := reconcileGroup(c, g); err != nil {
			return err
		}
	}

	// Step 3: Reconcile satellites
	for _, s := range spec.Spec.Satellites {
		configName := ""
		if spec.Spec.Config != nil {
			configName = spec.Spec.Config.Name
		}
		if s.ConfigName != "" {
			configName = s.ConfigName
		}
		if err := reconcileSatellite(c, s, configName); err != nil {
			return err
		}
	}

	fmt.Printf("\nFleet %q applied successfully.\n", spec.Metadata.Name)
	return nil
}

func reconcileConfig(c *client.Client, cfg *FleetConfig) error {
	// Check if config already exists
	_, err := c.GetConfig(cfg.Name)
	if err == nil {
		fmt.Printf("  config/%s  unchanged\n", cfg.Name)
		return nil
	}

	// Build the config payload
	configData := map[string]interface{}{}
	if cfg.AppConfig != nil {
		var appCfg interface{}
		if err := json.Unmarshal(cfg.AppConfig, &appCfg); err == nil {
			configData["app_config"] = appCfg
		}
	}
	if cfg.ZotConfig != nil {
		var zotCfg interface{}
		if err := json.Unmarshal(cfg.ZotConfig, &zotCfg); err == nil {
			configData["zot_config"] = zotCfg
		}
	}

	rawConfig, _ := json.Marshal(configData)

	req := client.CreateConfigRequest{
		ConfigName: cfg.Name,
		ConfigData: rawConfig,
	}

	if err := c.CreateConfig(req); err != nil {
		output.PrintError("config/%s  failed: %v", cfg.Name, err)
		return err
	}

	fmt.Printf("  config/%s  created\n", cfg.Name)
	return nil
}

func reconcileGroup(c *client.Client, g FleetGroup) error {
	req := client.GroupSyncRequest{
		Group:     g.Name,
		Registry:  g.Registry,
		Artifacts: g.Artifacts,
	}

	_, err := c.SyncGroup(req)
	if err != nil {
		output.PrintError("group/%s  failed: %v", g.Name, err)
		return err
	}

	fmt.Printf("  group/%s  synced\n", g.Name)
	return nil
}

func reconcileSatellite(c *client.Client, s FleetSatellite, defaultConfig string) error {
	// Check if satellite already exists
	_, err := c.GetSatellite(s.Name)
	if err == nil {
		fmt.Printf("  satellite/%s  unchanged\n", s.Name)
		return nil
	}

	configName := defaultConfig
	if s.ConfigName != "" {
		configName = s.ConfigName
	}

	groups := s.Groups

	req := client.RegisterSatelliteRequest{
		Name:       s.Name,
		ConfigName: configName,
		Groups:     groups,
	}

	resp, err := c.RegisterSatellite(req)
	if err != nil {
		output.PrintError("satellite/%s  failed: %v", s.Name, err)
		return err
	}

	fmt.Printf("  satellite/%s  created (token: %s...)\n", s.Name, truncate(resp.Token, 12))
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func init() {
	applyCmd.Flags().StringVarP(&applyFile, "file", "f", "", "Path to fleet YAML file")
	_ = applyCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(applyCmd)
}
