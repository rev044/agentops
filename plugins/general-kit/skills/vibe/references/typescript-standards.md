# TypeScript Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-01-21
**Purpose:** Canonical TypeScript standards for vibe skill validation

---

## Table of Contents

1. [Strict Configuration](#strict-configuration)
2. [ESLint Configuration](#eslint-configuration)
3. [Type System Patterns](#type-system-patterns)
4. [Generic Constraints](#generic-constraints)
5. [Utility Types](#utility-types)
6. [Error Handling](#error-handling)
7. [Module Template](#module-template)
8. [Anti-Patterns Avoided](#anti-patterns-avoided)
9. [Compliance Assessment](#compliance-assessment)

---

## Strict Configuration

### Full tsconfig.json

Every TypeScript project MUST use strict mode:

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "lib": ["ES2022"],
    "outDir": "./dist",
    "rootDir": "./src",

    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "exactOptionalPropertyTypes": true,

    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

---

## ESLint Configuration

### eslint.config.js (Flat Config)

```javascript
import eslint from '@eslint/js';
import tseslint from 'typescript-eslint';

export default tseslint.config(
  eslint.configs.recommended,
  ...tseslint.configs.strictTypeChecked,
  ...tseslint.configs.stylisticTypeChecked,
  {
    languageOptions: {
      parserOptions: {
        project: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      '@typescript-eslint/explicit-function-return-type': 'error',
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/prefer-nullish-coalescing': 'error',
      '@typescript-eslint/prefer-optional-chain': 'error',
      '@typescript-eslint/no-floating-promises': 'error',
      '@typescript-eslint/await-thenable': 'error',
    },
  },
  {
    ignores: ['dist/', 'node_modules/', '*.js'],
  }
);
```

---

## Type System Patterns

### Discriminated Unions

Use discriminated unions for state modeling:

```typescript
type Result<T, E> =
  | { status: 'success'; data: T }
  | { status: 'error'; error: E };

function handleResult<T, E>(result: Result<T, E>): void {
  switch (result.status) {
    case 'success':
      console.log(result.data);
      break;
    case 'error':
      console.error(result.error);
      break;
  }
}
```

### Const Assertions

Use `as const` for literal types:

```typescript
const CONFIG = {
  apiVersion: 'v1',
  retries: 3,
  endpoints: ['primary', 'fallback'],
} as const;
```

### Branded Types

Use branded types for type-safe IDs:

```typescript
type UserId = string & { readonly __brand: 'UserId' };
type OrderId = string & { readonly __brand: 'OrderId' };

function createUserId(id: string): UserId {
  return id as UserId;
}
```

---

## Generic Constraints

### Constrained Generics

Always constrain generics when possible:

```typescript
// Good - constrained generic
function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}

// Good - multiple constraints
function merge<T extends object, U extends object>(a: T, b: U): T & U {
  return { ...a, ...b };
}
```

---

## Utility Types

### Built-in Utilities

Use built-in utility types over manual definitions:

```typescript
// Partial - all properties optional
type PartialUser = Partial<User>;

// Required - all properties required
type RequiredConfig = Required<Config>;

// Pick - select properties
type UserPreview = Pick<User, 'id' | 'name'>;

// Omit - exclude properties
type UserWithoutPassword = Omit<User, 'password'>;

// Record - typed object
type UserMap = Record<string, User>;
```

---

## Error Handling

### Result Pattern

Prefer explicit error handling over exceptions:

```typescript
type Result<T, E = Error> =
  | { ok: true; value: T }
  | { ok: false; error: E };

function parseJson<T>(json: string): Result<T, SyntaxError> {
  try {
    return { ok: true, value: JSON.parse(json) as T };
  } catch (e) {
    return { ok: false, error: e as SyntaxError };
  }
}
```

### Type Guards

Use type guards for runtime type narrowing:

```typescript
function isUser(value: unknown): value is User {
  return (
    typeof value === 'object' &&
    value !== null &&
    'id' in value &&
    'name' in value
  );
}
```

---

## Module Template

Standard template for TypeScript modules:

```typescript
/**
 * Module description.
 * @module module-name
 */

// Types first
export interface Config {
  readonly apiUrl: string;
  readonly timeout: number;
}

export type Handler<T> = (data: T) => Promise<void>;

// Type guards
export function isConfig(value: unknown): value is Config {
  return (
    typeof value === 'object' &&
    value !== null &&
    'apiUrl' in value &&
    'timeout' in value
  );
}

// Public API
export function createClient(config: Config): Client {
  return new Client(config);
}
```

---

## Anti-Patterns Avoided

### No Any Escape

```typescript
// Bad - defeats type safety
const data = response as any;
const typed = data as User;

// Good - use unknown + type guard
const data: unknown = response;
if (isUser(data)) {
  const typed: User = data;
}
```

### No Non-null Assertion Spam

```typescript
// Bad - runtime errors if assumption wrong
const name = user!.profile!.displayName!;

// Good - proper null handling
const name = user?.profile?.displayName ?? 'Anonymous';
```

### No Enum for Strings

```typescript
// Bad - verbose, poor tree-shaking
enum Color {
  Red = 'RED',
  Blue = 'BLUE',
}

// Good - union type
type Color = 'RED' | 'BLUE';
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

### Grading Scale

| Grade | Criteria |
|-------|----------|
| A+ | 0 tsc errors, 0 any types, strict mode, 0 ESLint errors, 100% return types |
| A | 0 tsc errors, <3 any types (justified), <5 ESLint errors, 95%+ return types |
| A- | <5 tsc errors, <10 any types, <15 ESLint errors, 85%+ return types |
| B+ | <15 tsc errors, <20 any types, <30 ESLint errors, 75%+ return types |
| B | <30 tsc errors, <40 any types, <50 ESLint errors, 60%+ return types |
| C | Significant type safety issues |
| D | Not production-ready |
| F | Critical issues |

---

## Additional Resources

- [TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/)
- [typescript-eslint](https://typescript-eslint.io/)
- [Total TypeScript](https://www.totaltypescript.com/)
