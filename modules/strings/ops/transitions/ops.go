// Package transitions implements the strings/ops service — string
// manipulation primitives for WSL workflows. WSL has no native string
// operators (no trim, uppercase, split, contains, prefix checks), so
// validation methods that use strings.TrimSpace/ToUpper/etc. must stay
// in Go unless they delegate to these primitives.
//
// PROMOTION-CANDIDATE: stable since Wave 6, no acme-* deps, used in 2+ packages.
// Provides Trim/ToUpper/ToLower/Split/AssertRequired/AssertOneOf/AssertHasPrefix/
// Contains/EqualsIgnoreCase. No std-* equivalent. Consider promoting to
// std-strings after kuetix review.
//
// Design: each primitive is self-contained and returns a FlowStepResult.
// Validation primitives (AssertRequired, AssertMinLength, AssertHasPrefix,
// AssertOneOf) return success with the normalized value or fail with 400.
// Transformation primitives (Trim, ToUpper, ToLower, Split) always succeed
// and return the transformed value. Query primitives (Contains,
// EqualsIgnoreCase) return a boolean.
package transitions

import (
	"fmt"
	"strings"

	"github.com/kuetix/engine/engine/domain"
	"github.com/kuetix/engine/engine/domain/interfaces"
	"github.com/kuetix/engine/engine/workflow"
)

var _ interfaces.ServiceTransitions = (*opsTransitions)(nil)

type opsTransitions struct {
	workflow.BaseServiceTransition
}

func NewOpsTransitions() interfaces.ServiceTransitions {
	return &opsTransitions{}
}

// fail is the standard error constructor for the strings package.
func fail(code, message string) (r domain.FlowStepResult) {
	r.Success = false
	r.StatusCode = 400
	r.Error = fmt.Errorf("%s", message)
	r.Response = map[string]interface{}{"code": code, "message": message}
	return
}

// ─── Transformation primitives (always succeed) ─────────────────────

// Trim returns the input with leading/trailing whitespace removed.
// WSL: strings/ops/ops.Trim(value: $json.name)
func (t *opsTransitions) Trim(value string) (r domain.FlowStepResult) {
	r.Success = true
	r.StatusCode = 200
	r.Response = map[string]interface{}{"value": strings.TrimSpace(value)}
	return
}

// ToUpper returns the input uppercased. Does NOT trim — chain Trim first
// if needed.
// WSL: strings/ops/ops.ToUpper(value: $TrimResult.response.value)
func (t *opsTransitions) ToUpper(value string) (r domain.FlowStepResult) {
	r.Success = true
	r.StatusCode = 200
	r.Response = map[string]interface{}{"value": strings.ToUpper(value)}
	return
}

// ToLower returns the input lowercased. Does NOT trim — chain Trim first
// if needed.
// WSL: strings/ops/ops.ToLower(value: $TrimResult.response.value)
func (t *opsTransitions) ToLower(value string) (r domain.FlowStepResult) {
	r.Success = true
	r.StatusCode = 200
	r.Response = map[string]interface{}{"value": strings.ToLower(value)}
	return
}

// SplitNonEmpty splits a string by separator, trims whitespace from each part,
// and drops empty entries. Returns {items: [], count: 0} for empty/whitespace
// input. Used for parsing comma-separated event lists where empty entries
// should be skipped.
// WSL: strings/ops/ops.SplitNonEmpty(value: $json.events, separator: ",")
func (t *opsTransitions) SplitNonEmpty(value string, separator string) (r domain.FlowStepResult) {
	if separator == "" {
		separator = ","
	}
	parts := strings.Split(value, separator)
	out := make([]interface{}, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	r.Success = true
	r.StatusCode = 200
	r.Response = map[string]interface{}{"items": out, "count": len(out)}
	return
}

// ─── Validation primitives (fail with 400 on invalid input) ──────────

// AssertRequired trims the value and fails with 400 if empty. Returns
// the trimmed value on success. Replaces Go's `s = strings.TrimSpace(s);
// if s == "" { return fail(...) }` pattern.
// WSL: strings/ops/ops.AssertRequired(value: $json.name, code: "invalid_name", message: "name is required")
func (t *opsTransitions) AssertRequired(value string, code string, message string) (r domain.FlowStepResult) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		if code == "" {
			code = "invalid_input"
		}
		if message == "" {
			message = "value is required"
		}
		return fail(code, message)
	}
	r.Success = true
	r.StatusCode = 200
	r.Response = map[string]interface{}{"value": trimmed, "valid": true}
	return
}

// AssertRequiredUpper trims + uppercases the value and fails with 400 if
// empty. Returns the normalized (uppercase, trimmed) value on success.
// Combines Trim + ToUpper + AssertRequired for the common code-normalization
// pattern used by ValidateGroupCode, ValidatePermissionCode, ValidateSku, etc.
// WSL: strings/ops/ops.AssertRequiredUpper(value: $json.code, code: "invalid_code", message: "code is required")
func (t *opsTransitions) AssertRequiredUpper(value string, code string, message string) (r domain.FlowStepResult) {
	trimmed := strings.ToUpper(strings.TrimSpace(value))
	if trimmed == "" {
		if code == "" {
			code = "invalid_input"
		}
		if message == "" {
			message = "value is required"
		}
		return fail(code, message)
	}
	r.Success = true
	r.StatusCode = 200
	r.Response = map[string]interface{}{"value": trimmed, "valid": true}
	return
}

// AssertRequiredLower trims + lowercases the value and fails with 400 if
// empty. Returns the normalized (lowercase, trimmed) value on success.
// WSL: strings/ops/ops.AssertRequiredLower(value: $json.type, code: "invalid_type", message: "type is required")
func (t *opsTransitions) AssertRequiredLower(value string, code string, message string) (r domain.FlowStepResult) {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		if code == "" {
			code = "invalid_input"
		}
		if message == "" {
			message = "value is required"
		}
		return fail(code, message)
	}
	r.Success = true
	r.StatusCode = 200
	r.Response = map[string]interface{}{"value": trimmed, "valid": true}
	return
}

// AssertMinLength fails with 400 if the value's rune count is less than min.
// Does NOT trim — chain AssertRequired first if trimming is needed.
// WSL: strings/ops/ops.AssertMinLength(value: $json.password, min: 6, code: "password_too_short", message: "password must be at least 6 characters")
func (t *opsTransitions) AssertMinLength(value string, min interface{}, code string, message string) (r domain.FlowStepResult) {
	minLen := toInt(min)
	if len([]rune(value)) < minLen {
		if code == "" {
			code = "too_short"
		}
		if message == "" {
			message = fmt.Sprintf("value must be at least %d characters long", minLen)
		}
		return fail(code, message)
	}
	r.Success = true
	r.StatusCode = 200
	r.Response = map[string]interface{}{"value": value, "valid": true}
	return
}

// AssertExactLength fails with 400 if the value's rune count is not exactly n.
// Does NOT trim — chain AssertRequired first if trimming is needed.
// WSL: strings/ops/ops.AssertExactLength(value: $code, n: 3, code: "invalid_code", message: "code must be 3 letters")
func (t *opsTransitions) AssertExactLength(value string, n interface{}, code string, message string) (r domain.FlowStepResult) {
	want := toInt(n)
	if len([]rune(value)) != want {
		if code == "" {
			code = "invalid_length"
		}
		if message == "" {
			message = fmt.Sprintf("value must be exactly %d characters long", want)
		}
		return fail(code, message)
	}
	r.Success = true
	r.StatusCode = 200
	r.Response = map[string]interface{}{"value": value, "valid": true}
	return
}

// AssertHasPrefix fails with 400 if the value does not start with any of the
// given prefixes. Accepts a single prefix or a comma-separated list of
// accepted prefixes (e.g. "http://,https://"). Case-sensitive.
// WSL: strings/ops/ops.AssertHasPrefix(value: $json.url, prefixes: "http://,https://", code: "invalid_url", message: "url must start with http:// or https://")
func (t *opsTransitions) AssertHasPrefix(value string, prefixes string, code string, message string) (r domain.FlowStepResult) {
	value = strings.TrimSpace(value)
	if value == "" {
		if code == "" {
			code = "invalid_input"
		}
		if message == "" {
			message = "value is required"
		}
		return fail(code, message)
	}
	accepted := strings.Split(prefixes, ",")
	for _, p := range accepted {
		p = strings.TrimSpace(p)
		if p != "" && strings.HasPrefix(value, p) {
			r.Success = true
			r.StatusCode = 200
			r.Response = map[string]interface{}{"value": value, "valid": true}
			return
		}
	}
	if code == "" {
		code = "invalid_prefix"
	}
	if message == "" {
		message = fmt.Sprintf("value must start with one of: %s", prefixes)
	}
	return fail(code, message)
}

// AssertOneOf fails with 400 if the (trimmed) value is not in the given list.
// Accepts a comma-separated list of accepted values. Case-sensitive —
// pre-normalize (ToUpper/ToLower) if case-insensitive match is needed.
// WSL: strings/ops/ops.AssertOneOf(value: $type, list: "asset,liability,equity,revenue,expense", code: "invalid_type", message: "type must be one of: ...")
func (t *opsTransitions) AssertOneOf(value string, list string, code string, message string) (r domain.FlowStepResult) {
	value = strings.TrimSpace(value)
	if value == "" {
		if code == "" {
			code = "invalid_input"
		}
		if message == "" {
			message = "value is required"
		}
		return fail(code, message)
	}
	accepted := strings.Split(list, ",")
	for _, a := range accepted {
		if strings.TrimSpace(a) == value {
			r.Success = true
			r.StatusCode = 200
			r.Response = map[string]interface{}{"value": value, "valid": true}
			return
		}
	}
	if code == "" {
		code = "invalid_value"
	}
	if message == "" {
		message = fmt.Sprintf("value must be one of: %s", list)
	}
	return fail(code, message)
}

// ─── Query primitives (return booleans, never fail) ─────────────────

// Contains reports whether value contains substring. Case-sensitive.
// WSL: strings/ops/ops.Contains(value: $json.name, substring: "acme")
func (t *opsTransitions) Contains(value string, substring string) (r domain.FlowStepResult) {
	r.Success = true
	r.StatusCode = 200
	r.Response = map[string]interface{}{"contains": strings.Contains(value, substring)}
	return
}

// Equals reports whether value equals other, case-sensitive. Useful for
// WSL workflows that need a boolean comparison result (WSL has no ==
// operator in action args).
// WSL: strings/ops/ops.Equals(value: $element.id, other: $url.locationId)
func (t *opsTransitions) Equals(value string, other string) (r domain.FlowStepResult) {
	r.Success = true
	r.StatusCode = 200
	r.Response = map[string]interface{}{"equals": value == other}
	return
}

// AssertNotEquals fails with 409 if value equals other. Returns the value
// on success. Used to short-circuit a workflow when two values are the same
// (e.g. same-currency conversion) by routing on error to the short-circuit
// response state. The "error" path IS the same-currency case.
// WSL: strings/ops/ops.AssertNotEquals(value: $from, other: $to, code: "same_currency", message: "from and to must differ")
func (t *opsTransitions) AssertNotEquals(value interface{}, other interface{}, code string, message string) (r domain.FlowStepResult) {
	vs := strings.TrimSpace(fmt.Sprintf("%v", value))
	os := strings.TrimSpace(fmt.Sprintf("%v", other))
	if vs == os {
		if code == "" {
			code = "same_value"
		}
		if message == "" {
			message = "values must differ"
		}
		return fail(code, message)
	}
	r.Success = true
	r.StatusCode = 200
	r.Response = map[string]interface{}{"value": vs, "other": os, "valid": true}
	return
}

// AssertEquals fails with 409 if value does NOT equal other. Returns the
// value on success. This is the WSL equivalent of an `if x == y` branch:
// the on-success path continues when values match; the on-error path
// handles the mismatch case. Together with AssertNotEquals, this gives
// WSL workflows a way to branch on equality with a specific HTTP status
// and error message on the failure path.
// (Historical note: this was written when `on success when` was broken —
// Gotcha 6, fixed July 2026 on engine `acme-dev`. The primitive is kept
// because it's a clean reusable way to assert equality with a custom
// error response; for pure boolean branching, `on success when` now works.)
// WSL: strings/ops/ops.AssertEquals(value: $json.state, other: "open", code: "wrong_state", message: "must be open")
func (t *opsTransitions) AssertEquals(value interface{}, other interface{}, code string, message string) (r domain.FlowStepResult) {
	vs := strings.TrimSpace(fmt.Sprintf("%v", value))
	os := strings.TrimSpace(fmt.Sprintf("%v", other))
	if vs != os {
		if code == "" {
			code = "value_mismatch"
		}
		if message == "" {
			message = "values must match"
		}
		return fail(code, message)
	}
	r.Success = true
	r.StatusCode = 200
	r.Response = map[string]interface{}{"value": vs, "other": os, "valid": true}
	return
}

// Switch returns the value mapped from `mapping[value]`, or `defaultVal` if
// the key is not present. Enables WSL workflows to express Go's `switch`
// statement as a map lookup. mapping is a WSL map literal (JSON object).
// WSL: strings/ops/ops.Switch(value: $json.type, mapping: {asset:debit, expense:debit, liability:credit, equity:credit, revenue:credit}, defaultVal: "")
func (t *opsTransitions) Switch(value string, mapping map[string]interface{}, defaultVal string) (r domain.FlowStepResult) {
	r.Success = true
	r.StatusCode = 200
	if mapping == nil {
		r.Response = map[string]interface{}{"value": defaultVal}
		return
	}
	if v, ok := mapping[value]; ok {
		r.Response = map[string]interface{}{"value": v}
		return
	}
	r.Response = map[string]interface{}{"value": defaultVal}
	return
}

// JoinNonEmpty joins non-empty string values from a slice with a separator,
// skipping empty/whitespace-only entries. Used to build pipe-separated ID
// chains (e.g. journalMoveId = "captureId|refundId") without conditional
// logic. If all values are empty, returns {value: ""}. If only one value
// is non-empty, returns it without the separator.
// WSL: strings/ops/ops.JoinNonEmpty(separator: "|", values: [$existing, $moveId])
func (t *opsTransitions) JoinNonEmpty(separator string, values []interface{}) (r domain.FlowStepResult) {
	r.Success = true
	r.StatusCode = 200
	var parts []string
	for _, v := range values {
		s, ok := v.(string)
		if !ok {
			continue
		}
		if s = strings.TrimSpace(s); s != "" {
			parts = append(parts, s)
		}
	}
	r.Response = map[string]interface{}{"value": strings.Join(parts, separator)}
	return
}

// ─── helpers ─────────────────────────────────────────────────────────

// toInt coerces interface{} to int. WSL passes JSON numbers as float64
// and integer literals may arrive as string; both are handled here.
func toInt(v interface{}) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case string:
		n := 0
		fmt.Sscanf(strings.TrimSpace(x), "%d", &n)
		return n
	}
	return 0
}
