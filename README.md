# EnvSeal CLI

[![ci](https://github.com/flootic/envseal/actions/workflows/ci.yaml/badge.svg)](https://github.com/flootic/envseal/actions/workflows/ci.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/envseal/cli.svg)](https://pkg.go.dev/github.com/envseal/cli)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

**EnvSeal** is a decentralized, *local-first*, cloud-agnostic secret management CLI tool. Unlike traditional solutions, it does not store your secrets on a central server. Instead, it uses asymmetric cryptography to store encrypted secrets directly in your Git repository (Single Source of Truth) and uses a P2P protocol for the secure distribution of access keys among developers.

## Features

- **Decentralized Storage**: Secrets are stored in your Git repository, eliminating the need for a central server.
- **Local-First**: Work with your secrets offline and sync changes when you're back online.
- **Asymmetric Cryptography**: Securely encrypt and decrypt secrets using public/private key pairs.
- **P2P Key Distribution**: Share access keys securely among team members without relying on a central authority.
- **Git Integration**: Seamlessly integrates with Git workflows, making it easy to manage secrets alongside your code.
- **Cross-Platform**: Available on Windows, macOS, and Linux.

## Installation

You can install EnvSeal CLI using Go:

```bash
go install github.com/envseal/cli@latest
```

Alternatively, you can download pre-built binaries from the [releases page](https://github.com/envseal/cli/releases)
or install via Homebrew on macOS:

```bash
brew install envseal/tap/envseal
```

## Usage

After installation, you can start using EnvSeal CLI with the following commands:

```bash
envseal init                            # Initialize EnvSeal in your Git repository
envseal set <key>=<value>               # Set a new secret
envseal unset <key>                     # Remove a secret
envseal users add <user> <public_key>   # Add a user with their public key
envseal users remove <user>             # Remove a user
envseal rekey [--rotate]                # Encrypt secrets and update access permissions
envseal exec -- <command>               # Execute a command with secrets injected into the environment
envseal doctor                          # Check the integrity of your EnvSeal setup
```

Print all commands with `envseal --help` and get detailed help for each command with `envseal <command> --help`.

## Contributing

Contributions are welcome! Please read the [contributing guidelines](CONTRIBUTING.md) for more information on how to get started.

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.
