# Rust Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-02-09
**Purpose:** Canonical Rust standards for vibe skill validation

---

## Table of Contents

1. [Project Structure](#project-structure)
2. [Cargo Configuration](#cargo-configuration)
3. [Code Formatting](#code-formatting)
4. [Ownership & Borrowing](#ownership--borrowing)
5. [Error Handling Patterns](#error-handling-patterns)
6. [Trait & Type System Design](#trait--type-system-design)
7. [Concurrency Patterns](#concurrency-patterns)
8. [Unsafe Code](#unsafe-code)
9. [Testing Patterns](#testing-patterns)
10. [Code Quality Metrics & Anti-Patterns](#code-quality-metrics--anti-patterns)

---

## Project Structure

### ✅ **Standard Crate Layout**

```
my-project/
├── Cargo.toml             # Package manifest
├── Cargo.lock             # Dependency lock (commit for binaries, .gitignore for libs)
├── src/
│   ├── lib.rs             # Library root (public API surface)
│   ├── main.rs            # Binary entrypoint (or use src/bin/)
│   ├── bin/
│   │   ├── server.rs      # Additional binary
│   │   └── cli.rs         # Additional binary
│   ├── models/
│   │   └── mod.rs         # Domain types
│   └── handlers/
│       └── mod.rs         # Request handlers
├── tests/                 # Integration tests (each file is a separate crate)
│   ├── integration.rs
│   └── e2e.rs
├── examples/              # Runnable examples (`cargo run --example`)
│   └── basic_usage.rs
├── benches/               # Benchmarks (`cargo bench`)
│   └── throughput.rs
└── build.rs               # Build script (optional)
```

**Principles:**
- ✅ `src/lib.rs` defines the public API; `src/main.rs` consumes it
- ✅ `src/bin/` for multiple binaries within one crate
- ✅ `tests/` for integration tests (compiled as separate crates)
- ✅ `examples/` for documentation-as-code
- ✅ Commit `Cargo.lock` for binaries, omit for libraries

### ⚠️ **Module Organization**

```rust
// GOOD - Explicit re-exports in lib.rs
pub mod config;
pub mod handlers;
pub mod models;

pub use config::Config;
pub use models::AppError;

// BAD - Deep nesting with no re-exports
// Forces users to write: my_crate::handlers::http::v1::webhook::process
```

**Module Size Thresholds:**

| File Size | Status | Action |
|-----------|--------|--------|
| < 300 lines | ✅ Excellent | Maintain |
| 300-500 lines | ✅ Acceptable | Monitor |
| 500-800 lines | ⚠️ Warning | Consider splitting |
| 800+ lines | ❌ Critical | Split into submodules |

---

## Cargo Configuration

### ✅ **Dependency Management**

```toml
[package]
name = "my-service"
version = "0.1.0"
edition = "2021"
rust-version = "1.75"       # MSRV - minimum supported Rust version

[dependencies]
tokio = { version = "1", features = ["full"] }
serde = { version = "1", features = ["derive"] }
serde_json = "1"
thiserror = "2"
tracing = "0.1"

[dev-dependencies]
tokio = { version = "1", features = ["test-util", "macros"] }
proptest = "1"
criterion = { version = "0.5", features = ["html_reports"] }

[build-dependencies]
prost-build = "0.13"        # Only if needed at build time
```

**Requirements:**
- ✅ Pin `edition` and `rust-version` for reproducibility
- ✅ Use feature flags to minimize compile-time and binary size
- ✅ Separate `dev-dependencies` from production deps
- ✅ Never use wildcard versions (`*`)

### ✅ **Feature Flags**

```toml
[features]
default = ["json"]
json = ["dep:serde_json"]
tls = ["dep:rustls"]
full = ["json", "tls"]

# Optional dependencies gated by feature
[dependencies]
serde_json = { version = "1", optional = true }
rustls = { version = "0.23", optional = true }
```

**Why This Matters:**
- Users opt into functionality they need
- Reduces compile time and binary size
- Avoids pulling transitive dependencies unnecessarily

### ✅ **Profile Configuration**

```toml
[profile.release]
lto = true          # Link-time optimization
codegen-units = 1   # Single codegen unit for max optimization
strip = true        # Strip debug symbols from binary
panic = "abort"     # Smaller binary, no unwinding

[profile.dev]
opt-level = 0       # Fast compile
debug = true        # Full debug info

[profile.test]
opt-level = 1       # Slight optimization for faster test runs
```

### ✅ **Workspace Configuration**

```toml
# Root Cargo.toml
[workspace]
members = [
    "crates/core",
    "crates/api",
    "crates/cli",
]

[workspace.dependencies]
serde = { version = "1", features = ["derive"] }
tokio = { version = "1", features = ["full"] }

# In member Cargo.toml
[dependencies]
serde = { workspace = true }
tokio = { workspace = true }
```

**Benefits:**
- Single lockfile across all crates
- Unified dependency versions
- `cargo test --workspace` runs all tests

---

## Code Formatting

### ✅ **rustfmt Configuration**

```toml
# rustfmt.toml
edition = "2021"
max_width = 100
tab_spaces = 4
use_field_init_shorthand = true
use_try_shorthand = true
imports_granularity = "Module"
group_imports = "StdExternalCrate"
```

**Requirements:**
- ✅ Run `cargo fmt --check` in CI (zero-tolerance for formatting drift)
- ✅ `group_imports = "StdExternalCrate"` enforces import order: std, external, crate-local

### ✅ **Import Grouping**

```rust
// GOOD - Grouped and ordered
use std::collections::HashMap;
use std::sync::Arc;

use serde::{Deserialize, Serialize};
use tokio::sync::Mutex;

use crate::config::Config;
use crate::models::AppError;

// BAD - Unorganized imports
use crate::config::Config;
use std::collections::HashMap;
use serde::Serialize;
use std::sync::Arc;
use crate::models::AppError;
use tokio::sync::Mutex;
```

### ✅ **Naming Conventions**

| Item | Convention | Example |
|------|-----------|---------|
| Types, Traits | `UpperCamelCase` | `HttpClient`, `Serialize` |
| Functions, Methods | `snake_case` | `process_request` |
| Local Variables | `snake_case` | `retry_count` |
| Constants | `SCREAMING_SNAKE_CASE` | `MAX_RETRIES` |
| Modules | `snake_case` | `error_handling` |
| Type Parameters | Single uppercase or `CamelCase` | `T`, `Item` |
| Lifetimes | Short lowercase | `'a`, `'ctx` |
| Crate Names | `kebab-case` (Cargo.toml) | `my-service` |
| Feature Flags | `kebab-case` | `full-json` |

### ⚠️ **Naming Anti-Patterns**

| Pattern | Problem | Instead |
|---------|---------|---------|
| `get_` prefix on getters | Redundant in Rust | `fn name(&self)` not `fn get_name(&self)` |
| `FooStruct` suffix | Redundant | `Foo` |
| `IFoo` prefix on traits | Not idiomatic Rust | `Foo` trait, `FooImpl` if needed |
| Single-letter variable names | Unreadable (except in closures/iterators) | Descriptive names |

---

## Ownership & Borrowing

### ✅ **Prefer Borrowing Over Ownership**

```rust
// GOOD - Borrows the string, caller retains ownership
fn validate_email(email: &str) -> bool {
    email.contains('@') && email.contains('.')
}

// BAD - Takes ownership unnecessarily
fn validate_email(email: String) -> bool {
    email.contains('@') && email.contains('.')
}
```

**Why This Matters:**
- Borrowing avoids unnecessary allocations and clones
- Caller retains ownership for reuse
- `&str` accepts both `String` and `&str` via deref coercion

### ✅ **Lifetime Annotations**

```rust
// GOOD - Explicit lifetime ties output to input
fn first_word(s: &str) -> &str {
    s.split_whitespace().next().unwrap_or("")
}

// GOOD - Multiple lifetimes when inputs have different scopes
fn longest<'a>(x: &'a str, y: &'a str) -> &'a str {
    if x.len() > y.len() { x } else { y }
}

// GOOD - Struct borrowing data
struct Config<'a> {
    name: &'a str,
    version: &'a str,
}
```

**Lifetime Elision Rules (when annotations are NOT needed):**
1. Each reference parameter gets its own lifetime
2. If exactly one input lifetime, it applies to all output lifetimes
3. If `&self` or `&mut self`, its lifetime applies to all output lifetimes

### ✅ **Copy vs Clone**

```rust
// GOOD - Small, stack-only types implement Copy
#[derive(Debug, Clone, Copy, PartialEq)]
struct Point {
    x: f64,
    y: f64,
}

// GOOD - Types with heap data implement Clone only
#[derive(Debug, Clone)]
struct Config {
    name: String,       // String is Clone but NOT Copy
    retries: u32,
}
```

| Trait | Behavior | Use When |
|-------|----------|----------|
| `Copy` | Implicit bitwise copy | Small stack-only types (integers, bools, tuples of Copy types) |
| `Clone` | Explicit `.clone()` | Heap-allocated or expensive-to-copy types |
| Neither | Move semantics | Unique resources (file handles, connections) |

### ⚠️ **Common Ownership Mistakes**

```rust
// BAD - Unnecessary clone to satisfy borrow checker
let name = config.name.clone();
process(&name);
process2(&config.name); // Could have borrowed directly

// GOOD - Borrow instead of clone
process(&config.name);
process2(&config.name);

// BAD - Moving out of a shared reference
fn take_name(config: &Config) -> String {
    config.name // ERROR: cannot move out of borrowed content
}

// GOOD - Clone when you truly need ownership from a borrow
fn take_name(config: &Config) -> String {
    config.name.clone()
}
```

---

## Error Handling Patterns

### ✅ **Custom Error Types with thiserror**

```rust
use thiserror::Error;

#[derive(Debug, Error)]
pub enum AppError {
    #[error("configuration error: {0}")]
    Config(String),

    #[error("database query failed: {source}")]
    Database {
        #[source]
        source: sqlx::Error,
    },

    #[error("HTTP request failed: {url}")]
    Http {
        url: String,
        #[source]
        source: reqwest::Error,
    },

    #[error("not found: {entity} with id {id}")]
    NotFound { entity: &'static str, id: String },

    #[error(transparent)]
    Unexpected(#[from] anyhow::Error),
}
```

**Requirements:**
- ✅ Use `thiserror` for library error types (structured, matchable)
- ✅ Use `anyhow` for application-level errors (ergonomic, context-rich)
- ✅ Implement `#[source]` for error chain inspection
- ✅ Implement `#[from]` for automatic conversion via `?`
- ✅ Human-readable display messages

### ✅ **The ? Operator and Error Propagation**

```rust
// GOOD - Clean error propagation with context
use anyhow::{Context, Result};

fn load_config(path: &str) -> Result<Config> {
    let contents = std::fs::read_to_string(path)
        .with_context(|| format!("failed to read config from {path}"))?;

    let config: Config = toml::from_str(&contents)
        .with_context(|| format!("failed to parse config from {path}"))?;

    config.validate()
        .context("config validation failed")?;

    Ok(config)
}

// BAD - Manual match on every error
fn load_config(path: &str) -> Result<Config, Box<dyn std::error::Error>> {
    let contents = match std::fs::read_to_string(path) {
        Ok(c) => c,
        Err(e) => return Err(Box::new(e)),
    };
    // ... tedious repetition
}
```

### ✅ **Result Type Aliases**

```rust
// GOOD - Crate-level Result alias
pub type Result<T> = std::result::Result<T, AppError>;

// Usage throughout the crate
pub fn get_user(id: u64) -> Result<User> {
    // AppError is the implicit error type
    Ok(User { id, name: "Alice".into() })
}
```

### ⚠️ **Error Handling Anti-Patterns**

| Pattern | Problem | Instead |
|---------|---------|---------|
| `.unwrap()` in production | Panics on None/Err | Use `?`, `.unwrap_or()`, or match |
| `Box<dyn Error>` everywhere | Loses type info | Use `thiserror` enums |
| String errors | Not matchable | Use typed errors |
| Swallowing errors silently | Hides bugs | Log or propagate |
| `panic!()` for expected failures | Crashes the process | Return `Result` |

**Unwrap Threshold:**

| Context | `.unwrap()` Allowed? |
|---------|---------------------|
| Tests | ✅ Yes |
| Examples | ✅ Yes |
| Build scripts | ⚠️ Acceptable with comment |
| Library code | ❌ Never |
| Binary (main) | ⚠️ Only after validation |

---

## Trait & Type System Design

### ✅ **Trait Design**

```rust
// GOOD - Small, focused traits
pub trait Validate {
    fn validate(&self) -> Result<(), ValidationError>;
}

pub trait Persist {
    fn save(&self, store: &dyn Store) -> Result<()>;
    fn load(id: &str, store: &dyn Store) -> Result<Self>
    where
        Self: Sized;
}

// Compose traits via supertraits
pub trait Entity: Validate + Persist + std::fmt::Debug {}
```

**Anti-Pattern (God Trait):**
```rust
// BAD - Too many methods, forces implementors to define everything
pub trait Service {
    fn start(&self) -> Result<()>;
    fn stop(&self) -> Result<()>;
    fn health(&self) -> HealthStatus;
    fn metrics(&self) -> Metrics;
    fn configure(&mut self, config: Config);
    fn validate(&self) -> Result<()>;
    // ... 15 more methods
}
```

### ✅ **Generics vs Trait Objects**

```rust
// GOOD - Static dispatch (monomorphized, zero-cost abstraction)
fn process<T: Serialize + Send>(item: T) -> Result<()> {
    let json = serde_json::to_string(&item)?;
    send_to_queue(&json)
}

// GOOD - Dynamic dispatch (runtime polymorphism, smaller binary)
fn process_any(item: &dyn Serialize) -> Result<()> {
    let json = serde_json::to_value(item)?;
    send_to_queue(&json.to_string())
}
```

| Approach | Binary Size | Performance | Use When |
|----------|------------|-------------|----------|
| Generics (`T: Trait`) | Larger (monomorphized) | Faster (inlined) | Hot paths, known types at compile time |
| Trait Objects (`dyn Trait`) | Smaller | Vtable overhead | Collections of mixed types, plugin systems |
| `impl Trait` (return) | Smaller | Inlined | Returning closures or iterators |

### ✅ **Associated Types vs Generics**

```rust
// GOOD - Associated type when there's ONE natural choice per impl
pub trait Iterator {
    type Item;
    fn next(&mut self) -> Option<Self::Item>;
}

// GOOD - Generic parameter when impl can work with MANY types
pub trait From<T> {
    fn from(value: T) -> Self;
}
```

### ✅ **Derive Macros**

```rust
// GOOD - Derive common traits
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct User {
    pub id: u64,
    pub name: String,
    pub email: String,
}
```

**Standard Derive Order:**

| Priority | Traits | Purpose |
|----------|--------|---------|
| 1 | `Debug` | Always derive for debugging |
| 2 | `Clone`, `Copy` | If semantically appropriate |
| 3 | `PartialEq`, `Eq` | If comparison is needed |
| 4 | `Hash` | If used as HashMap key |
| 5 | `Serialize`, `Deserialize` | If serialized |
| 6 | `Default` | If zero-value makes sense |

---

## Concurrency Patterns

### ✅ **Shared State with Arc<Mutex<T>>**

```rust
use std::sync::Arc;
use tokio::sync::Mutex;

#[derive(Clone)]
struct AppState {
    db: Arc<Mutex<Database>>,
    cache: Arc<dashmap::DashMap<String, String>>,
}

// GOOD - Lock scope is minimal
async fn get_user(state: &AppState, id: u64) -> Result<User> {
    let db = state.db.lock().await;
    let user = db.query_user(id).await?;
    drop(db); // Explicit drop releases lock before further processing
    Ok(user)
}

// BAD - Holding lock across await points
async fn bad_get_user(state: &AppState, id: u64) -> Result<User> {
    let db = state.db.lock().await;
    let user = db.query_user(id).await?; // Lock held across .await!
    let enriched = enrich_user(user).await; // Still holding lock!
    Ok(enriched)
}
```

**Lock Duration Thresholds:**

| Duration | Status | Action |
|----------|--------|--------|
| < 1 ms | ✅ Excellent | Maintain |
| 1-10 ms | ⚠️ Warning | Review scope |
| > 10 ms | ❌ Critical | Refactor (clone-and-release pattern) |

### ✅ **Channel Patterns**

```rust
use tokio::sync::mpsc;

// GOOD - Bounded channel with backpressure
let (tx, mut rx) = mpsc::channel::<Event>(100);

// Producer
tokio::spawn(async move {
    for event in events {
        if tx.send(event).await.is_err() {
            tracing::warn!("receiver dropped, stopping producer");
            break;
        }
    }
});

// Consumer
tokio::spawn(async move {
    while let Some(event) = rx.recv().await {
        process_event(event).await;
    }
});
```

### ✅ **Send and Sync Bounds**

```rust
// GOOD - Explicit Send + Sync bounds for spawned futures
fn spawn_worker<F>(task: F) -> tokio::task::JoinHandle<()>
where
    F: Future<Output = ()> + Send + 'static,
{
    tokio::spawn(task)
}

// GOOD - Ensure types are thread-safe
struct SharedConfig {
    data: Arc<RwLock<HashMap<String, String>>>,  // Send + Sync
}
```

| Marker | Meaning | NOT Send/Sync |
|--------|---------|---------------|
| `Send` | Can be transferred across threads | `Rc<T>`, `*const T` |
| `Sync` | Can be shared between threads via `&T` | `Cell<T>`, `RefCell<T>` |
| Both | Safe for concurrent access | `Arc<Mutex<T>>` is both |

### ✅ **Async/Await Best Practices**

```rust
// GOOD - Select for racing multiple futures
tokio::select! {
    result = process_request(&req) => {
        handle_response(result).await;
    }
    _ = tokio::time::sleep(Duration::from_secs(30)) => {
        return Err(AppError::Timeout);
    }
    _ = shutdown_signal.recv() => {
        tracing::info!("shutting down gracefully");
        return Ok(());
    }
}

// GOOD - Spawn blocking work off the async runtime
let hash = tokio::task::spawn_blocking(move || {
    argon2::hash_encoded(password.as_bytes(), &salt, &config)
}).await??;
```

---

## Unsafe Code

### ✅ **SAFETY Comments (Required)**

```rust
// GOOD - Every unsafe block has a SAFETY comment
let value = unsafe {
    // SAFETY: We verified that `ptr` is non-null and properly aligned
    // in the check above (line 42). The pointed-to data is initialized
    // by `init_buffer()` called on line 38 and has not been freed.
    *ptr
};

// BAD - Unsafe with no justification
let value = unsafe { *ptr };
```

**Requirements:**
- ✅ Every `unsafe` block must have a `// SAFETY:` comment
- ✅ Comment must explain WHY the invariants are upheld
- ✅ Reference the specific preconditions being satisfied

### ✅ **Minimizing Unsafe Scope**

```rust
// GOOD - Minimal unsafe block, safe wrapper
pub fn get_element(slice: &[u8], index: usize) -> Option<u8> {
    if index < slice.len() {
        // SAFETY: We just verified index is within bounds
        Some(unsafe { *slice.get_unchecked(index) })
    } else {
        None
    }
}

// BAD - Entire function is unsafe when only one operation needs it
pub unsafe fn get_element(slice: &[u8], index: usize) -> u8 {
    let ptr = slice.as_ptr().add(index);
    let extra = compute_offset(ptr); // This doesn't need unsafe!
    let result = *ptr;
    log_access(result);               // This doesn't need unsafe!
    result
}
```

### ✅ **FFI (Foreign Function Interface)**

```rust
// GOOD - Safe wrapper around FFI
mod ffi {
    extern "C" {
        fn c_process(data: *const u8, len: usize) -> i32;
    }
}

/// Process data using the C library.
///
/// # Errors
/// Returns `Err` if the C function returns a non-zero exit code.
pub fn process(data: &[u8]) -> Result<(), FfiError> {
    // SAFETY: `data.as_ptr()` is valid for `data.len()` bytes.
    // The C function does not retain the pointer after returning.
    let result = unsafe { ffi::c_process(data.as_ptr(), data.len()) };
    if result == 0 {
        Ok(())
    } else {
        Err(FfiError::ExitCode(result))
    }
}
```

### ⚠️ **Unsafe Code Thresholds**

| Metric | Status | Action |
|--------|--------|--------|
| 0 unsafe blocks | ✅ Ideal | Maintain |
| 1-5 with SAFETY comments | ✅ Acceptable | Audit quarterly |
| 6-15 with SAFETY comments | ⚠️ Warning | Justify each, seek safe alternatives |
| Any without SAFETY comments | ❌ Critical | Add comments immediately |
| `#[allow(unsafe_code)]` crate-wide | ❌ Critical | Remove, audit all unsafe |

---

## Testing Patterns

### ✅ **Unit Tests (Inline Modules)**

```rust
pub fn calculate_discount(price: f64, tier: &str) -> f64 {
    match tier {
        "gold" => price * 0.20,
        "silver" => price * 0.10,
        _ => 0.0,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn gold_tier_gets_twenty_percent() {
        let discount = calculate_discount(100.0, "gold");
        assert!((discount - 20.0).abs() < f64::EPSILON);
    }

    #[test]
    fn unknown_tier_gets_no_discount() {
        assert_eq!(calculate_discount(100.0, "bronze"), 0.0);
    }

    #[test]
    fn zero_price_returns_zero() {
        assert_eq!(calculate_discount(0.0, "gold"), 0.0);
    }
}
```

**Requirements:**
- ✅ `#[cfg(test)]` module in the same file as the code
- ✅ Test names describe the expected behavior
- ✅ `use super::*` imports the parent module

### ✅ **Integration Tests**

```rust
// tests/api_integration.rs
// Each file in tests/ is compiled as a separate crate

use my_service::{Config, Server};

#[tokio::test]
async fn server_responds_to_health_check() {
    let config = Config::test_default();
    let server = Server::start(config).await.unwrap();

    let resp = reqwest::get(&format!("{}/health", server.url()))
        .await
        .unwrap();

    assert_eq!(resp.status(), 200);
    server.shutdown().await;
}
```

### ✅ **Doc Tests**

```rust
/// Parses a duration string like "5s", "100ms", "2m".
///
/// # Examples
///
/// ```
/// use my_crate::parse_duration;
///
/// let d = parse_duration("5s").unwrap();
/// assert_eq!(d, std::time::Duration::from_secs(5));
///
/// let d = parse_duration("100ms").unwrap();
/// assert_eq!(d, std::time::Duration::from_millis(100));
/// ```
///
/// # Errors
///
/// Returns `Err` if the string is not a valid duration format.
pub fn parse_duration(s: &str) -> Result<Duration, ParseError> {
    // ...
}
```

**Why Doc Tests Matter:**
- Examples in documentation are compiled and tested
- Guarantees documentation stays accurate
- `cargo test` runs doc tests by default

### ✅ **Property-Based Testing with proptest**

```rust
use proptest::prelude::*;

proptest! {
    #[test]
    fn roundtrip_serialization(input in "\\PC{1,100}") {
        let serialized = serde_json::to_string(&input).unwrap();
        let deserialized: String = serde_json::from_str(&serialized).unwrap();
        prop_assert_eq!(input, deserialized);
    }

    #[test]
    fn discount_never_exceeds_price(price in 0.0f64..10000.0, tier in "gold|silver|bronze") {
        let discount = calculate_discount(price, &tier);
        prop_assert!(discount <= price);
        prop_assert!(discount >= 0.0);
    }
}
```

### ✅ **Benchmarks with Criterion**

```rust
// benches/throughput.rs
use criterion::{black_box, criterion_group, criterion_main, Criterion};
use my_service::process;

fn benchmark_process(c: &mut Criterion) {
    let data = setup_test_data();

    c.bench_function("process_1000_items", |b| {
        b.iter(|| process(black_box(&data)))
    });
}

criterion_group!(benches, benchmark_process);
criterion_main!(benches);
```

**Running:**
```bash
cargo bench                       # Run all benchmarks
cargo bench -- process            # Run matching benchmarks
```

### Test Type Summary

| Type | Location | Runs With | Purpose |
|------|----------|-----------|---------|
| Unit | `#[cfg(test)]` inline | `cargo test` | Test private functions |
| Integration | `tests/` directory | `cargo test` | Test public API |
| Doc | `///` comments | `cargo test` | Verify examples |
| Property | Inline or `tests/` | `cargo test` | Fuzz invariants |
| Benchmark | `benches/` | `cargo bench` | Performance regression |

---

## Security Practices

### ✅ **Minimize Unsafe Code**

```rust
// CORRECT — isolate unsafe behind a safe API
pub fn read_buffer(ptr: *const u8, len: usize) -> &[u8] {
    // SAFETY: caller guarantees ptr is valid for `len` bytes,
    // properly aligned, and the memory won't be mutated during
    // the lifetime of the returned slice.
    unsafe { std::slice::from_raw_parts(ptr, len) }
}

// INCORRECT — unsafe scattered through business logic
pub fn process(data: *const u8) {
    unsafe {
        // Multiple unsafe operations without justification
        let val = *data;
        let next = *data.add(1);
    }
}
```

**Unsafe Audit Criteria:**
- Every `unsafe` block MUST have a `// SAFETY:` comment explaining the invariant
- Minimize the scope of `unsafe` — wrap in safe abstractions
- Prefer safe alternatives: `Vec`, `Box`, `Rc`/`Arc` over raw pointers
- Audit all `unsafe impl Send` and `unsafe impl Sync` for correctness

### ✅ **FFI Safety**

```rust
// CORRECT — safe wrapper around FFI
extern "C" {
    fn c_process(data: *const u8, len: usize) -> i32;
}

/// Process data through the C library.
///
/// # Panics
/// Panics if `data` is empty.
pub fn process(data: &[u8]) -> Result<(), FfiError> {
    assert!(!data.is_empty(), "data must not be empty");
    // SAFETY: data.as_ptr() is valid for data.len() bytes,
    // and c_process does not retain the pointer.
    let result = unsafe { c_process(data.as_ptr(), data.len()) };
    match result {
        0 => Ok(()),
        code => Err(FfiError::ReturnCode(code)),
    }
}
```

**FFI Rules:**
- Always validate inputs before crossing the FFI boundary
- Wrap every `extern "C"` function in a safe Rust API
- Never expose raw pointers in public APIs
- Use `CStr`/`CString` for string interchange, never cast directly

### ✅ **Input Validation**

```rust
use std::net::IpAddr;

pub fn parse_config(input: &str) -> Result<Config, ConfigError> {
    let config: Config = toml::from_str(input)
        .map_err(ConfigError::Parse)?;

    // Validate bounds after deserialization
    if config.port == 0 || config.port > 65535 {
        return Err(ConfigError::InvalidPort(config.port));
    }
    if config.max_connections > 10_000 {
        return Err(ConfigError::ExceedsLimit("max_connections", 10_000));
    }

    Ok(config)
}
```

**Validation Rules:**
- Validate all external data at system boundaries (CLI args, env vars, files, network)
- Use newtypes to enforce invariants at the type level
- Prefer `TryFrom` over unchecked conversions

### ✅ **Dependency Auditing**

```bash
# Audit for known vulnerabilities
cargo audit

# Check for unmaintained or yanked crates
cargo audit --deny warnings

# In CI, fail the build on any advisory
cargo audit --deny vulnerability --deny unmaintained --deny yanked
```

**Dependency Rules:**
- Run `cargo audit` in CI on every PR
- Pin major versions in `Cargo.toml` (e.g., `serde = "1"`)
- Review new dependencies for `unsafe` usage and maintenance status
- Prefer crates from the RustSec-reviewed ecosystem

### Security ALWAYS / NEVER

| ALWAYS | NEVER |
|--------|-------|
| Add `// SAFETY:` comment on every `unsafe` block | Use `unsafe` without documenting the invariant |
| Wrap FFI calls in safe Rust abstractions | Expose raw pointers in public APIs |
| Validate external inputs at system boundaries | Trust deserialized data without bounds checks |
| Run `cargo audit` in CI | Ignore advisory warnings on dependencies |
| Use `CStr`/`CString` for C string interchange | Cast `*const u8` to `&str` without validation |
| Minimize `unsafe` block scope | Scatter `unsafe` through business logic |

---

## Documentation Standards

### ✅ **Rustdoc Conventions**

```rust
//! # mycrate
//!
//! A library for processing widgets efficiently.
//!
//! ## Quick Start
//!
//! ```rust
//! use mycrate::Widget;
//! let w = Widget::new("example");
//! assert!(w.is_valid());
//! ```

/// A widget that can be processed.
///
/// Widgets are the core data type. They must be created
/// via [`Widget::new`] to ensure invariants are upheld.
///
/// # Examples
///
/// ```
/// use mycrate::Widget;
///
/// let widget = Widget::new("test");
/// assert_eq!(widget.name(), "test");
/// ```
pub struct Widget {
    name: String,
}

impl Widget {
    /// Creates a new widget with the given name.
    ///
    /// # Panics
    ///
    /// Panics if `name` is empty.
    ///
    /// # Examples
    ///
    /// ```
    /// use mycrate::Widget;
    /// let w = Widget::new("example");
    /// ```
    pub fn new(name: &str) -> Self {
        assert!(!name.is_empty(), "name must not be empty");
        Self { name: name.to_string() }
    }
}
```

### ✅ **Doc Test Patterns**

```rust
/// Parses a duration string like "5s", "100ms".
///
/// # Examples
///
/// ```
/// # use mycrate::parse_duration;
/// assert_eq!(parse_duration("5s").unwrap().as_secs(), 5);
/// assert_eq!(parse_duration("100ms").unwrap().as_millis(), 100);
/// ```
///
/// # Errors
///
/// Returns [`ParseError`] if the format is unrecognized.
///
/// ```
/// # use mycrate::parse_duration;
/// assert!(parse_duration("invalid").is_err());
/// ```
pub fn parse_duration(s: &str) -> Result<Duration, ParseError> {
    // ...
}
```

**Doc Test Rules:**
- Doc tests compile and run with `cargo test` — treat them as real tests
- Use `# ` prefix to hide setup lines (imports, boilerplate)
- Use `no_run` for examples that need network/filesystem
- Use `should_panic` for examples demonstrating failure

### ✅ **Module Documentation**

```rust
//! # handlers
//!
//! HTTP request handlers for the API.
//!
//! Each handler follows the pattern:
//! 1. Parse and validate input
//! 2. Call domain logic
//! 3. Map result to HTTP response
//!
//! See [`crate::domain`] for business logic.
```

### ✅ **`#[doc(hidden)]` Usage**

```rust
// Hide implementation details from public docs
#[doc(hidden)]
pub mod __internal {
    // Used by macros, not part of public API
}

// Hide trait impls that are required but not user-facing
#[doc(hidden)]
pub fn __macro_helper() {}
```

**When to use `#[doc(hidden)]`:**
- Macro support functions that must be `pub` but are not API
- Trait implementations required by the compiler but meaningless to users
- Never hide things to avoid documenting them

### Documentation ALWAYS / NEVER

| ALWAYS | NEVER |
|--------|-------|
| Use `///` for public items, `//!` for module-level docs | Leave public API items undocumented |
| Include `# Examples` section on public functions | Write doc tests that don't actually assert behavior |
| Document `# Panics`, `# Errors`, `# Safety` sections | Use `#[doc(hidden)]` to avoid writing documentation |
| Run `cargo test` to verify doc examples compile | Assume doc examples stay correct without CI |
| Link to related items with [`ident`] syntax | Duplicate information already in type signatures |
| Use `# ` to hide boilerplate in doc tests | Write doc tests that require external services |

---

## Code Quality Metrics & Anti-Patterns

> See `common-standards.md` for universal coverage targets, testing principles, and anti-patterns across all languages.

### ✅ **Clippy Lint Levels**

```toml
# Cargo.toml or clippy.toml
[lints.clippy]
# Deny — treat as errors
unwrap_used = "deny"
expect_used = "deny"
panic = "deny"
todo = "deny"

# Warn — flag for review
clone_on_ref_ptr = "warn"
large_enum_variant = "warn"
needless_pass_by_value = "warn"
implicit_clone = "warn"
missing_errors_doc = "warn"
missing_panics_doc = "warn"
```

**Recommended CI Command:**
```bash
cargo clippy --all-targets --all-features -- -D warnings
```

### ✅ **Lint Category Enforcement**

| Category | CI Policy | Rationale |
|----------|-----------|-----------|
| `clippy::correctness` | ❌ Deny (fail build) | Likely bugs |
| `clippy::suspicious` | ❌ Deny (fail build) | Probably wrong |
| `clippy::pedantic` | ⚠️ Warn | Style improvements |
| `clippy::nursery` | ⚠️ Optional | Experimental lints |
| `clippy::cargo` | ⚠️ Warn | Cargo.toml hygiene |

### 📊 **Complexity Thresholds**

| Complexity Range | Status | Action |
|-----------------|--------|--------|
| CC 1-5 (Simple) | ✅ Excellent | Maintain |
| CC 6-10 (OK) | ✅ Acceptable | Monitor |
| CC 11-15 (High) | ⚠️ Warning | Refactor recommended |
| CC 16+ (Very High) | ❌ Critical | Refactor required |

**Coverage Targets:**

| Metric | Minimum | Target |
|--------|---------|--------|
| Line coverage | 60% | 80%+ |
| Branch coverage | 50% | 70%+ |
| Critical path coverage | 90% | 100% |

### ❌ **Named Anti-Patterns**

**1. Stringly-Typed Code**
```rust
// BAD - Strings for everything
fn set_status(status: &str) { /* "active", "idle", "error" */ }

// GOOD - Enums encode valid states
enum Status { Active, Idle, Error }
fn set_status(status: Status) { /* ... */ }
```

**2. Clone-Happy Code**
```rust
// BAD - Cloning to avoid borrow checker fights
fn process(data: &Data) {
    let owned = data.clone();   // Unnecessary allocation
    compute(&owned);
}

// GOOD - Work with references
fn process(data: &Data) {
    compute(data);
}
```

**3. Typestate Neglect**
```rust
// BAD - Runtime checks for compile-time invariants
struct Connection { is_authenticated: bool }
fn query(conn: &Connection) {
    assert!(conn.is_authenticated); // Runtime panic
}

// GOOD - Typestate pattern enforces at compile time
struct Unauthenticated;
struct Authenticated;
struct Connection<State> { _state: std::marker::PhantomData<State> }

impl Connection<Unauthenticated> {
    fn authenticate(self, creds: &Credentials) -> Result<Connection<Authenticated>> {
        // ...
    }
}

impl Connection<Authenticated> {
    fn query(&self, sql: &str) -> Result<Rows> {
        // Can only be called on authenticated connections
    }
}
```

**4. Arc<Mutex<T>> Everywhere**
```rust
// BAD - Mutex when only reads happen
let config = Arc::new(Mutex::new(load_config()));

// GOOD - Use RwLock for read-heavy workloads
let config = Arc::new(RwLock::new(load_config()));

// BETTER - Use Arc<T> if config is immutable after init
let config = Arc::new(load_config());
```

**5. Ignoring Must-Use Types**
```rust
// BAD - Ignoring a Result
fn fire_and_forget() {
    send_notification(); // Warning: unused Result
}

// GOOD - Explicitly acknowledge
fn fire_and_forget() {
    let _ = send_notification(); // Intentional ignore
}
```

**6. Unbounded Collections**
```rust
// BAD - No size limit on cache
let mut cache: HashMap<String, Data> = HashMap::new();
// Grows forever...

// GOOD - Bounded with eviction
let cache = lru::LruCache::new(NonZeroUsize::new(10_000).unwrap());
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

| Category | Assessment Criteria | Evidence Required |
|----------|-------------------|-------------------|
| Project Structure | Standard layout, module sizes, re-exports | File count per module, module line counts |
| Cargo Config | MSRV set, features used, profiles configured | Cargo.toml audit, dep count |
| Code Formatting | rustfmt clean, naming conventions followed | `cargo fmt --check` output, naming violations |
| Ownership & Borrowing | Minimal clones, correct lifetimes, no unnecessary ownership | Clone count, borrow checker workarounds |
| Error Handling | thiserror/anyhow usage, no unwrap in prod, context added | Unwrap count, error type audit |
| Traits & Types | Small traits, appropriate dispatch, derive usage | Methods per trait, dyn vs generic ratio |
| Concurrency | Minimal lock scope, bounded channels, Send/Sync correct | Lock duration, channel audit |
| Unsafe Code | SAFETY comments, minimal scope, safe wrappers | Unsafe block count, comment coverage |
| Testing | Unit + integration + doc tests, property tests | Coverage %, test type distribution |
| Code Quality | Clippy clean, low complexity, no named anti-patterns | Clippy findings, CC distribution |

**Grading Scale:**

| Grade | Finding Threshold | Description |
|-------|------------------|-------------|
| A+ | 0-2 minor findings | Exemplary - industry best practices |
| A | <5 HIGH findings | Excellent - strong practices |
| A- | 5-15 HIGH findings | Very Good - solid practices |
| B+ | 15-25 HIGH findings | Good - acceptable practices |
| B | 25-40 HIGH findings | Satisfactory - needs improvement |
| C+ | 40-60 HIGH findings | Needs Improvement - multiple issues |
| C | 60+ HIGH findings | Significant Issues - major refactoring |
| D | 1+ CRITICAL findings | Major Problems - not production-ready |
| F | Multiple CRITICAL | Critical Issues - complete rewrite |

**Example Assessment:**

| Category | Grade | Evidence |
|----------|-------|----------|
| Error Handling | A | 0 unwraps in lib code, 45 proper `?` propagations, thiserror enums |
| Ownership | A- | 3 unnecessary clones flagged, all lifetimes correct |
| Concurrency | A+ | All locks < 1ms scope, bounded channels, no deadlock paths |
| Unsafe Code | A+ | 0 unsafe blocks in application code |
| Testing | B+ | 72% line coverage, doc tests on public API, no property tests |
| **OVERALL** | **A- (Excellent)** | **8 HIGH, 22 MEDIUM findings** |

---

## Vibe Integration

### Prescan Patterns

| Pattern | Severity | Detection |
|---------|----------|-----------|
| PR-01: `.unwrap()` in library code | HIGH | grep for `.unwrap()` outside `#[cfg(test)]` |
| PR-02: Missing SAFETY comments | CRITICAL | `unsafe` blocks without `// SAFETY:` |
| PR-03: Clippy warnings | HIGH | `cargo clippy` JSON output parsing |
| PR-04: Unformatted code | MEDIUM | `cargo fmt --check` exit code |
| PR-05: Unused dependencies | LOW | `cargo machete` output |

### Semantic Analysis

Deep validation includes:
- Ownership pattern analysis (clone frequency, lifetime correctness)
- Trait design review (ISP compliance, dispatch appropriateness)
- Concurrency safety audit (lock scope, Send/Sync bounds)
- Unsafe code audit (SAFETY comments, scope minimization)

### JIT Loading

**Tier 1 (Fast):** Load `~/.agents/skills/standards/references/rust.md` (5KB)
**Tier 2 (Deep):** Load this document (~20KB) for comprehensive audit
**Override:** Use `.agents/validation/RUST_*.md` if project-specific standards exist

---

## Additional Resources

- [The Rust Programming Language](https://doc.rust-lang.org/book/)
- [Rust API Guidelines](https://rust-lang.github.io/api-guidelines/)
- [Rust Design Patterns](https://rust-unofficial.github.io/patterns/)
- [Clippy Lints](https://rust-lang.github.io/rust-clippy/master/)
- [Rust Performance Book](https://nnethercote.github.io/perf-book/)
- [The Rustonomicon](https://doc.rust-lang.org/nomicon/) (unsafe Rust)

---

**Related:** `rust-patterns.md` for quick reference examples
