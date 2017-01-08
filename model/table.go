package model

import (
	"bytes"
	"io"

	"github.com/schemalex/schemalex/internal/util"
)

func NewTable(name string) Table {
	return &table{
		name: name,
	}
}

func (t *table) ID() string {
	return "table#" + t.name
}

func (t *table) LookupColumn(name string) (TableColumn, bool) {
	for col := range t.Columns() {
		if col.ID() == name {
			return col, true
		}
	}
	return nil, false
}

func (t *table) LookupIndex(id string) (Index, bool) {
	for idx := range t.Indexes() {
		if idx.ID() == id {
			return idx, true
		}
	}
	return nil, false
}

func (t *table) AddColumn(v TableColumn) {
	t.columns = append(t.columns, v)
}

func (t *table) AddIndex(v Index) {
	t.indexes = append(t.indexes, v)
}

func (t *table) AddOption(v TableOption) {
	t.options = append(t.options, v)
}

func (t *table) Name() string {
	return t.name
}

func (t *table) IsIfNotExists() bool {
	return t.ifnotexists
}

func (t *table) IsTemporary() bool {
	return t.temporary
}

func (t *table) SetIfNotExists(v bool) {
	t.ifnotexists = v
}

func (t *table) SetTemporary(v bool) {
	t.temporary = v
}

func (t *table) Columns() chan TableColumn {
	ch := make(chan TableColumn, len(t.columns))
	for _, col := range t.columns {
		ch <- col
	}
	close(ch)
	return ch
}

func (t *table) Indexes() chan Index {
	ch := make(chan Index, len(t.indexes))
	for _, idx := range t.indexes {
		ch <- idx
	}
	close(ch)
	return ch
}

func (t *table) Options() chan TableOption {
	ch := make(chan TableOption, len(t.options))
	for _, idx := range t.options {
		ch <- idx
	}
	close(ch)
	return ch
}

func (t *table) WriteTo(dst io.Writer) (int64, error) {
	var b bytes.Buffer

	b.WriteString("CREATE")
	if t.IsTemporary() {
		b.WriteString(" TEMPORARY")
	}

	b.WriteString(" TABLE")
	if t.IsIfNotExists() {
		b.WriteString(" IF NOT EXISTS")
	}

	b.WriteByte(' ')
	b.WriteString(util.Backquote(t.Name()))
	b.WriteString(" (")

	colch := t.Columns()
	idxch := t.Indexes()
	fields := make([]io.WriterTo, 0, len(colch)+len(idxch))
	for col := range colch {
		fields = append(fields, col)
	}
	for idx := range idxch {
		fields = append(fields, idx)
	}

	for i, stmt := range fields {
		b.WriteByte('\n')
		if _, err := stmt.WriteTo(&b); err != nil {
			return 0, err
		}
		if i < len(fields)-1 {
			b.WriteByte(',')
		}
	}

	b.WriteString("\n)")

	optch := t.Options()
	if l := len(optch); l > 0 {
		b.WriteByte(' ')
		var i int
		for option := range optch {
			if _, err := option.WriteTo(&b); err != nil {
				return 0, err
			}

			if i < l-1 {
				b.WriteString(", ")
			}
			i++
		}
	}

	return b.WriteTo(dst)
}

func NewTableOption(k, v string) TableOption {
	return &tableopt{
		key:   k,
		value: v,
	}
}

func (t *tableopt) ID() string    { return "tableopt#" + t.key }
func (t *tableopt) Key() string   { return t.key }
func (t *tableopt) Value() string { return t.value }

func (t *tableopt) WriteTo(dst io.Writer) (int64, error) {
	var buf bytes.Buffer
	buf.WriteString(t.Key())
	buf.WriteString(" = ")
	buf.WriteString(t.Value())

	return buf.WriteTo(dst)
}
