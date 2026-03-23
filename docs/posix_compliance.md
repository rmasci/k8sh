# POSIX Compliance for k8sh

## Overview

Adding POSIX compliance to k8sh would make it compatible with standard shell scripts and tools, improving its utility for automation and existing workflows. This document outlines what POSIX compliance means for k8sh and how to implement it.

## Current State Analysis

### Current Command Set
k8sh currently implements:
- **Basic Commands**: pwd, ls, cd, cat, mkdir, rm, cp, mv, touch
- **Text Processing**: head, tail, grep, wc, sort
- **System Info**: ps, env, df, du
- **K8s-specific**: pods, use, namespace, ip, clear
- **Editor**: vi/vim

### POSIX Gaps
Missing POSIX utilities:
- **Shell Built-ins**: echo, printf, export, unset, readonly, alias, unalias
- **Flow Control**: if, for, while, case, function definitions
- **Redirection**: >, >>, <, |, 2>, etc.
- **Process Management**: &, jobs, fg, bg, kill, wait
- **Text Utilities**: cut, tr, uniq, find, xargs
- **System Utilities**: date, whoami, id, uname, sleep
- **Environment**: set, env, export, PATH handling

## Implementation Plan

### Phase 1: Core POSIX Built-ins
1. **echo/printf** - Output formatting
2. **export/unset** - Environment variable management
3. **alias/unalias** - Command aliases
4. **set** - Shell options and variables

### Phase 2: Redirection and Pipes
1. **I/O Redirection** - >, >>, <, 2>, &>
2. **Pipes** - | operator
3. **Here Documents** - <<EOF syntax
4. **Command Substitution** - $(cmd) and `cmd`

### Phase 3: Flow Control
1. **if/then/else/fi** - Conditional execution
2. **for/do/done** - Loop constructs
3. **while/do/done** - Conditional loops
4. **case/esac** - Pattern matching

### Phase 4: Advanced POSIX Features
1. **Functions** - User-defined functions
2. **Arrays** - Indexed and associative arrays
3. **Parameter Expansion** - ${var:-default}, ${var#prefix}, etc.
4. **Process Management** - Background jobs, job control

## Technical Implementation

### Parser Architecture
```go
type POSIXParser struct {
    lexer    *Lexer
    ast      *AST
    env      *Environment
    aliases  map[string]string
    functions map[string]*Function
}

type ASTNode interface {
    Execute(ctx context.Context, env *Environment) (string, error)
}

type CommandNode struct {
    Name   string
    Args   []ASTNode
    Redirs []Redirection
}
```

### Environment Management
```go
type Environment struct {
    vars     map[string]string
    exports  map[string]bool
    readonly map[string]bool
    functions map[string]*Function
    aliases  map[string]string
    options  ShellOptions
}
```

### Redirection Support
```go
type Redirection struct {
    Type     RedirType  // >, >>, <, 2>, etc.
    Source   int        // File descriptor
    Target   string     // File path or descriptor
}
```

## Benefits of POSIX Compliance

1. **Script Compatibility** - Run existing shell scripts unchanged
2. **Tool Integration** - Work with standard Unix tools and pipelines
3. **User Familiarity** - Standard shell behavior and syntax
4. **Automation** - Better support for CI/CD and automation
5. **Portability** - Scripts work across different POSIX shells

## Implementation Challenges

### 1. Container Context
- POSIX assumes local filesystem access
- Need to translate to container operations
- Some POSIX features may not map cleanly to K8s

### 2. State Management
- Shell state across commands
- Environment variable persistence
- Process tracking in container context

### 3. Performance
- Parsing overhead vs current simple switch
- Network latency for container operations
- State synchronization

## Compatibility Strategy

### Option 1: Full POSIX Mode
- Complete POSIX shell implementation
- Separate command mode: `k8sh --posix`
- Maintains backward compatibility

### Option 2: POSIX Extensions
- Add POSIX features as extensions
- Gradual migration path
- Maintains current architecture

### Option 3: Hybrid Approach
- Core POSIX built-ins + K8s extensions
- Seamless integration
- Best of both worlds

## Recommended Approach

**Option 3: Hybrid Approach** is recommended because:

1. **Incremental** - Can add features gradually
2. **Compatible** - Maintains existing functionality
3. **Powerful** - Combines POSIX with K8s features
4. **User-friendly** - No mode switching required

## Next Steps

1. **Research** - Study POSIX specification (IEEE 1003.1)
2. **Design** - Create detailed architecture
3. **Prototype** - Implement core built-ins
4. **Test** - POSIX compliance test suite
5. **Iterate** - Add features based on feedback

## Testing Strategy

### POSIX Test Suite
- Use existing POSIX test suites (like those from OpenGroup)
- Test conformance to specification
- Validate behavior against bash/sh

### K8s Integration Tests
- Ensure POSIX features work with container operations
- Test error handling in K8s context
- Validate performance impact

## Conclusion

Adding POSIX compliance would significantly enhance k8sh's utility and make it more suitable for production use. The hybrid approach provides the best balance of compatibility, functionality, and implementation complexity.
