# harbor-satellite-fleet

> **Note:** This repository is a functional proof-of-concept created for the CNCF LFX 2026 Mentorship application: *Harbor Satellite Ground Control CLI & Kubernetes Fleet Operator*.

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

### operator (Kubernetes Fleet Operator)

A Kubernetes operator that manages Harbor Satellite instances as custom resources, reconciled against Ground Control.

**Satellite CR example:**

```yaml
apiVersion: fleet.harbor.io/v1alpha1
kind: Satellite
metadata:
  name: edge-node-1
  namespace: harbor-system
spec:
  groundControlURL: http://ground-control:8080
  configName: prod-config
  groups:
    - us-west-edge
  secretRef: gc-admin-credentials
```

**What the operator does:**
- Watches `Satellite` custom resources
- Registers the satellite with Ground Control on CR creation, stores ZTR token in a Secret
- Syncs satellite to specified groups and assigns the config
- Removes the satellite from Ground Control when the CR is deleted (finalizer pattern)
- Reports status conditions: `Registered`, `GroupsSynced`, `ConfigAssigned`, `Ready`
- Re-queues every 60s for health monitoring

**Deploy with Helm:**
```bash
helm install satellite-operator operator/charts/satellite-operator \
  --namespace harbor-system --create-namespace

kubectl apply -f operator/config/samples/fleet_v1alpha1_satellite.yaml

kubectl get satellites -n harbor-system
# NAME          PHASE   GC-ID   CONFIG        AGE
# edge-node-1   Ready   13      prod-config   2m
```

## Tech Stack

- Go 1.23+
- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [kubebuilder v4](https://book.kubebuilder.io/) - CRD + controller scaffolding
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) - Kubernetes operator framework
- [Helm](https://helm.sh/) - Operator packaging and deployment

## Related

- [Harbor Satellite](https://github.com/container-registry/harbor-satellite)
- [OpenAPI Spec PR #451](https://github.com/container-registry/harbor-satellite/pull/451)
- [LFX Issue #375](https://github.com/container-registry/harbor-satellite/issues/375)

## License

This project is licensed under the [Apache License 2.0](LICENSE).
