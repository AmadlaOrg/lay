<img src=".assets/lay.jpg" alt="Electronics photo" style="width: 400px;" align="right">

# lay

Lay helps with installing or compiling applications.

There are many ways to install and compile software. Lay abstracts the myriad
of package managers, build systems, and container runtimes behind a single CLI
so you don't have to look up per-distro package names or build flags.

## Features

- **Binary install** from GitHub, GitLab, or Codeberg releases with automatic
  platform detection, version pinning, direct URL support, and JAR application
  support
- **Compile from source** with auto-detected build system (autotools, cmake,
  meson, make, cargo, go)
- **System package management** across distros (apt, dnf, yum, pacman, zypper,
  apk, nix, snap, flatpak, dpkg, rpm)
- **Container runtime** abstraction (Docker, Podman)

## Install

```bash
go install github.com/AmadlaOrg/lay@latest
```

## Usage

### Binary install

```bash
# Latest release from GitHub (default)
lay binary install sharkdp/fd

# Specific version
lay binary install sharkdp/fd@v9.0.0

# Explicit GitHub prefix
lay binary install github:sharkdp/fd

# GitLab release
lay binary install gitlab:user/project

# Codeberg release
lay binary install codeberg:user/repo@v1.0.0

# Custom install directory
lay binary install --to /opt/bin BurntSushi/ripgrep

# Direct URL (any host)
lay binary install https://example.com/tool-linux-amd64.tar.gz

# JAR application (creates a launcher script wrapping java -jar)
lay binary install ~/Downloads/tika-app-3.2.3.jar

# JAR with custom command name
lay binary install ~/Downloads/tika-app-3.2.3.jar --name tika

# Forge release with only JAR assets (auto-detected)
lay binary install apache/tika
```

### Binary compile

```bash
# Compile from local source (auto-detect build system)
lay binary compile .

# Specify build system explicitly
lay binary compile --build-system golang .

# Compile from GitHub repo
lay binary compile user/repo
```

### Package management

```bash
# Install packages
lay package install curl wget

# Search for packages
lay package search nodejs

# Override package manager
lay package --manager dnf install vim
```

### Container

```bash
# Commands are passed through to the detected runtime
lay container run --rm alpine echo hello

# Override runtime
lay container --runtime docker ps
```

### Settings

```bash
# Show current configuration
lay settings
```

### JAR applications

When installing a `.jar` file (locally or from a forge release), lay:

1. Verifies that Java is installed and available in `PATH`
2. Copies the JAR to `~/.local/share/lay/jars/`
3. Creates a launcher script in the target directory that wraps `java -jar`

The command name is derived automatically by stripping the `.jar` extension and
version suffix (e.g. `tika-app-3.2.3.jar` becomes `tika-app`). Use `--name` to
override.

For forge sources, if no native binary matches the current platform, lay falls
back to JAR assets automatically. It excludes `-sources`, `-javadoc`, and
`-tests` JARs and prefers fat/standalone/app JARs.

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `LAY_BINARY_PATH` | Binary install directory | `~/.local/bin` |
| `LAY_PACKAGE_MANAGER` | Override auto-detected package manager | auto-detect |
| `LAY_CONTAINER_RUNTIME` | Override auto-detected container runtime | auto-detect |

Command-line flags take precedence over environment variables, which take
precedence over auto-detection.

## Supported Platforms

### Linux package managers

apt, dnf, yum, pacman, zypper, apk, nix, snap, flatpak, dpkg, rpm

### Windows package managers

Chocolatey, winget, scoop

### Build systems

| System | Detected by |
|---|---|
| autotools | `configure`, `configure.ac` |
| cmake | `CMakeLists.txt` |
| meson | `meson.build` |
| cargo | `Cargo.toml` |
| golang | `go.mod` |
| makefile | `Makefile`, `makefile`, `GNUmakefile` |

### Container runtimes

Docker, Podman (preferred when both are available)

## License

[MIT](./LICENSE)

---

Made in Québec 🏴󠁣󠁡󠁱󠁣󠁿, Canada 🇨🇦!
