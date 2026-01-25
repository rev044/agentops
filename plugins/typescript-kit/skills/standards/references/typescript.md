# TypeScript Style Guide - Tier 1 Quick Reference

<!-- Tier 1: Generic standards (~5KB), always loaded -->
<!-- Tier 2: Deep standards in vibe/references/typescript-standards.md (~20KB), loaded on --deep -->
<!-- Last synced: 2026-01-21 -->

> **Purpose:** Quick reference for TypeScript coding standards. For comprehensive patterns, load Tier 2.

---

## Quick Reference

| Standard | Value | Validation |
|----------|-------|------------|
| **TypeScript** | 5.0+ | `tsc --version` |
| **Strict Mode** | Required | `"strict": true` in tsconfig.json |
| **Linter** | ESLint + typescript-eslint | `eslint . --ext .ts,.tsx` |
| **Formatter** | Prettier | `.prettierrc` at repo root |
| **Gate** | `tsc --noEmit` must pass | CI check |

---

## tsconfig.json (Minimum)

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitReturns": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  }
}
```

---

## Common Errors

| Symptom | Cause | Fix |
|---------|-------|-----|
| `TS2322: Type X not assignable to Y` | Type mismatch | Check types, add assertion |
| `TS2531: Object is possibly null` | Null safety missing | Add `if (x)` guard |
| `TS2339: Property does not exist` | Missing type def | Add to interface |
| `TS7006: Parameter has implicit any` | Missing annotation | Add explicit type |
| `TS2554: Expected N arguments, got M` | Wrong arg count | Check signature |
| `ESLint: no-floating-promises` | Unhandled promise | Add `await` or `void` |

---

## Anti-Patterns

| Name | Pattern | Instead |
|------|---------|---------|
| Any Escape | `as any` | Fix types, use type guards |
| Non-null Spam | `x!.y!.z!` | Proper null checks |
| Type-Only Missing | `import { Type }` | `import type { Type }` |
| Index Abuse | `[key: string]: any` | Explicit properties |
| Enum for Strings | `enum Color {}` | Union: `type Color = "RED" \| "BLUE"` |
| Callback Hell | Nested `.then()` | async/await |

---

## Type System Quick Patterns

### Discriminated Unions

```typescript
type Result<T, E> =
  | { status: 'success'; data: T }
  | { status: 'error'; error: E };
```

### Type Guards

```typescript
function isUser(value: unknown): value is User {
  return typeof value === 'object' && value !== null && 'id' in value;
}
```

### Const Assertions

```typescript
const CONFIG = { apiVersion: 'v1', retries: 3 } as const;
```

---

## Utility Types

| Type | Purpose | Example |
|------|---------|---------|
| `Partial<T>` | All optional | `Partial<User>` |
| `Required<T>` | All required | `Required<Config>` |
| `Pick<T, K>` | Select props | `Pick<User, 'id' \| 'name'>` |
| `Omit<T, K>` | Exclude props | `Omit<User, 'password'>` |
| `Record<K, V>` | Typed object | `Record<string, User>` |

---

## Summary Checklist

| Category | Requirement |
|----------|-------------|
| **Config** | `strict: true` in tsconfig.json |
| **Index Access** | `noUncheckedIndexedAccess` enabled |
| **Linting** | ESLint + typescript-eslint |
| **Type Check** | `tsc --noEmit` passes |
| **Any Types** | No `any` (enforced via ESLint) |
| **Return Types** | Explicit on all exports |
| **State** | Discriminated unions |
| **Runtime Checks** | Type guards |
| **Generics** | Always constrained |
| **Imports** | Use `import type` for types |

---

## Talos Prescan Checks

| Check | Pattern | Rationale |
|-------|---------|-----------|
| PRE-002 | TODO/FIXME markers | Track technical debt |
| PRE-010 | `any` type usage | Type safety violation |
| PRE-011 | Non-null assertion spam | Runtime error risk |
| PRE-012 | Missing `import type` | Bundler bloat |

---

## JIT Loading

**Tier 2 (Deep Standards):** For comprehensive patterns including:
- Full tsconfig.json and ESLint configuration
- Generic constraints and defaults
- Conditional and template literal types
- Result pattern for error handling
- Module template
- Validation & evidence requirements

Load: `vibe/references/typescript-standards.md`
