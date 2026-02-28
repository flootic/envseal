# EnvSeal

[![ci](https://github.com/flootic/envseal/actions/workflows/ci.yaml/badge.svg)](https://github.com/flootic/envseal/actions/workflows/ci.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/flootic/envseal.svg)](https://pkg.go.dev/github.com/flootic/envseal)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

**EnvSeal** is a decentralized, *local-first*, cloud-agnostic secret management tool. Unlike traditional solutions, it does not store your secrets on a central server. Instead, it uses asymmetric cryptography to store encrypted secrets directly in your Git repository (Single Source of Truth) and uses a P2P protocol for the secure distribution of access keys among developers.

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
go install github.com/flootic/envseal/cmd/envseal-cli@latest
```

Alternatively, you can download pre-built binaries from the [releases page](https://github.com/flootic/envseal/releases).

## Example

1. Initialize EnvSeal in your Git repository:

    ```bash
    cd my-project
    envseal-cli init
    ```

2. Set a new secret:

    ```bash
    envseal-cli set DATABASE_URL=postgres://user:password@localhost:5432/mydb
    ```

3. Add a user with their public key:

    ```bash
    envseal-cli users add alice <public_key>
    ```

4. Rekey the secrets to update access permissions:

    ```bash
    envseal-cli rekey --rotate
    ```

5. Print all secrets (for debugging purposes):

    ```bash
    envseal-cli exec -- printenv | grep SECRET_
    ```

## Usage

After installation, you can start using EnvSeal CLI with the following commands:

```bash
envseal-cli init                            # Initialize EnvSeal in your Git repository
envseal-cli set <key>=<value>               # Set a new secret
envseal-cli unset <key>                     # Remove a secret
envseal-cli users add <user> <public_key>   # Add a user with their public key
envseal-cli users remove <user>             # Remove a user
envseal-cli join                            # Request access to a project using p2p (mDNS) and 6-digit code.
envseal-cli rekey [--rotate]                # Encrypt secrets and update access permissions
envseal-cli exec -- <command>               # Execute a command with secrets injected into the environment
envseal-cli doctor                          # Check the integrity of your EnvSeal setup
envseal-cli print                           # Print all secrets in plaintext (for debugging purposes)
envseal-cli whoami                          # Print the public key of the current identity
```

Print all commands with `envseal-cli --help` and get detailed help for each command with `envseal-cli <command> --help`.

## Contributing

Contributions are welcome! Please read the [contributing guidelines](CONTRIBUTING.md) for more information on how to get started.

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.
