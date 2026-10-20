package toml

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Value represents a parsed TOML primitive or table
type Value struct {
	Type  string      // "string", "int", "bool", "float", "array", "table"
	Raw   interface{}
}

func (v *Value) String() string {
	if v == nil {
		return ""
	}
	if s, ok := v.Raw.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v.Raw)
}

func (v *Value) Int() (int64, error) {
	if v == nil {
		return 0, errors.New("nil value")
	}
	switch val := v.Raw.(type) {
	case int64:
		return val, nil
	case int:
		return int64(val), nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v.Raw)
	}
}

func (v *Value) Bool() (bool, error) {
	if v == nil {
		return false, errors.New("nil value")
	}
	switch val := v.Raw.(type) {
	case bool:
		return val, nil
	case string:
		return strconv.ParseBool(val)
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v.Raw)
	}
}

func (v *Value) Table() map[string]*Value {
	if v == nil {
		return nil
	}
	if t, ok := v.Raw.(map[string]*Value); ok {
		return t
	}
	return nil
}

// Document represents a structured parsed TOML tree
type Document struct {
	Root map[string]*Value
}

// Parse parses a TOML configuration string into a Document
func Parse(input string) (*Document, error) {
	doc := &Document{
		Root: make(map[string]*Value),
	}

	lines := strings.Split(input, "\n")
	currentSection := doc.Root

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section header [table]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			secName := strings.TrimSpace(line[1 : len(line)-1])
			if secName == "" {
				return nil, fmt.Errorf("line %d: empty section name", lineNum+1)
			}
			sectionTable := make(map[string]*Value)
			doc.Root[secName] = &Value{Type: "table", Raw: sectionTable}
			currentSection = sectionTable
			continue
		}

		// Key-value pair: key = value
		eqIdx := strings.Index(line, "=")
		if eqIdx == -1 {
			return nil, fmt.Errorf("line %d: missing '=' in key-value statement", lineNum+1)
		}

		key := strings.TrimSpace(line[:eqIdx])
		rawVal := strings.TrimSpace(line[eqIdx+1:])

		// Strip inline comment
		if commIdx := strings.Index(rawVal, "#"); commIdx != -1 {
			// Ensure comment is not inside string quotes
			if !isInsideQuotes(rawVal, commIdx) {
				rawVal = strings.TrimSpace(rawVal[:commIdx])
			}
		}

		val, err := parseValue(rawVal)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum+1, err)
		}
		currentSection[key] = val
	}

	return doc, nil
}

func isInsideQuotes(s string, idx int) bool {
	quoteCount := 0
	for i := 0; i < idx; i++ {
		if s[i] == '"' || s[i] == '\'' {
			quoteCount++
		}
	}
	return (quoteCount % 2) != 0
}

func parseValue(s string) (*Value, error) {
	if len(s) == 0 {
		return &Value{Type: "string", Raw: ""}, nil
	}

	// String: "..." or '...'
	if (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) ||
		(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		return &Value{Type: "string", Raw: s[1 : len(s)-1]}, nil
	}

	// Boolean
	if s == "true" {
		return &Value{Type: "bool", Raw: true}, nil
	}
	if s == "false" {
		return &Value{Type: "bool", Raw: false}, nil
	}

	// Integer
	if i, err := strconv.ParseInt(s, 0, 64); err == nil {
		return &Value{Type: "int", Raw: i}, nil
	}

	// Float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return &Value{Type: "float", Raw: f}, nil
	}

	// Default fallback to unquoted string
	return &Value{Type: "string", Raw: s}, nil
}

// Get extracts a value from the TOML document using dot-notation (e.g. "server.port")
func (d *Document) Get(path string) *Value {
	parts := strings.Split(path, ".")
	curr := d.Root

	for i, part := range parts {
		val, exists := curr[part]
		if !exists {
			return nil
		}
		if i == len(parts)-1 {
			return val
		}
		if val.Type == "table" {
			curr = val.Table()
		} else {
			return nil
		}
	}
	return nil
}

// Encode converts a document map back to valid TOML string
func (d *Document) Encode() string {
	var sb strings.Builder

	// Top level keys first
	for k, v := range d.Root {
		if v.Type != "table" {
			sb.WriteString(fmt.Sprintf("%s = %s\n", k, formatValue(v)))
		}
	}

	// Tables next
	for k, v := range d.Root {
		if v.Type == "table" {
			sb.WriteString(fmt.Sprintf("\n[%s]\n", k))
			for subK, subV := range v.Table() {
				sb.WriteString(fmt.Sprintf("%s = %s\n", subK, formatValue(subV)))
			}
		}
	}
	return sb.String()
}

func formatValue(v *Value) string {
	if v == nil {
		return "\"\""
	}
	switch v.Type {
	case "string":
		return fmt.Sprintf("%q", v.Raw)
	default:
		return fmt.Sprintf("%v", v.Raw)
	}
}
