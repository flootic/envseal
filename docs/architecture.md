# EnvSeal Architecture

EnvSeal is a decentralized, local-first CLI tool for managing encrypted environment secrets.
It uses Git as the single source of truth, with no dependency on external secret management services.

## Project Structure

```
envseal/
├── cmd/
│   └── envseal-cli/                # Entrypoint for the CLI application
├── internal/
│   └── cli/
│       ├── audit/                  # Audit logging implementation
│       ├── commands/               # Implementations of all CLI commands (init, set, exec, users add, etc.)
│       ├── config/                 # Manifest and identity file handling
│       ├── crypto/                 # Age encryption and key management
│       └── p2p/                    # Peer-to-peer pairing implementation
├── pkg/
│   └── filesystem/                 # Atomic file writing utility
```

## Core Concepts

### Two-Tier Encryption

EnvSeal uses an envelope encryption scheme with two layers:

1. **Data Encryption Key (DEK)** — a 32-byte random key that encrypts all secret values using ChaCha20-Poly1305.
2. **Per-user key wrapping** — the DEK is encrypted separately for each authorized user using Age (X25519) asymmetric encryption.

This means adding or removing a user only requires re-encrypting the DEK, not every secret.

### File Layout

Two files live at the project root (both typically committed to Git):

**`envseal.yaml`** (manifest) — stores the project name and the list of authorized users with their public keys.

**`secrets.enc.yaml`** (encrypted secrets) — stores:

- `_envseal:` metadata block containing per-recipient wrapped DEKs
- `secrets:` map of key-value pairs where each value is `ENC[age,chacha20,<base64>]`

### Identity

Each user has a local identity stored at `~/.envseal/identity` containing their X25519 private key (file mode `0600`).
The matching public key is registered in the project manifest.

## Component Overview

```
┌─────────────────────────────────────────────────┐
│                  CLI Commands                   │
│         (internal/cli/commands/*.go)            │
│                                                 │
│  init · set · unset · exec · print · rekey      │
│  status · doctor · whoami · join · users · hook │
└──────────────────┬──────────────────────────────┘
                   │  uses Deps interfaces
                   ▼
┌──────────────────────────────────────────────────┐
│               Dependency Layer                   │
│              (commands/deps.go)                  │
│                                                  │
│  IdentityStore    SecretsStore    ManifestStore  │
└───────┬──────────────┬───────────────┬───────────┘
        │              │               │
        ▼              ▼               ▼
┌─────────────┐ ┌────────────┐ ┌─────────────────┐
│   crypto/   │ │  config/   │ │  config/        │
│  age.go     │ │ secret_    │ │ manifest.go     │
│  keygen.go  │ │ file.go    │ │                 │
│  code.go    │ │ identity.go│ │                 │
└─────────────┘ └──────┬─────┘ └────────┬────────┘
                       │                │
                       ▼                ▼
               ┌──────────────────────────────┐
               │        pkg/filesystem        │
               │        AtomicWriteFile       │
               └──────────────────────────────┘
```

## Key Flows

### Initialization (`envseal init`)

```
1. Generate X25519 identity (if none exists)      → ~/.envseal/identity
2. Create manifest with the current user           → envseal.yaml
3. Generate a random DEK
4. Wrap DEK for the current user's public key
5. Write metadata + empty secrets map              → secrets.enc.yaml
```

### Setting a Secret (`envseal set KEY=VALUE`)

```
1. Load identity from disk
2. Load secrets.enc.yaml
3. Unlock (try to decrypt DEK using identity)
4. Encrypt VALUE with ChaCha20-Poly1305 using DEK
5. Store as ENC[age,chacha20,<base64>] under secrets map
6. Lock (zero DEK from memory)
7. Atomic write to disk
```

### Running a Command (`envseal exec -- <cmd>`)

```
1. Load and unlock secrets
2. Decrypt all secret values
3. Inject as environment variables into child process
4. Lock and zero DEK
```

### Adding a User (`envseal users add`)

```
1. Add public key to manifest
2. Unlock the secret file
3. Re-encrypt DEK for the updated recipient list (rekey)
4. Save both files
```

### P2P Pairing (`envseal join` + `envseal users add --p2p`)

```
New User                          Admin
────────                          ─────
join (generates 6-digit code)     users add --p2p <code>
  → broadcasts pubkey via mDNS      → discovers pubkey via mDNS
  → listens for TCP ack             → adds to manifest + rekeys
                                    → sends TCP ack
  ← receives ack, confirms
```

## Security Properties

- **Encryption at rest** — secrets are always encrypted on disk with ChaCha20-Poly1305.
- **Per-user access control** — each user's copy of the DEK is independently wrapped with Age.
- **Memory hygiene** — the DEK is zeroed from memory immediately after use via `Lock()`.
- **Atomic writes** — file writes use temp-file-then-rename to avoid corruption.
- **Strict permissions** — identity files are written with `0600`; secret files with `0600`.
- **Audit trail** — every CLI invocation is logged to `~/.envseal/audit.log`.
- **No secret leakage** — error messages are generic to avoid leaking cryptographic details.
