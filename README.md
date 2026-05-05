# fleet-ctl

A minimal Kubernetes fleet operator and CLI, built to demonstrate the core patterns used in [Harbor Satellite's Ground Control CLI and Fleet Operator](https://github.com/container-registry/harbor-satellite/issues/375).

## What This Covers

- **Cobra CLI** - `register`, `list`, `inspect`, `delete`, `apply -f` (GitOps declarative mode)
- **OpenAPI client** - Generated Go client from an OpenAPI spec using `oapi-codegen`
- **Kubernetes CRD** - `Fleet` custom resource with spec, status, and conditions
- **controller-runtime** - Reconciliation loop syncing CRDs against control plane state
- **Validating Webhook** - Input validation on CRD creation/update
- **E2E Testing** - Integration tests using `envtest` / `kind`

## Tech Stack

- Go 1.23+
- [kubebuilder](https://book.kubebuilder.io/) - CRD + controller scaffolding
- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) - OpenAPI to Go client
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) - Kubernetes operator framework

## License

This project is licensed under the [Apache License 2.0](LICENSE).
