# Code Quality Rules

## Comments (HARD)

- No comments inside code. No inline `//` or `#` explanations, no trailing comments, no section dividers, no commented-out code, no "this fixes X" or "why this is correct" notes.
- The only accepted comments are documentation comments that immediately precede a declaration and explain it: godoc for Go, JSDoc/TSDoc for TypeScript and JavaScript, Javadoc for Java, docstrings for Python, NatSpec for Solidity.
- Doc comments describe what the function or type does, its parameters, return values, and error conditions — never implementation narration.
- If code seems to need an inline comment to be understood, rewrite the code (rename, extract a function) instead of commenting it.

## Hard limits

- Functions: max 100 lines
- Cyclomatic complexity: max 8
- Function parameters: max 5 (use a config struct if more are needed)
- File length: max 500 lines (split into focused files if exceeded)

## Naming

- Go: exported names describe the action (`ProcessDocument`, not `DocumentProcessor`)
- TypeScript: PascalCase for components, camelCase for hooks and utilities

## Structure

- One concern per file
- No package-level mutable state except loggers and metrics registries
- Dependency injection over package-level init functions

## Enforcement

Apply these limits as the floor for every change. When a function passes 100 lines, split it before continuing. When parameters pass 5, move them into a dedicated request or config struct.
