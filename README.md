# harbor-satellite-fleet

Monorepo for [Harbor Satellite](https://github.com/container-registry/harbor-satellite) fleet management tooling, covering the [Ground Control CLI and Kubernetes Fleet Operator](https://github.com/container-registry/harbor-satellite/issues/375).

## Components

### groundctl (CLI)

A command-line tool for managing Harbor Satellite fleets through the Ground Control API.

**Commands:**

| Command | Description |
|---|---|
| `groundctl login` | Authenticate with Ground Control |
| `groundctl satellite list` | List all registered satellites |
| `groundctl satellite get <name>` | Get satellite details |
| `groundctl satellite register` | Register a new satellite |
| `groundctl satellite delete <name>` | Delete a satellite |
| `groundctl satellite status <name>` | Get satellite health status |
| `groundctl group list` | List all groups |
| `groundctl group get <name>` | Get group details |
| `groundctl config list` | List all configs |
| `groundctl config get <name>` | Get config details |
| `groundctl apply -f fleet.yaml` | Declarative fleet reconciliation |
| `groundctl version` | Print version |

**Quick Start:**

```bash
cd groundctl
go build -o groundctl .

# Login
./groundctl login --url http://localhost:8080 --user admin --password <pw>

# List resources
./groundctl satellite list
./groundctl group list
./groundctl config list

# Declarative fleet management (GitOps-style)
./groundctl apply -f examples/fleet.yaml
```

**Example fleet.yaml:**

```yaml
apiVersion: groundcontrol/v1
kind: SatelliteFleet
metadata:
  name: edge-production
spec:
  config:
    name: prod-config
  groups:
    - name: us-west-edge
      artifacts:
        - repository: library/nginx
          tag: ["latest"]
  satellites:
    - name: edge-sat-1
      groups:
        - us-west-edge
```

**Design:**

The `pkg/client/` package is a reusable Go client for the Ground Control API, designed to be shared between the CLI and the Kubernetes operator. The operator's reconcile loop will call the same `client.RegisterSatellite()`, `client.SyncGroup()`, and `client.SetSatelliteConfig()` methods that the CLI uses.

### operator (planned)

Kubernetes operator managing Satellite instances as custom resources, enabling GitOps-driven fleet management through ArgoCD or Flux.

## Tech Stack

- Go 1.23+
- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [kubebuilder](https://book.kubebuilder.io/) - CRD + controller scaffolding (operator)
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) - Kubernetes operator framework (operator)

## Related

- [Harbor Satellite](https://github.com/container-registry/harbor-satellite)
- [OpenAPI Spec PR #451](https://github.com/container-registry/harbor-satellite/pull/451)
- [LFX Issue #375](https://github.com/container-registry/harbor-satellite/issues/375)

## License

This project is licensed under the [Apache License 2.0](LICENSE).
