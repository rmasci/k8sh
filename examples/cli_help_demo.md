# 🎯 Enhanced CLI Help System

k8sh now provides a comprehensive CLI help system that shows available commands and per-command help!

## 🚀 Enhanced Help Features

### **Main Help (`k8sh --help`)**

#### **Before (Basic):**
```
k8sh [command]

Available Commands:
  completion  Generate completion script
  help        Help about any command
  posix       Start POSIX-compliant shell

Flags:
  -h, --help   help for k8sh
```

#### **After (Comprehensive):**
```
k8sh is an OS-independent pseudo-shell for Kubernetes pods that works
without requiring any tools in target containers. Supports distroless,
scratch, alpine, debian, and ubuntu-based images.

🚀 QUICK START:
  k8sh                    # Start interactive shell
  k8sh posix              # Start POSIX-compliant shell
  k8sh --help             # Show this help
  k8sh [command] --help   # Get help on specific command

📁 FEATURES:
  • File operations: ls, cd, cat, vi, mkdir, rm, cp, mv, touch
  • Text processing: head, tail, grep, wc, sort
  • System info: ps, env, df, du, ip
  • Kubernetes: pods, use, namespace
  • Download files: download <src> <dst>
  • POSIX shell: pipelines, redirection, variables

⚙️  CONFIGURATION:
  Uses standard Kubernetes config at ~/.kube/config
  First-time setup: minikube start, gcloud init, or manual config

🎯 EXAMPLES:
  k8sh                    # Start shell and select pod
  k8sh posix              # Advanced POSIX features
  k8sh posix --help       # POSIX shell help

Usage:
  k8sh [flags]
  k8sh [command]

Available Commands:
  completion  Generate completion script
  help        Help about any command
  posix       Start POSIX-compliant shell
  version     Show version information

Flags:
  -h, --help   help for k8sh

Use "k8sh [command] --help" for more information about a command.
```

## 📋 Available Commands

### **1. `k8sh` (Default)**
- **Purpose**: Start interactive k8sh shell
- **Usage**: `k8sh`
- **Help**: `k8sh --help`

### **2. `k8sh posix`**
- **Purpose**: Start POSIX-compliant shell
- **Usage**: `k8sh posix`
- **Help**: `k8sh posix --help`

### **3. `k8sh version`**
- **Purpose**: Show version and build information
- **Usage**: `k8sh version`
- **Help**: `k8sh version --help`

### **4. `k8sh completion`**
- **Purpose**: Generate shell completion scripts
- **Usage**: `k8sh completion [bash|zsh|fish|powershell]`
- **Help**: `k8sh completion --help`

## 🎯 Per-Command Help

### **POSIX Shell Help (`k8sh posix --help`)**
```
Start a POSIX-compliant shell with full command parsing,
pipelines, redirection, and built-in commands.

🐚 POSIX SHELL FEATURES:
  • Command pipelines: cmd1 | cmd2 | cmd3
  • I/O redirection: >, >>, <, 2>, &>
  • Variable expansion: $VAR, ${VAR}
  • Command substitution: $(cmd)
  • Built-in commands: echo, printf, export, cd, pwd, help
  • Environment management: export, unset, readonly
  • Shell options and error handling

🎯 POSIX EXAMPLES:
  k8sh posix                    # Start POSIX shell
  k8sh posix --help             # Show this help
  
  # Inside POSIX shell:
  echo "Hello POSIX!"           # Basic command
  export MY_VAR=value           # Set variable
  echo $MY_VAR                  # Variable expansion
  cat file.txt | grep "error"   # Pipeline
  ls -la > files.txt            # Redirection

📚 POSIX COMPLIANCE:
  Full POSIX shell compliance with:
  • AST-based parsing and execution
  • Environment and variable management
  • Built-in command implementation
  • Pipeline and redirection support
  • Job control and process management

🚀 USAGE:
  k8sh posix                    # Start POSIX shell
  help                          # Show shell commands
  exit                          # Return to k8sh

The POSIX shell provides a complete Unix-like shell experience
in Kubernetes containers without requiring any tools in the target
container. Perfect for distroless and minimal environments!
```

### **Version Command (`k8sh version --help`)**
```
Display version information for k8sh including build details,
supported platforms, and feature status.

📊 VERSION INFORMATION:
  • Current version and build details
  • Supported platforms and architectures
  • Feature availability and status
  • Kubernetes compatibility information

🎯 USAGE:
  k8sh version                 # Show version info
  k8sh version --help          # Show this help

This command helps you verify your k8sh installation and
check which features are available in your build.
```

### **Completion Command (`k8sh completion --help`)**
```
Generate shell completion scripts for k8sh to enable
tab completion in your preferred shell.

🎯 USAGE:
  k8sh completion bash        # Generate bash completion
  k8sh completion zsh         # Generate zsh completion
  k8sh completion fish        # Generate fish completion
  k8sh completion powershell # Generate PowerShell completion

📦 INSTALLATION:
  # Bash
  k8sh completion bash > ~/.k8sh-completion.bash
  echo 'source ~/.k8sh-completion.bash' >> ~/.bashrc

  # Zsh
  k8sh completion zsh > ~/.k8sh-completion.zsh
  echo 'source ~/.k8sh-completion.zsh' >> ~/.zshrc

  # Fish
  k8sh completion fish > ~/.config/fish/completions/k8sh.fish

This enables tab completion for k8sh commands and options
in your terminal!
```

## 🎊 Version Information

### **Version Output (`k8sh version`)**
```
🐚 k8sh - Kubernetes Pseudo-Shell
=====================================
Version: 1.0.0
Build:   Cross-platform distribution
Platform: darwin/arm64
Go:      go1.25.7

🎯 SUPPORTED PLATFORMS:
  • macOS (Intel, Apple Silicon)
  • Linux (Intel, ARM)
  • Windows (Intel)

📁 FEATURES:
  ✅ File operations (ls, cd, cat, vi, etc.)
  ✅ Text processing (grep, wc, sort, etc.)
  ✅ System information (ps, env, df, du, ip)
  ✅ Kubernetes integration (pods, use, namespace)
  ✅ File download (download command)
  ✅ POSIX shell (pipelines, redirection, variables)
  ✅ Tab completion and history
  ✅ Cross-platform builds

⚙️  KUBERNETES:
  • Standard ~/.kube/config support
  • Multiple namespace support
  • Container selection
  • Ephemeral container support

📦 BUILT FOR DISTRIBUTION:
  • No container dependencies
  • Works in distroless containers
  • Single binary deployment
  • Professional shell experience

For more information: https://github.com/rmasci/k8sh
```

## 🔧 Shell Completion

### **Bash Completion Setup**
```bash
# Generate completion script
k8sh completion bash > ~/.k8sh-completion.bash

# Add to bashrc
echo 'source ~/.k8sh-completion.bash' >> ~/.bashrc

# Reload bashrc
source ~/.bashrc

# Now you can use tab completion!
k8sh [TAB]           # Shows: completion, help, posix, version
k8sh posix [TAB]     # Shows: --help
```

### **Zsh Completion Setup**
```bash
# Generate completion script
k8sh completion zsh > ~/.k8sh-completion.zsh

# Add to zshrc
echo 'source ~/.k8sh-completion.zsh' >> ~/.zshrc

# Reload zshrc
source ~/.zshrc
```

## 🎯 Help System Benefits

### **✅ User Experience Improvements:**
- **Clear Command Discovery**: Users can see all available commands
- **Per-Command Help**: Detailed help for each command
- **Usage Examples**: Practical examples for each feature
- **Professional Presentation**: Well-formatted, easy to read

### **✅ Developer Experience:**
- **Self-Documenting**: Commands explain themselves
- **Consistent Interface**: Standard help format across all commands
- **Installation Guidance**: Clear setup instructions
- **Feature Visibility**: Users discover all available features

### **✅ Professional CLI Standards:**
- **Cobra Integration**: Professional CLI framework
- **Shell Completion**: Tab completion for all commands
- **Version Information**: Build and platform details
- **Error Handling**: Graceful help on errors

## 🚀 Usage Examples

### **First-Time User Experience:**
```bash
# Discover available commands
k8sh --help

# Get help on specific command
k8sh posix --help

# Check version and platform
k8sh version

# Enable tab completion
k8sh completion bash > ~/.k8sh-completion.bash
source ~/.k8sh-completion.bash

# Start using with tab completion
k8sh [TAB]           # Shows all commands
k8sh posix [TAB]     # Shows posix options
```

### **Advanced User Workflow:**
```bash
# Check current build
k8sh version

# Get POSIX shell details
k8sh posix --help

# Generate completions for new shell
k8sh completion fish > ~/.config/fish/completions/k8sh.fish

# Start with full knowledge
k8sh --help  # Remind yourself of all options
```

## 🎉 **Professional CLI Experience Complete!**

The k8sh CLI now provides:
- ✅ **Comprehensive help system** with clear command discovery
- ✅ **Per-command help** with detailed explanations and examples
- ✅ **Version information** with build and platform details
- ✅ **Shell completion** for all major shells
- ✅ **Professional presentation** with consistent formatting
- ✅ **User-friendly examples** and setup instructions

**k8sh now provides a truly professional CLI experience that users will love!** 🚀
