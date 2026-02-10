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
6. [Conditional Types](#conditional-types)
7. [Error Handling](#error-handling)
8. [Module Template](#module-template)
9. [Code Quality Metrics](#code-quality-metrics)
10. [Testing Patterns](#testing-patterns)
11. [Anti-Patterns Avoided](#anti-patterns-avoided)
12. [Compliance Assessment](#compliance-assessment)

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

### Why Strict Matters

| Option | Effect |
|--------|--------|
| `strict: true` | Enables all strict type-checking options |
| `noUncheckedIndexedAccess` | Adds `undefined` to index signatures |
| `exactOptionalPropertyTypes` | Distinguishes `undefined` from missing |
| `noImplicitReturns` | All code paths must return |
| `noFallthroughCasesInSwitch` | Prevents accidental case fallthrough |

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

### Usage

```bash
# Lint check
npx eslint . --ext .ts,.tsx

# Fix auto-fixable issues
npx eslint . --ext .ts,.tsx --fix

# Type check only (no emit)
npx tsc --noEmit
```

---

## Type System Patterns

### Prefer Type Inference

Let TypeScript infer types when obvious:

```typescript
// Good - inference is clear
const users = ['alice', 'bob'];
const count = users.length;

// Good - explicit when non-obvious or API boundary
function getUser(id: string): User | undefined {
  return userMap.get(id);
}

// Bad - redundant annotation
const name: string = 'alice';
```

### Discriminated Unions

Use discriminated unions for state modeling:

```typescript
// Good - exhaustive pattern matching
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
    // TypeScript enforces exhaustiveness
  }
}
```

### Const Assertions

Use `as const` for literal types:

```typescript
// Good - preserves literal types
const CONFIG = {
  apiVersion: 'v1',
  retries: 3,
  endpoints: ['primary', 'fallback'],
} as const;

// Type: { readonly apiVersion: "v1"; readonly retries: 3; ... }
```

### Branded Types

Use branded types for type-safe IDs:

```typescript
type UserId = string & { readonly __brand: 'UserId' };
type OrderId = string & { readonly __brand: 'OrderId' };

function createUserId(id: string): UserId {
  return id as UserId;
}

function getUser(id: UserId): User { ... }
function getOrder(id: OrderId): Order { ... }

// Type error: can't pass UserId where OrderId expected
const user = getUser(createUserId('123'));
const order = getOrder(createUserId('123')); // Error!
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

// Bad - unconstrained (allows any)
function unsafe<T>(value: T): T {
  return value;
}
```

### Generic Defaults

Provide defaults for optional type parameters:

```typescript
interface ApiResponse<T = unknown, E = Error> {
  data?: T;
  error?: E;
  status: number;
}

// Uses defaults
const response: ApiResponse = { status: 200 };

// Override defaults
const typed: ApiResponse<User, ApiError> = { status: 200 };
```

### Generic Inference

Let TypeScript infer generic types when possible:

```typescript
// Good - infers T from argument
function identity<T>(value: T): T {
  return value;
}

const str = identity('hello'); // T inferred as string
const num = identity(42);      // T inferred as number

// Bad - unnecessary explicit type
const str2 = identity<string>('hello'); // Redundant
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

// Extract/Exclude - union manipulation
type StringOrNumber = Extract<string | number | boolean, string | number>;
```

### Custom Type Helpers

Create reusable type utilities:

```typescript
// Deep partial
type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};

// Non-nullable object values
type NonNullableValues<T> = {
  [K in keyof T]: NonNullable<T[K]>;
};

// Extract function return types from object
type ReturnTypes<T extends Record<string, (...args: never[]) => unknown>> = {
  [K in keyof T]: ReturnType<T[K]>;
};

// Make specific keys required
type WithRequired<T, K extends keyof T> = T & Required<Pick<T, K>>;
```

---

## Conditional Types

### Type-Level Logic

Use conditional types for dynamic typing:

```typescript
// Infer array element type
type ElementOf<T> = T extends readonly (infer E)[] ? E : never;

// Flatten promise type
type Awaited<T> = T extends Promise<infer U> ? Awaited<U> : T;

// Function parameter extraction
type FirstParam<T> = T extends (first: infer P, ...args: never[]) => unknown
  ? P
  : never;

// Conditional return type
type ApiResult<T> = T extends 'user'
  ? User
  : T extends 'order'
  ? Order
  : never;
```

### Template Literal Types

Use template literals for string manipulation:

```typescript
// Event handler naming
type EventName = 'click' | 'change' | 'submit';
type HandlerName = `on${Capitalize<EventName>}`;
// Result: "onClick" | "onChange" | "onSubmit"

// Path building
type ApiPath<T extends string> = `/api/v1/${T}`;
type UserPath = ApiPath<'users'>; // "/api/v1/users"

// Property getters/setters
type Getters<T> = {
  [K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};
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

// Usage
const result = parseJson<User>(input);
if (result.ok) {
  console.log(result.value.name);
} else {
  console.error(result.error.message);
}
```

### Type Guards

Use type guards for runtime type narrowing:

```typescript
// User-defined type guard
function isUser(value: unknown): value is User {
  return (
    typeof value === 'object' &&
    value !== null &&
    'id' in value &&
    'name' in value
  );
}

// Assertion function
function assertUser(value: unknown): asserts value is User {
  if (!isUser(value)) {
    throw new Error('Invalid user');
  }
}

// Usage
function processData(data: unknown): void {
  if (isUser(data)) {
    // data is User here
    console.log(data.name);
  }

  // Or with assertion
  assertUser(data);
  // data is User from here on
  console.log(data.id);
}
```

### Error Classes

Create typed error classes:

```typescript
class AppError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly statusCode: number = 500,
  ) {
    super(message);
    this.name = 'AppError';
  }
}

class ValidationError extends AppError {
  constructor(
    message: string,
    public readonly field: string,
  ) {
    super(message, 'VALIDATION_ERROR', 400);
    this.name = 'ValidationError';
  }
}

// Type guard for error handling
function isAppError(error: unknown): error is AppError {
  return error instanceof AppError;
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

// Constants
const DEFAULT_TIMEOUT = 5000;

// Private helpers (not exported)
function validateConfig(config: Config): void {
  if (!config.apiUrl) {
    throw new Error('apiUrl is required');
  }
}

// Public API
export function createClient(config: Config): Client {
  validateConfig(config);
  return new Client(config);
}

export class Client {
  readonly #config: Config;

  constructor(config: Config) {
    this.#config = config;
  }

  async fetch<T>(path: string): Promise<T> {
    const response = await fetch(`${this.#config.apiUrl}${path}`);
    return response.json() as Promise<T>;
  }
}
```

---

## Code Quality Metrics

> See `common-standards.md` for universal coverage targets and testing principles.

### Type Coverage Metrics

| Metric | Target | Validation |
|--------|--------|------------|
| tsc errors | 0 | `tsc --noEmit` |
| any types | 0 | `grep -r ": any"` |
| Explicit returns | 100% on exports | `grep "^export function"` |
| Type-only imports | 100% | Check `import type` usage |

### Validation Commands

```bash
# Type check (no emit)
tsc --noEmit
# Output: "Found X errors" → Count these

# ESLint violations
npx eslint . --ext .ts,.tsx
# Output: "X problems (Y errors, Z warnings)" → Report all

# Count any types
grep -r ": any" src/ | wc -l
# Report: "5 any types found"

# Count explicit return types on exports
grep -r "^export function" src/ | grep -c ": .* {"
# Compare to total export function count

# Type-only imports check
grep -r "^import {" src/ | grep -vc "import type"
# Report: "12 value imports (should be type-only)"
```

---

## Testing Patterns

### Jest/Vitest Configuration

Standard test runner configuration for TypeScript projects:

```typescript
// vitest.config.ts (preferred) or jest.config.ts
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',          // For React; use 'node' for backend
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      exclude: ['**/*.d.ts', '**/*.test.ts', '**/test/**'],
    },
  },
});
```

| Setting | Recommendation |
|---------|---------------|
| Runner | Vitest (preferred) or Jest 29+ with ts-jest |
| Environment | `jsdom` for UI, `node` for backend/CLI |
| Globals | `true` — avoids `import { describe, it }` boilerplate |
| Coverage provider | `v8` (fast) or `istanbul` (precise) |

### React Testing Library Patterns

Test components by user behavior, not implementation:

```typescript
// Good - tests user-visible behavior
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

test('submits form with valid data', async () => {
  const onSubmit = vi.fn();
  const user = userEvent.setup();

  render(<LoginForm onSubmit={onSubmit} />);

  await user.type(screen.getByLabelText('Email'), 'alice@example.com');
  await user.type(screen.getByLabelText('Password'), 'secret123');
  await user.click(screen.getByRole('button', { name: /sign in/i }));

  expect(onSubmit).toHaveBeenCalledWith({
    email: 'alice@example.com',
    password: 'secret123',
  });
});

// Bad - tests implementation details
test('sets state on input change', () => {
  const { container } = render(<LoginForm />);
  const input = container.querySelector('input[name="email"]')!;
  fireEvent.change(input, { target: { value: 'alice@example.com' } });
  // Brittle: relies on DOM structure and internal state
});
```

**Query Priority (prefer top to bottom):**

| Priority | Query | When |
|----------|-------|------|
| 1 | `getByRole` | Interactive elements (buttons, inputs, headings) |
| 2 | `getByLabelText` | Form fields |
| 3 | `getByText` | Non-interactive text content |
| 4 | `getByTestId` | Last resort — no accessible selector available |

### MSW (Mock Service Worker) for API Mocking

Mock API calls at the network level, not the implementation level:

```typescript
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';

const handlers = [
  http.get('/api/users/:id', ({ params }) => {
    return HttpResponse.json({
      id: params.id,
      name: 'Alice',
      email: 'alice@example.com',
    });
  }),

  http.post('/api/users', async ({ request }) => {
    const body = await request.json();
    return HttpResponse.json(body, { status: 201 });
  }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

// Override for specific test
test('handles server error', async () => {
  server.use(
    http.get('/api/users/:id', () => {
      return HttpResponse.json({ message: 'Internal error' }, { status: 500 });
    }),
  );
  // ... test error handling
});
```

### Async Testing Patterns

Use `waitFor` and async queries for asynchronous UI updates:

```typescript
// Good - waits for async state updates
import { render, screen, waitFor } from '@testing-library/react';

test('loads and displays user data', async () => {
  render(<UserProfile userId="123" />);

  // findBy* waits for element to appear (combines getBy + waitFor)
  const name = await screen.findByText('Alice');
  expect(name).toBeInTheDocument();

  // waitFor for assertions on async state
  await waitFor(() => {
    expect(screen.getByRole('status')).toHaveTextContent('Active');
  });
});

// Bad - manual timers and arbitrary delays
test('loads data', async () => {
  render(<UserProfile userId="123" />);
  await new Promise((r) => setTimeout(r, 1000)); // Flaky!
  expect(screen.getByText('Alice')).toBeInTheDocument();
});
```

### Snapshot Testing

| Use Snapshots For | Avoid Snapshots For |
|-------------------|---------------------|
| Serialized data structures (API responses, configs) | Full component trees (too brittle) |
| Error message formatting | Styled components (CSS changes break snapshots) |
| CLI output strings | Large objects (unreadable diffs) |

```typescript
// Good - small, focused snapshot
test('formats error response', () => {
  const error = formatApiError(404, 'User not found');
  expect(error).toMatchInlineSnapshot(`
    {
      "code": 404,
      "message": "User not found",
      "type": "NOT_FOUND",
    }
  `);
});

// Bad - entire component tree snapshot
test('renders dashboard', () => {
  const { container } = render(<Dashboard />);
  expect(container).toMatchSnapshot(); // 500+ line snapshot nobody reviews
});
```

### Coverage Targets

| Level | Minimum | Target | Notes |
|-------|---------|--------|-------|
| Overall | 60% | 80% | Enforced in CI |
| Critical paths | 80% | 90% | Auth, payments, data mutations |
| Utility functions | 80% | 95% | Pure functions are easy to test |
| Type guards | 100% | 100% | Runtime type safety boundary |

```bash
# Run tests with coverage
npx vitest run --coverage

# Check coverage thresholds (in vitest.config.ts)
# coverage.thresholds: { lines: 60, branches: 60, functions: 60 }
```

### Test Organization

```typescript
describe('UserService', () => {
  // Group by method
  describe('createUser', () => {
    it('creates user with valid data', async () => { /* ... */ });
    it('throws ValidationError for duplicate email', async () => { /* ... */ });
    it('hashes password before storing', async () => { /* ... */ });
  });

  describe('deleteUser', () => {
    it('soft-deletes user by setting deletedAt', async () => { /* ... */ });
    it('throws NotFoundError for unknown id', async () => { /* ... */ });
  });
});
```

**Naming conventions:**

| Convention | Example |
|-----------|---------|
| File naming | `user-service.test.ts` (co-located) or `__tests__/user-service.test.ts` |
| Describe blocks | Class/module name: `describe('UserService', ...)` |
| Test names | Behavior: `it('throws ValidationError for duplicate email')` |
| Setup files | `src/test/setup.ts` for global setup (MSW, custom matchers) |

### ALWAYS / NEVER Rules

| Rule | Type | Rationale |
|------|------|-----------|
| ALWAYS use `userEvent` over `fireEvent` | ALWAYS | `userEvent` simulates real browser behavior (focus, hover, keystrokes) |
| ALWAYS use `findBy*` for async elements | ALWAYS | Avoids race conditions; auto-retries until timeout |
| ALWAYS set `onUnhandledRequest: 'error'` in MSW | ALWAYS | Catches unmocked API calls that indicate missing test setup |
| ALWAYS co-locate test files with source | ALWAYS | Easier navigation; test dies when source is deleted |
| NEVER use `container.querySelector` in RTL tests | NEVER | Bypasses accessibility queries; tests implementation not behavior |
| NEVER use `setTimeout` / manual delays in tests | NEVER | Flaky; use `waitFor` or `findBy*` instead |
| NEVER snapshot full component trees | NEVER | Unreadable diffs; nobody reviews 500-line snapshots |
| NEVER mock what you don't own without MSW | NEVER | Direct `jest.mock('axios')` couples tests to HTTP library choice |

---

## Anti-Patterns Avoided

> See `common-standards.md` for universal anti-patterns across all languages.

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

### No Index Signature Abuse

```typescript
// Bad - no type safety
interface Config {
  [key: string]: any;
}

// Good - explicit properties
interface Config {
  apiUrl: string;
  timeout: number;
  features: string[];
}

// Or generic when truly dynamic
type Config<T extends string> = Record<T, string>;
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

// Or const object for runtime values
const Color = {
  Red: 'RED',
  Blue: 'BLUE',
} as const;
type Color = typeof Color[keyof typeof Color];
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

### Assessment Categories

| Category | Evidence Required |
|----------|------------------|
| **Type Safety** | tsc error count, any usage count, strict mode enabled |
| **Code Quality** | ESLint violations count, unused variables |
| **Type Coverage** | Explicit return types on exports (count), any/unknown ratio |
| **Best Practices** | Discriminated union usage, type guard count |
| **Testing** | Test file count, coverage % |

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

### Example Assessment

```markdown
## TypeScript Standards Compliance

**Target:** src/
**Date:** 2026-01-21

| Category | Grade | Evidence |
|----------|-------|----------|
| Type Safety | A+ | 0 tsc errors, 0 any types, strict mode |
| Code Quality | A- | 8 ESLint violations (6 auto-fixable) |
| Type Coverage | A | 48/52 exports typed (92%) |
| Best Practices | A | 12 discriminated unions, 8 type guards |
| **OVERALL** | **A** | **2 HIGH, 6 MEDIUM findings** |
```

---

## Vibe Integration

### Prescan Patterns

| Pattern | Severity | Detection |
|---------|----------|-----------|
| P10: any type usage | HIGH | `: any` without justification |
| P11: Non-null assertion spam | MEDIUM | Multiple `!` in chain |
| P12: Missing import type | LOW | `import {` for type-only |

### JIT Loading

**Tier 1 (Fast):** Load `~/.agents/skills/standards/references/typescript.md` (5KB)
**Tier 2 (Deep):** Load this document (18KB) for comprehensive audit

---

## Additional Resources

- [TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/)
- [typescript-eslint](https://typescript-eslint.io/)
- [Total TypeScript](https://www.totaltypescript.com/)
- [Type Challenges](https://github.com/type-challenges/type-challenges)

---

**Related:** Quick reference in Tier 1 `typescript.md`
