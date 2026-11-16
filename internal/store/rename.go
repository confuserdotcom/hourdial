package store

import "fmt"

// labelColumn maps a caller's column name to the constant spelled in SQL,
// so only these literals are ever spliced into a query.
// It returns "" for anything else.
func labelColumn(name string) string {
	switch name {
	case "category":
		return "category"
	case "project":
		return "project"
	}
	return ""
}

// RenameLabel renames a category or project ("category" or "project") from
// one name to another, case-insensitively, in manual overrides, away labels,
// kept labels, timeline edits and cached suggestions. Renamed away rows get atMS so caches see the change.
func (s *Store) RenameLabel(column, from, to string, atMS int64) error {
	col := labelColumn(column)
	if col == "" {
		return fmt.Errorf("store: rename label: column %q", column)
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range []struct {
		sql  string
		args []any
	}{
		{`UPDATE override SET ` + col + ` = ? WHERE lower(` + col + `) = lower(?)`, []any{to, from}},
		{`UPDATE away SET ` + col + ` = ?, at_ms = ? WHERE lower(` + col + `) = lower(?)`, []any{to, atMS, from}},
		{`UPDATE kept SET ` + col + ` = ? WHERE lower(` + col + `) = lower(?)`, []any{to, from}},
		{`UPDATE edit SET ` + col + ` = ? WHERE lower(` + col + `) = lower(?)`, []any{to, from}},
		{`UPDATE suggestion SET ` + col + ` = ? WHERE lower(` + col + `) = lower(?)`, []any{to, from}},
	} {
		if _, err := tx.Exec(q.sql, q.args...); err != nil {
			return fmt.Errorf("store: rename label: %w", err)
		}
	}
	return tx.Commit()
}

// LabelUses counts the rows that name a category or project, as
// RenameLabel would rename them, except cached suggestions.
func (s *Store) LabelUses(column, name string) (int, error) {
	col := labelColumn(column)
	if col == "" {
		return 0, fmt.Errorf("store: label uses: column %q", column)
	}
	n := 0
	for _, table := range []string{"override", "away", "kept", "edit"} {
		var c int
		if err := s.db.QueryRow(`SELECT COUNT(*) FROM `+table+` WHERE lower(`+col+`) = lower(?)`, name).Scan(&c); err != nil {
			return 0, fmt.Errorf("store: label uses: %w", err)
		}
		n += c
	}
	return n, nil
}
