package transitions

import (
	"testing"
)

func TestTrim(t *testing.T) {
	r := (&opsTransitions{}).Trim("  hello  ")
	if !r.Success {
		t.Fatalf("Trim failed: %v", r.Error)
	}
	if v := r.Response.(map[string]interface{})["value"]; v != "hello" {
		t.Errorf("value = %v, want hello", v)
	}
}

func TestToUpper(t *testing.T) {
	r := (&opsTransitions{}).ToUpper("abc")
	if v := r.Response.(map[string]interface{})["value"]; v != "ABC" {
		t.Errorf("value = %v, want ABC", v)
	}
}

func TestToLower(t *testing.T) {
	r := (&opsTransitions{}).ToLower("ABC")
	if v := r.Response.(map[string]interface{})["value"]; v != "abc" {
		t.Errorf("value = %v, want abc", v)
	}
}

func TestSplitNonEmptyFiltersEmpties(t *testing.T) {
	r := (&opsTransitions{}).SplitNonEmpty("a, , b, , c", ",")
	items := r.Response.(map[string]interface{})["items"].([]interface{})
	if len(items) != 3 {
		t.Fatalf("expected 3 non-empty items, got %d", len(items))
	}
	if items[0] != "a" || items[1] != "b" || items[2] != "c" {
		t.Errorf("items = %v, want [a b c]", items)
	}
}

func TestSplitNonEmptyEmptyInput(t *testing.T) {
	r := (&opsTransitions{}).SplitNonEmpty("", ",")
	items := r.Response.(map[string]interface{})["items"].([]interface{})
	if len(items) != 0 {
		t.Errorf("expected 0 items for empty input, got %d", len(items))
	}
}

func TestSplitNonEmptySingleValue(t *testing.T) {
	r := (&opsTransitions{}).SplitNonEmpty("*", ",")
	items := r.Response.(map[string]interface{})["items"].([]interface{})
	if len(items) != 1 || items[0] != "*" {
		t.Errorf("expected [*], got %v", items)
	}
}

func TestAssertRequiredSuccess(t *testing.T) {
	r := (&opsTransitions{}).AssertRequired("  alice  ", "invalid_name", "name is required")
	if !r.Success {
		t.Fatalf("expected success: %v", r.Error)
	}
	if v := r.Response.(map[string]interface{})["value"]; v != "alice" {
		t.Errorf("value = %v, want alice", v)
	}
}

func TestAssertRequiredEmpty(t *testing.T) {
	r := (&opsTransitions{}).AssertRequired("   ", "invalid_name", "name is required")
	if r.Success {
		t.Error("expected failure for empty input")
	}
	if r.StatusCode != 400 {
		t.Errorf("status = %d, want 400", r.StatusCode)
	}
}

func TestAssertRequiredUpperSuccess(t *testing.T) {
	r := (&opsTransitions{}).AssertRequiredUpper("  admins  ", "invalid_code", "code is required")
	if !r.Success {
		t.Fatalf("expected success: %v", r.Error)
	}
	if v := r.Response.(map[string]interface{})["value"]; v != "ADMINS" {
		t.Errorf("value = %v, want ADMINS", v)
	}
}

func TestAssertRequiredUpperEmpty(t *testing.T) {
	r := (&opsTransitions{}).AssertRequiredUpper("", "invalid_code", "code is required")
	if r.Success {
		t.Error("expected failure for empty input")
	}
}

func TestAssertRequiredLowerSuccess(t *testing.T) {
	r := (&opsTransitions{}).AssertRequiredLower("  Asset  ", "invalid_type", "type is required")
	if !r.Success {
		t.Fatalf("expected success: %v", r.Error)
	}
	if v := r.Response.(map[string]interface{})["value"]; v != "asset" {
		t.Errorf("value = %v, want asset", v)
	}
}

func TestAssertMinLengthSuccess(t *testing.T) {
	r := (&opsTransitions{}).AssertMinLength("secret1", 6, "password_too_short", "password must be at least 6 characters")
	if !r.Success {
		t.Fatalf("expected success: %v", r.Error)
	}
}

func TestAssertMinLengthTooShort(t *testing.T) {
	r := (&opsTransitions{}).AssertMinLength("abc", 6, "password_too_short", "password must be at least 6 characters")
	if r.Success {
		t.Error("expected failure for short input")
	}
	if r.StatusCode != 400 {
		t.Errorf("status = %d, want 400", r.StatusCode)
	}
}

func TestAssertExactLengthSuccess(t *testing.T) {
	r := (&opsTransitions{}).AssertExactLength("USD", 3, "invalid_code", "code must be 3 letters")
	if !r.Success {
		t.Fatalf("expected success: %v", r.Error)
	}
}

func TestAssertExactLengthWrong(t *testing.T) {
	r := (&opsTransitions{}).AssertExactLength("US", 3, "invalid_code", "code must be 3 letters")
	if r.Success {
		t.Error("expected failure for wrong length")
	}
}

func TestAssertHasPrefixSuccess(t *testing.T) {
	r := (&opsTransitions{}).AssertHasPrefix("https://example.com", "http://,https://", "invalid_url", "url must start with http:// or https://")
	if !r.Success {
		t.Fatalf("expected success: %v", r.Error)
	}
}

func TestAssertHasPrefixFail(t *testing.T) {
	r := (&opsTransitions{}).AssertHasPrefix("ftp://example.com", "http://,https://", "invalid_url", "url must start with http:// or https://")
	if r.Success {
		t.Error("expected failure for bad prefix")
	}
}

func TestAssertHasPrefixEmpty(t *testing.T) {
	r := (&opsTransitions{}).AssertHasPrefix("   ", "http://,https://", "invalid_url", "url is required")
	if r.Success {
		t.Error("expected failure for empty input")
	}
}

func TestAssertOneOfSuccess(t *testing.T) {
	r := (&opsTransitions{}).AssertOneOf("asset", "asset,liability,equity,revenue,expense", "invalid_type", "type must be one of: ...")
	if !r.Success {
		t.Fatalf("expected success: %v", r.Error)
	}
}

func TestAssertOneOfFail(t *testing.T) {
	r := (&opsTransitions{}).AssertOneOf("inventory", "asset,liability,equity,revenue,expense", "invalid_type", "type must be one of: ...")
	if r.Success {
		t.Error("expected failure for value not in list")
	}
}

func TestContainsTrue(t *testing.T) {
	r := (&opsTransitions{}).Contains("hello world", "world")
	if c := r.Response.(map[string]interface{})["contains"]; c != true {
		t.Errorf("contains = %v, want true", c)
	}
}

func TestContainsFalse(t *testing.T) {
	r := (&opsTransitions{}).Contains("hello world", "acme")
	if c := r.Response.(map[string]interface{})["contains"]; c != false {
		t.Errorf("contains = %v, want false", c)
	}
}

func TestJoinNonEmptyBothPresent(t *testing.T) {
	r := (&opsTransitions{}).JoinNonEmpty("|", []interface{}{"captureId", "refundId"})
	if v := r.Response.(map[string]interface{})["value"]; v != "captureId|refundId" {
		t.Errorf("value = %v, want 'captureId|refundId'", v)
	}
}

func TestJoinNonEmptyFirstEmpty(t *testing.T) {
	r := (&opsTransitions{}).JoinNonEmpty("|", []interface{}{"", "refundId"})
	if v := r.Response.(map[string]interface{})["value"]; v != "refundId" {
		t.Errorf("value = %v, want 'refundId'", v)
	}
}

func TestJoinNonEmptyAllEmpty(t *testing.T) {
	r := (&opsTransitions{}).JoinNonEmpty("|", []interface{}{"", "  "})
	if v := r.Response.(map[string]interface{})["value"]; v != "" {
		t.Errorf("value = %v, want ''", v)
	}
}

func TestJoinNonEmptySingleValue(t *testing.T) {
	r := (&opsTransitions{}).JoinNonEmpty("|", []interface{}{"onlyOne", ""})
	if v := r.Response.(map[string]interface{})["value"]; v != "onlyOne" {
		t.Errorf("value = %v, want 'onlyOne'", v)
	}
}
