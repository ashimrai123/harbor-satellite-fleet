package client

import (
	"encoding/json"
	"fmt"
	"time"
)

// NullTime handles SQLC's sql.NullTime JSON format: {"Time":"...","Valid":bool}
type NullTime struct {
	Time  time.Time
	Valid bool
}

func (nt *NullTime) UnmarshalJSON(data []byte) error {
	var raw struct {
		Time  time.Time `json:"Time"`
		Valid bool      `json:"Valid"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		// Try plain time string
		var t time.Time
		if err2 := json.Unmarshal(data, &t); err2 == nil {
			nt.Time = t
			nt.Valid = true
			return nil
		}
		return err
	}
	nt.Time = raw.Time
	nt.Valid = raw.Valid
	return nil
}

func (nt NullTime) String() string {
	if !nt.Valid {
		return "-"
	}
	return nt.Time.Format("2006-01-02 15:04:05")
}

// NullString handles SQLC's sql.NullString JSON format: {"String":"...","Valid":bool}
type NullString struct {
	String string
	Valid  bool
}

func (ns *NullString) UnmarshalJSON(data []byte) error {
	var raw struct {
		String string `json:"String"`
		Valid  bool   `json:"Valid"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		// Try plain string
		var s string
		if err2 := json.Unmarshal(data, &s); err2 == nil {
			ns.String = s
			ns.Valid = true
			return nil
		}
		return err
	}
	ns.String = raw.String
	ns.Valid = raw.Valid
	return nil
}

func (ns NullString) Val() string {
	if !ns.Valid || ns.String == "" {
		return "-"
	}
	return ns.String
}

// NullInt32 handles SQLC's sql.NullInt32 JSON format.
type NullInt32 struct {
	Int32 int32
	Valid bool
}

func (ni *NullInt32) UnmarshalJSON(data []byte) error {
	var raw struct {
		Int32 int32 `json:"Int32"`
		Valid bool  `json:"Valid"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		var v int32
		if err2 := json.Unmarshal(data, &v); err2 == nil {
			ni.Int32 = v
			ni.Valid = true
			return nil
		}
		return err
	}
	ni.Int32 = raw.Int32
	ni.Valid = raw.Valid
	return nil
}

func (ni NullInt32) String() string {
	if !ni.Valid {
		return "-"
	}
	return fmt.Sprintf("%d", ni.Int32)
}

// NullInt64 handles SQLC's sql.NullInt64 JSON format.
type NullInt64 struct {
	Int64 int64
	Valid bool
}

func (ni *NullInt64) UnmarshalJSON(data []byte) error {
	var raw struct {
		Int64 int64 `json:"Int64"`
		Valid bool  `json:"Valid"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		var v int64
		if err2 := json.Unmarshal(data, &v); err2 == nil {
			ni.Int64 = v
			ni.Valid = true
			return nil
		}
		return err
	}
	ni.Int64 = raw.Int64
	ni.Valid = raw.Valid
	return nil
}

func (ni NullInt64) String() string {
	if !ni.Valid {
		return "-"
	}
	return fmt.Sprintf("%d", ni.Int64)
}

func (ni NullInt64) Bytes() string {
	if !ni.Valid {
		return "-"
	}
	mb := float64(ni.Int64) / (1024 * 1024)
	return fmt.Sprintf("%.1f MB", mb)
}

func (ni NullInt64) Ms() string {
	if !ni.Valid {
		return "-"
	}
	return fmt.Sprintf("%d ms", ni.Int64)
}
