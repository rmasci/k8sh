#!/bin/bash

# POSIX Compliance Examples for k8sh
# This script demonstrates POSIX shell features that work with k8sh

echo "=== k8sh POSIX Compliance Examples ==="
echo

# Basic POSIX builtins
echo "1. Basic POSIX Builtins:"
echo "Hello World" | cat
printf "Formatted output: %s %d\n" "test" 42
echo

# Variable operations
echo "2. Variable Operations:"
export MY_VAR="hello"
echo "MY_VAR = $MY_VAR"
readonly READONLY_VAR="cannot change"
echo "READONLY_VAR = $READONLY_VAR"
echo

# Command substitution
echo "3. Command Substitution:"
CURRENT_DIR=$(pwd)
echo "Current directory: $CURRENT_DIR"
echo "Date: $(date)"
echo

# Redirection
echo "4. I/O Redirection:"
echo "This goes to stdout"
echo "Error message" >&2
echo "Content" > /tmp/test_file.txt
echo "Appended content" >> /tmp/test_file.txt
echo "File contents:"
cat /tmp/test_file.txt
echo

# Pipes and pipelines
echo "5. Pipes and Pipelines:"
echo -e "apple\nbanana\ncherry\napple" | sort | uniq
echo "One two three" | wc -w
echo

# Control structures (when implemented)
echo "6. Control Structures:"
# if command
if [ -n "$MY_VAR" ]; then
    echo "MY_VAR is set to: $MY_VAR"
fi

# for loop
for fruit in apple banana cherry; do
    echo "Fruit: $fruit"
done

# while loop
count=0
while [ $count -lt 3 ]; do
    echo "Count: $count"
    count=$((count + 1))
done
echo

# Functions
echo "7. Functions:"
my_function() {
    echo "Function called with args: $1 $2"
    return 0
}

my_function "arg1" "arg2"
echo "Function return code: $?"
echo

# Aliases
echo "8. Aliases:"
alias ll="ls -la"
alias grep="grep --color=auto"
echo "Aliases defined: ll, grep"
echo

# Environment and shell options
echo "9. Shell Options:"
set -x  # Enable tracing
echo "This command will be traced"
set +x  # Disable tracing
echo

# Job control (when implemented)
echo "10. Job Control:"
sleep 1 &
echo "Background job started"
jobs
wait
echo

# Parameter expansion
echo "11. Parameter Expansion:"
FILE="/path/to/file.txt"
echo "Filename: ${FILE##*/}"
echo "Directory: ${FILE%/*}"
echo "Extension: ${FILE##*.}"
echo

# Array operations (when implemented)
echo "12. Arrays:"
my_array=(apple banana cherry)
echo "First element: ${my_array[0]}"
echo "All elements: ${my_array[@]}"
echo

# Signal handling (when implemented)
echo "13. Signal Handling:"
trap 'echo "Interrupted!"' INT
echo "Press Ctrl-C to test (or wait 2 seconds)"
sleep 2 || echo "No interrupt received"
trap - INT
echo

# Here documents
echo "14. Here Documents:"
cat << EOF
This is a here document.
It can span multiple lines.
Variables work here: $MY_VAR
EOF
echo

# Process substitution (when implemented)
echo "15. Process Substitution:"
# diff <(sort file1.txt) <(sort file2.txt)
echo "Process substitution not yet implemented in k8sh"
echo

# Advanced patterns
echo "16. Advanced Patterns:"
case $MY_VAR in
    hello)
        echo "Variable is 'hello'"
        ;;
    world)
        echo "Variable is 'world'"
        ;;
    *)
        echo "Variable is '$MY_VAR'"
        ;;
esac
echo

# k8sh-specific extensions
echo "17. k8sh-specific POSIX Extensions:"
echo "K8SH_POD: $K8SH_POD"
echo "K8SH_CONTAINER: $K8SH_CONTAINER"
echo "K8SH_NAMESPACE: $K8SH_NAMESPACE"
echo

# Cleanup
echo "18. Cleanup:"
rm -f /tmp/test_file.txt
unset MY_VAR
unalias ll grep 2>/dev/null || true

echo "=== POSIX Examples Complete ==="
