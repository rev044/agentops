---
name: languages
description: >
  Use when: "Python", "Go", "Golang", "Rust", "Java", "TypeScript", "shell script",
  "bash", "async/await", "goroutines", "ownership", "virtual threads", "type system",
  "decorators", "generators", "channels", "interfaces", "Spring Boot", "POSIX".
version: 1.0.0
author: "AgentOps Team"
license: "MIT"
---

# Languages Skill

Production patterns for Python, Go, Rust, Java, TypeScript, and Shell scripting.

## Quick Reference

| Language | Key Patterns | When to Use |
|----------|--------------|-------------|
| **Python** | decorators, generators, async/await, pytest | Data processing, APIs, scripting |
| **Go** | goroutines, channels, interfaces | Concurrent systems, CLI tools |
| **Rust** | ownership, lifetimes, async tokio | Systems programming, performance |
| **Java** | virtual threads, Spring Boot 3.x | Enterprise, microservices |
| **TypeScript** | advanced types, generics, strict mode | Frontend, Node.js backends |
| **Shell** | set -euo pipefail, POSIX compliance | Automation, CI/CD scripts |

---

## Python

### Focus Areas
- Advanced features: decorators, metaclasses, descriptors
- Async/await and concurrent programming
- Performance optimization and profiling
- Design patterns and SOLID principles
- Testing with pytest, mocking, fixtures
- Type hints and static analysis (mypy, ruff)

### Approach
1. Pythonic code - follow PEP 8 and Python idioms
2. Prefer composition over inheritance
3. Use generators for memory efficiency
4. Comprehensive error handling with custom exceptions
5. Test coverage above 90% with edge cases

### Output
- Clean Python code with type hints
- Unit tests with pytest and fixtures
- Performance benchmarks for critical paths
- Documentation with docstrings and examples

---

## Go

### Focus Areas
- Concurrency patterns (goroutines, channels, sync)
- Interface design and composition
- Error handling patterns (no exceptions)
- Performance optimization and profiling
- Testing with table-driven tests
- Module management and versioning

### Approach
1. Keep it simple - Go's strength is simplicity
2. Use interfaces for abstraction
3. Handle errors explicitly
4. Prefer channels over shared memory
5. Write table-driven tests

### Output
- Idiomatic Go code following effective Go
- Comprehensive tests with benchmarks
- Clear error handling patterns
- Well-documented public APIs

---

## Rust

### Focus Areas
- Ownership, borrowing, lifetimes
- Async programming with Tokio
- Error handling with Result/Option
- Trait design and generics
- Memory safety patterns
- FFI and systems integration

### Approach
1. Embrace the borrow checker
2. Use Result for recoverable errors
3. Prefer iterators over loops
4. Design for zero-cost abstractions
5. Test with cargo test and miri

### Output
- Safe, performant Rust code
- Proper error types with thiserror
- Async code with tokio
- Documentation with rustdoc examples

---

## Java

### Focus Areas
- Modern Java 21+ features (virtual threads, records, pattern matching)
- Spring Boot 3.x and reactive programming
- Microservices patterns
- Testing with JUnit 5 and Mockito
- GraalVM native compilation
- Structured concurrency

### Approach
1. Use virtual threads for I/O-bound work
2. Prefer records for data carriers
3. Use sealed classes for domain modeling
4. Leverage pattern matching in switch
5. Test with JUnit 5 parameterized tests

### Output
- Modern Java code using latest features
- Spring Boot applications with proper configuration
- Comprehensive tests with high coverage
- Native-image compatible code when needed

---

## TypeScript

### Focus Areas
- Advanced type system (generics, conditional types, mapped types)
- Strict mode configuration
- Type inference optimization
- Utility types and type guards
- Integration with React, Node.js
- Build tooling (tsc, esbuild, vite)

### Approach
1. Enable strict mode always
2. Prefer type inference over explicit types
3. Use discriminated unions for state
4. Create type guards for runtime checks
5. Avoid any - use unknown instead

### Output
- Type-safe code with minimal type assertions
- Well-typed APIs with proper generics
- Custom utility types when needed
- tsconfig optimized for project needs

---

## Shell Scripting

### Focus Areas
- Bash 4.0+ features
- POSIX compliance for portability
- Error handling with set -euo pipefail
- Shellcheck compliance
- Process management
- CI/CD script patterns

### Approach
1. Always use `set -euo pipefail`
2. Quote all variables: `"$var"`
3. Use shellcheck for validation
4. Prefer functions over scripts
5. Handle signals properly

### Output
- Robust shell scripts with proper error handling
- Shellcheck-clean code
- Cross-platform compatibility notes
- Documentation with usage examples

---

## Common Patterns

### Error Handling

| Language | Pattern |
|----------|---------|
| Python | `try/except` with custom exceptions |
| Go | `if err != nil { return err }` |
| Rust | `Result<T, E>` with `?` operator |
| Java | `try-catch` or `Optional<T>` |
| TypeScript | `try-catch` with typed errors |
| Shell | `set -e` + explicit checks |

### Testing

| Language | Framework | Pattern |
|----------|-----------|---------|
| Python | pytest | fixtures, parametrize |
| Go | testing | table-driven tests |
| Rust | cargo test | doc tests, integration tests |
| Java | JUnit 5 | @ParameterizedTest |
| TypeScript | Jest/Vitest | describe/it blocks |
| Shell | bats | test functions |

### Async/Concurrency

| Language | Mechanism |
|----------|-----------|
| Python | asyncio, await |
| Go | goroutines, channels |
| Rust | tokio, async/await |
| Java | virtual threads, CompletableFuture |
| TypeScript | Promise, async/await |
| Shell | background jobs (&), wait |
