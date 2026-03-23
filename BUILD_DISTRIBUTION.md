# 🏗 k8sh Build Distribution Guide

## Overview

k8sh now supports **cross-platform builds** with proper distribution in a `releases/` directory instead of `bin/`. This makes it easier to distribute and use k8sh across different platforms.

## 🎯 Build Targets

### Standard Build (Current Platform)
```bash
make build          # Builds for current platform to releases/k8sh
```

### Cross-Platform Build (All Platforms)
```bash
make build-all      # Builds for all supported platforms
```

### Platform Support Matrix

| Platform | Architecture | Binary Name | Status |
|----------|--------------|-------------|--------|
| macOS (Darwin) | amd64 | k8sh-darwin-amd64 | ✅ |
| macOS (Darwin) | arm64 | k8sh-darwin-arm64 | ✅ |
| Linux | amd64 | k8sh-linux-amd64 | ✅ |
| Linux | arm64 | k8sh-linux-arm64 | ✅ |
| Windows | amd64 | k8sh-windows-amd64.exe | ✅ |

## 📁 Directory Structure

### Before (bin/ directory)
```
k8sh/
├── bin/
│   ├── k8sh                    # Single platform binary
│   └── k8sh-windows-amd64.exe  # Windows binary (if built)
```

### After (releases/ directory)
```
k8sh/
├── releases/
│   ├── k8sh                     # Current platform binary
│   ├── k8sh-darwin-amd64        # macOS Intel
│   ├── k8sh-darwin-arm64        # macOS Apple Silicon
│   ├── k8sh-linux-amd64          # Linux Intel/AMD
│   ├── k8sh-linux-arm64          # Linux ARM
│   └── k8sh-windows-amd64.exe     # Windows
```

## 🚀 Usage Examples

### Development Build
```bash
# Build for your current platform
make build

# Output
✅ Built: releases/k8sh
```

### Release Build
```bash
# Build for all platforms
make build-all

# Output
🏗 Building k8sh for all platforms...
==================================
Building for macOS (amd64)...
Building for macOS (arm64)...
Building for Linux (amd64)...
Building for Linux (arm64)...
Building for Windows (amd64)...
✅ Build complete! Files in releases/:
total 747176
drwxr-xr-x@  8 rmasci  staff       256 Mar 23 10:17 .
drwxr-xr-x@ 17 rmasci  staff       544 Mar 23 10:15 ..
-rwxr-xr-x@  1 rmasci  staff  63493282 Mar 23 10:16 k8sh
-rwxr-xr-x@  1 rmasci  staff  66110240 Mar 23 10:16 k8sh-darwin-amd64
-rwxr-xr-x@  1 rmasci  staff  63968748 Mar 23 10:16 k8sh-darwin-arm64
-rwxr-xr-x@  1 rmasci  staf  60752729 Mar 23 10:16 k8sh-linux-amd64
-rwxr-xr-x@  1 rmasci  staff  64717312 Mar 23 10:16 k8sh-linux-arm64
-rwxr-xr-x@  1 rmasci  staf  64717312 Mar 23 10:17 k8sh-windows-amd64.exe

📦 Distribution ready!
```

### Clean Build Artifacts
```bash
make clean

# Output
🧹 Cleaning build artifacts...
✅ Clean complete!
```

## 🛠️ Installation

### For Users
```bash
# Download the appropriate binary for your platform
# Example for macOS Intel:
curl -L -o k8sh https://github.com/rmasci/k8sh/releases/latest/download/k8sh-darwin-amd64
chmod +x k8sh
./k8sh
```

### For Developers
```bash
# Clone and build
git clone https://github.com/rmasci/k8sh.git
cd k8sh
make build-all

# Or build for specific platform
GOOS=linux GOARCH=amd64 make build
```

## 🎯 Distribution Benefits

### ✅ **Cross-Platform Support**
- **macOS**: Intel and Apple Silicon (both architectures)
- **Linux**: x86_64 and ARM64 (both architectures)  
- **Windows**: x86_64 (primary Windows architecture)

### ✅ **Professional Distribution**
- **Consistent naming** with platform-architecture suffix
- **Single command** builds all platforms
- **Clean separation** from development artifacts
- **Easy deployment** with proper binary names

### ✅ **Developer Experience**
- **Fast builds** with parallel compilation
- **Clear output** with progress indicators
- **Clean targets** for artifact management
- **Standard Go tooling** with cross-compilation

## 🔄 Migration from bin/ to releases/

If you have existing `bin/k8sh` builds:

```bash
# Clean old builds
make clean

# Build new releases
make build-all

# Or move existing binary
mv bin/k8sh releases/k8sh-$(go env GOOS)-$(go env GOARCH)
```

## 📋 Make Target Reference

| Target | Description |
|--------|-------------|
| `make build` | Build for current platform to releases/ |
| `make build-all` | Build for all supported platforms to releases/ |
| `make clean` | Remove all build artifacts from releases/ |
| `make test` | Run all tests |
| `make fmt` | Format Go code |
| `make vet` | Run go vet |
| `make lint` | Run linter |

## 🎉 Ready for Distribution

The k8sh project now supports **professional-grade cross-platform distribution** with:
- ✅ **Multiple architecture support**
- ✅ **Consistent binary naming**
- ✅ **Automated build process**
- ✅ **Clean artifact management**
- ✅ **Developer-friendly workflow**

This makes k8sh ready for:
- 📦 **GitHub Releases** with proper binary attachments
- 🐳 **Homebrew** formula creation
- 📦 **Package managers** (apt, yum, etc.)
- 🐳 **Docker images** with multi-architecture support
- ☁️ **Cloud deployment** across all platforms

**The future of k8sh distribution is here!** 🚀
