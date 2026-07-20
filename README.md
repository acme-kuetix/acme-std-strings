# acme-std-strings

String manipulation primitives for WSL workflows. WSL has no native string operators (no trim, uppercase, split, contains, prefix checks), so validation methods that use `strings.TrimSpace`/`ToUpper`/etc. must stay in Go unless they delegate to these primitives.

## Primitives

### Transformation (always succeed)
- `Trim(value)` → `{value}`
- `ToUpper(value)` → `{value}`
- `ToLower(value)` → `{value}`
- `Split(value, separator)` → `{items: [], count}`

### Validation (fail with 400 on invalid input)
- `AssertRequired(value, code, message)` → `{value, valid}` — trim + non-empty check
- `AssertRequiredUpper(value, code, message)` → `{value, valid}` — trim + uppercase + non-empty
- `AssertRequiredLower(value, code, message)` → `{value, valid}` — trim + lowercase + non-empty
- `AssertMinLength(value, min, code, message)` → `{value, valid}`
- `AssertExactLength(value, n, code, message)` → `{value, valid}`
- `AssertHasPrefix(value, prefixes, code, message)` → `{value, valid}` — prefixes is comma-separated list
- `AssertOneOf(value, list, code, message)` → `{value, valid}` — list is comma-separated

### Query (return booleans, never fail)
- `Contains(value, substring)` → `{contains: bool}`
- `EqualsIgnoreCase(value, other)` → `{equals: bool}`

## WSL usage

```
import strings/ops

state ValidateCode {
  action strings/ops/ops.AssertRequiredUpper(
    value: $json.code,
    code: "invalid_code",
    message: "code is required"
  ) as CodeValid
  on success -> CheckUnique
  on error -> Failed
}
```
