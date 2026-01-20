# TypeScript Kit

> TypeScript development standards and tooling for AgentOps.

## Install

```bash
/plugin install typescript-kit@agentops
```

Requires: `solo-kit`

## What's Included

### Standards

Comprehensive TypeScript coding standards in `skills/standards/references/typescript.md`:
- Strict mode configuration
- Advanced type patterns
- Error handling
- Testing with Jest/Vitest
- React patterns (if applicable)
- Common anti-patterns to avoid

### Hooks

| Hook | Trigger | What It Does |
|------|---------|--------------|
| `prettier` | Edit *.{ts,tsx,js,jsx} | Auto-format with prettier |
| `tsc-check` | Edit *.{ts,tsx} | Type check with tsc |
| `console-log-warn` | Edit *.{ts,tsx,js,jsx} | Warn about console.log |

### Patterns

**Discriminated Unions**
```typescript
type Result<T> =
  | { success: true; data: T }
  | { success: false; error: string };

function process(input: Input): Result<Output> {
  try {
    return { success: true, data: transform(input) };
  } catch (e) {
    return { success: false, error: String(e) };
  }
}
```

**Type Guards**
```typescript
function isUser(obj: unknown): obj is User {
  return (
    typeof obj === 'object' &&
    obj !== null &&
    'id' in obj &&
    'email' in obj
  );
}
```

**Generic Constraints**
```typescript
function first<T extends { id: string }>(items: T[]): T | undefined {
  return items[0];
}
```

## Requirements

- Node.js 18+
- TypeScript 5.0+
- Optional: prettier (for hooks)

## License

MIT
