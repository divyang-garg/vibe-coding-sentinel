---
description: Shell Script Standards.
globs: ["**/*.sh", "**/*.bash", "**/*.zsh", "**/*.ps1", "**/*.bat"]
---
# Shell Script Standards
- Error Handling: Always use "set -e" and "set -u"
- Variable Quoting: Always quote variable expansions: "$VAR"
- Temporary Files: Use mktemp, never hardcode /tmp paths
- File Operations: Never use "rm -rf" with variables or user input
- Command Injection: Never use eval with user input
- Paths: Avoid hardcoded absolute paths
- Security: Validate all inputs before use
