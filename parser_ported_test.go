package nameparser_test

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/ysmood/got"

	"github.com/ysmood/nameparser"
)

type upstreamAssertion struct {
	Class           string          `json:"class"`
	Test            string          `json:"test"`
	Line            int             `json:"line"`
	Expr            string          `json:"expr"`
	Expected        any             `json:"expected"`
	Original        any             `json:"original"`
	Snapshot        map[string]any  `json:"snapshot"`
	ParseStable     bool            `json:"parse_stable"`
	ExpectedFailure bool            `json:"expected_failure"`
	Config          assertionConfig `json:"config"`
	Attr            any             `json:"attr,omitempty"`
}

type assertionConfig struct {
	StringFormat                 any `json:"string_format"`
	InitialsFormat               any `json:"initials_format"`
	InitialsDelimiter            any `json:"initials_delimiter"`
	EmptyAttributeDefault        any `json:"empty_attribute_default"`
	CapitalizeName               any `json:"capitalize_name"`
	ForceMixedCaseCapitalization any `json:"force_mixed_case_capitalization"`
	EmojiRegexEnabled            any `json:"emoji_regex_enabled"`
}

func TestUpstreamAssertions(t *testing.T) {
	g := got.T(t)

	data, err := os.ReadFile("testdata/upstream_assertions.json")
	g.E(err)

	var records []upstreamAssertion
	g.E(json.Unmarshal(data, &records))

	executed := 0
	skippedExpectedFailure := 0
	for i, rec := range records {
		if rec.ExpectedFailure {
			skippedExpectedFailure++
			continue
		}

		hn, err := buildNameForAssertion(rec)
		if err != nil {
			t.Fatalf("record %d %s.%s line %d build error: %v", i, rec.Class, rec.Test, rec.Line, err)
		}

		actual, err := evalExpression(hn, rec)
		if err != nil {
			t.Fatalf("record %d %s.%s line %d eval error: %v", i, rec.Class, rec.Test, rec.Line, err)
		}

		if !reflect.DeepEqual(normalizeValue(actual), normalizeValue(rec.Expected)) {
			t.Fatalf("record %d %s.%s line %d expr=%s expected=%#v actual=%#v original=%#v", i, rec.Class, rec.Test, rec.Line, rec.Expr, rec.Expected, actual, rec.Original)
		}
		executed++
	}

	// Upstream tests emit one assertion record per `self.m(...)` call.
	g.Eq(executed, len(records)-skippedExpectedFailure)
}

func buildNameForAssertion(rec upstreamAssertion) (*nameparser.HumanName, error) {
	var (
		hn  *nameparser.HumanName
		err error
	)

	if rec.ParseStable {
		hn, err = newFromOriginal(rec.Original)
	} else {
		opts := nameparser.HumanNameOptions{
			Title:    snapshotValue(rec.Snapshot, "title"),
			First:    snapshotValue(rec.Snapshot, "first"),
			Middle:   snapshotValue(rec.Snapshot, "middle"),
			Last:     snapshotValue(rec.Snapshot, "last"),
			Suffix:   snapshotValue(rec.Snapshot, "suffix"),
			Nickname: snapshotValue(rec.Snapshot, "nickname"),
		}
		hn, err = nameparser.NewWithOptions("", opts)
	}

	if err != nil {
		return nil, err
	}

	applyConfig(hn, rec.Config)
	return hn, nil
}

func applyConfig(hn *nameparser.HumanName, cfg assertionConfig) {
	if s, ok := cfg.StringFormat.(string); ok {
		hn.StringFormat = s
	}
	if s, ok := cfg.InitialsFormat.(string); ok {
		hn.InitialsFormat = s
	}
	if s, ok := cfg.InitialsDelimiter.(string); ok {
		hn.InitialsDelimiter = s
	}
	if s, ok := cfg.EmptyAttributeDefault.(string); ok {
		hn.C.EmptyAttributeDefault = s
	} else {
		// Python `None` empty defaults are represented as empty strings in this port.
		hn.C.EmptyAttributeDefault = ""
	}
	if b, ok := cfg.CapitalizeName.(bool); ok {
		hn.C.CapitalizeName = b
	}
	if b, ok := cfg.ForceMixedCaseCapitalization.(bool); ok {
		hn.C.ForceMixedCaseCapitalization = b
	}
	if b, ok := cfg.EmojiRegexEnabled.(bool); ok && !b {
		hn.C.Regexes.Emoji = nil
	}
	if hn.C.CapitalizeName {
		hn.Capitalize(nil)
	}
}

func snapshotValue(snapshot map[string]any, key string) any {
	if snapshot == nil {
		return nil
	}
	v, ok := snapshot[key]
	if !ok {
		return nil
	}
	if v == nil {
		return nil
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

func newFromOriginal(original any) (*nameparser.HumanName, error) {
	hn := nameparser.New("")
	switch v := original.(type) {
	case string:
		hn.SetFullName(v)
		return hn, nil
	case map[string]any:
		raw, ok := v["__bytes__"]
		if !ok {
			return nil, fmt.Errorf("unsupported original map payload: %#v", v)
		}
		arr, ok := raw.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid __bytes__ payload: %#v", raw)
		}
		buf := make([]byte, len(arr))
		for i, n := range arr {
			f, ok := n.(float64)
			if !ok {
				return nil, fmt.Errorf("invalid byte at index %d: %#v", i, n)
			}
			buf[i] = byte(int(f))
		}
		if err := hn.SetFullNameBytes(buf); err != nil {
			return nil, err
		}
		return hn, nil
	default:
		return nil, fmt.Errorf("unsupported original type %T", original)
	}
}

func evalExpression(hn *nameparser.HumanName, rec upstreamAssertion) (any, error) {
	switch rec.Expr {
	case "hn.first", "hh.first":
		return hn.First(), nil
	case "hn.last", "hh.last":
		return hn.Last(), nil
	case "hn.middle", "hh.middle":
		return hn.Middle(), nil
	case "hn.suffix":
		return hn.Suffix(), nil
	case "hn.title":
		return hn.Title(), nil
	case "hn.nickname":
		return hn.Nickname(), nil
	case "hn.initials()":
		return hn.Initials(), nil
	case "hn.initials_list()":
		return hn.InitialsList(), nil
	case "str(hn)", "u(hn)":
		return hn.String(), nil
	case "len(hn)":
		return hn.Len(), nil
	case "hn.full_name":
		return hn.FullName(), nil
	case "hn.surnames":
		return hn.Surnames(), nil
	case "hn.surnames_list":
		return hn.SurnamesList(), nil
	case "list(hn)":
		return hn.Members(), nil
	case "hn[1:]":
		return hn.MembersSlice(1, nil), nil
	case "hn[1:-2]":
		end := -2
		return hn.MembersSlice(1, &end), nil
	case "hn['first']":
		return hn.Member("first")
	case "hn['middle']":
		return hn.Member("middle")
	case "hn['last']":
		return hn.Member("last")
	case "hn['title']":
		return hn.Member("title")
	case "hn['suffix']":
		return hn.Member("suffix")
	case "getattr(hn, attr)":
		attr, ok := rec.Attr.(string)
		if !ok {
			return nil, fmt.Errorf("missing attr for getattr assertion: %#v", rec.Attr)
		}
		return hn.Member(attr)
	default:
		return nil, fmt.Errorf("unsupported expression: %q", rec.Expr)
	}
}

func normalizeValue(v any) any {
	switch x := v.(type) {
	case nil:
		// Python None assertions are adapted to empty string in this Go port.
		return ""
	case int:
		return x
	case float64:
		if float64(int(x)) == x {
			return int(x)
		}
		return x
	case string:
		return x
	case []string:
		out := make([]any, len(x))
		for i, item := range x {
			out[i] = normalizeValue(item)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, item := range x {
			out[i] = normalizeValue(item)
		}
		return out
	case map[string]any:
		if raw, ok := x["__bytes__"]; ok {
			return normalizeValue(raw)
		}
		out := map[string]any{}
		for k, v2 := range x {
			out[k] = normalizeValue(v2)
		}
		return out
	default:
		return fmt.Sprint(x)
	}
}
