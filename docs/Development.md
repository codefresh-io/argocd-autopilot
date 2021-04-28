# Development
This guide is meant for developers who want to contribute or debug `argocd-autopilot`.

### Adding a feature:
<>
1. Fork the repository
2. Run `make local` to build the project
### Linting:
We are using https://github.com/golangci/golangci-lint as our linter, you can integrate golangci-lint with the following IDEs:

- vscode: make sure `GOPATH` is setup correctly and run `make lint` this will download `golangci-lint` if it was not already installed on your machine. Then add the following to your `settings.json`:
```json
"go.lintTool": "golangci-lint",
"go.lintFlags": [
    "--fast"
],
```
