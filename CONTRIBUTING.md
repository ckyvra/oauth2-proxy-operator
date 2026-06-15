# Contributing

## Prerequisites

- Go 1.23+
- kubebuilder envtest binaries (`setup-envtest use 1.31.x`)
- Access to a Kubernetes cluster (Kind, minikube, or remote)

## Development

```bash
make build          # compile the operator
make test           # run integration tests
make run            # run locally against current kubeconfig
```

## Project structure

```
api/v1alpha1/       # CRD types
controllers/        # reconciliation logic
config/             # Kubernetes manifests (CRD, RBAC, samples)
```

## Pull request process

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make your changes
4. Run `make test` and ensure all tests pass
5. Commit with conventional commits: `feat:`, `fix:`, `ci:`, `docs:`, etc.
6. Push and open a Pull Request

## Conventional commits

Use [conventional commits](https://www.conventionalcommits.org/) for clear changelog generation:

- `feat:` — new feature
- `fix:` — bug fix
- `ci:` — CI/CD changes
- `docs:` — documentation
- `refactor:` — code restructuring
- `chore:` — maintenance tasks

## Code style

- Follow standard Go formatting (`gofmt`)
- No comments in code unless necessary for documentation
- All exported types and functions must have Go doc comments
