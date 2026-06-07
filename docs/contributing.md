# Contributing to AILINTER

We welcome contributions! Here's how to get started.

> **This page is under construction.** For now, see [CONTRIBUTING.md](https://github.com/ailinter/ailinter/blob/main/CONTRIBUTING.md) in the repository.

## Development Setup

```bash
git clone https://github.com/ailinter/ailinter.git
cd ailinter
make build
make test
```

## Code Standards

- All AI-generated or modified code must pass `ailinter check` before commit
- Minimum quality score: 90 for core repo, 80 for everything else
- Run `make lint` before pushing (go vet + staticcheck)
- Write tests for new features

## PR Workflow

1. Fork the repo
2. Create a feature branch
3. Make your changes
4. Run `make test && make lint && ailinter check .`
5. Open a PR with a clear description
