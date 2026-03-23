# k8sh - Kubernetes Pseudo-Shell

An OS-independent pseudo-shell for Kubernetes pods that works without requiring any tools in target containers. Supports distroless, scratch, alpine, debian, and ubuntu-based images.

## Features

- **Universal Compatibility**: Works on all container types including distroless images
- **Built-in Unix Tools**: 40+ essential Unix commands (cat, ls, grep, etc.)
- **Vi Editor**: Full-featured vi clone for editing files
- **Zero Container Dependencies**: Nothing installed in target containers
- **Persistent Shell**: Interactive shell experience, not one-off debugging

## Quick Start

```bash
# Build
go build -o k8sh ./cmd/k8sh

# Run
./k8sh
```

## Architecture

```
k8sh/
├── cmd/k8sh/          # CLI interface
├── pkg/k8s/           # Kubernetes client
├── pkg/ops/           # File operations
├── pkg/shell/         # Shell interface
├── pkg/editor/        # Vi editor
└── go.mod
```

## Development Status

- ✅ Project structure
- ✅ Basic Kubernetes client
- ✅ CLI interface with cobra
- ✅ Interactive shell with Bubble Tea
- ✅ File operations implementation
- ✅ Vi editor implementation
- 🚧 Unix tools implementation

## Commands

Currently implemented:
- `help` - Show help
- `exit` - Exit shell
- `pwd` - Print current directory
- `ls` - List directory
- `cd` - Change directory
- `cat` - Display file contents
- `vi/vim` - Edit files with vi editor
- `pods` - List available pods
- `use <pod>` - Select pod and container
- `namespace <name>` - Set namespace
- `clear` - Clear screen

## Vi Editor Features

The built-in vi editor supports:
- **Navigation**: h,j,k,l + arrow keys
- **Modes**: Normal, Insert, Visual, Command
- **Editing**: i,a,o,O (insert), x,dd (delete)
- **Yank/Paste**: yy (yank line), p,P (paste)
- **Search**: /pattern (forward), ?pattern (backward)
- **Navigation**: n,N (next/previous match)
- **Commands**: :w (save), :q (quit), :wq (save and quit)
- **Replace**: :s/old/new/g (global replace)

## Dependencies

- Go 1.25+
- Kubernetes cluster access
- kubectl configuration
