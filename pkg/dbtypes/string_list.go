// Package dbtypes holds small column types shared by feature models.
package dbtypes

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// StringList is a list of strings stored as a JSONB array. A nil list is
// written as [] so the column never holds JSON null.
type StringList []string

func (l StringList) Value() (driver.Value, error) {
	if l == nil {
		return "[]", nil
	}
	b, err := json.Marshal([]string(l))
	return string(b), err
}

func (l *StringList) Scan(src any) error {
	var raw []byte
	switch v := src.(type) {
	case nil:
		*l = StringList{}
		return nil
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return errors.New("dbtypes.StringList: unsupported column type")
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return err
	}
	if out == nil {
		out = []string{}
	}
	*l = out
	return nil
}

// GormDataType tells GORM the column type for migrations and casts.
func (StringList) GormDataType() string { return "jsonb" }

// Clean trims entries and drops blanks, keeping order.
func (l StringList) Clean() StringList {
	out := make(StringList, 0, len(l))
	for _, s := range l {
		if t := trim(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func trim(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
