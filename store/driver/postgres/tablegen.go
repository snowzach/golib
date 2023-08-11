package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/spf13/cast"
)

type DB interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func Generate[T any](t Table[T]) *Table[T] {

	if t.SelectFields == "" {
		t.SelectFields = t.GenerateSelectFields()
	}
	if t.GetByIDQuery == "" {
		t.GetByIDQuery = t.GenerateGetByIDQuery()
	}
	if t.DeleteByIDQuery == "" {
		t.DeleteByIDQuery = t.GenerateDeleteByIDQuery()
	}
	if t.InsertQuery == "" {
		t.InsertQuery = t.GenerateInsertQuery()
	}
	if t.UpdateQuery == "" {
		t.UpdateQuery = t.GenerateUpdateQuery()
	}
	if t.UpsertQuery == "" {
		t.UpsertQuery = t.GenerateUpsertQuery()
	}
	if t.Selector.Query == "" {
		t.Selector.Query = t.GenerateSelectorQuery()
	}

	return &t
}

func (t *Table[T]) GenerateSelectFields() string {

	var b strings.Builder
	for i, field := range t.Fields {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(t.Table)
		b.WriteString(".")
		b.WriteString(field.Name)
	}
	return b.String()

}

func (t *Table[T]) GenerateAdditionalFields(coalesce bool) string {

	var b strings.Builder
	for i, field := range t.Fields {
		if i > 0 {
			b.WriteString(",")
		}
		if coalesce {
			b.WriteString("COALESCE(")
		}
		b.WriteString(t.Table)
		b.WriteString(".")
		b.WriteString(field.Name)
		if coalesce {
			b.WriteString(",")
			switch v := field.NullVal.(type) {
			case string:
				b.WriteString("'")
				b.WriteString(v)
				b.WriteString("'")
			case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
				b.WriteString(cast.ToString(v))
			default:
				b.WriteString("'")
				b.WriteString(cast.ToString(v))
				b.WriteString("'")
			}
			b.WriteString(")")
		}
		b.WriteString(" AS \"")
		b.WriteString(strings.Trim(t.Table, `"`))
		b.WriteString(".")
		b.WriteString(strings.Trim(field.Name, `"`))
		b.WriteString("\"")
	}
	return b.String()
}

func (t *Table[T]) GenerateGetByIDQuery() string {

	var b strings.Builder
	b.WriteString("SELECT ")
	b.WriteString(t.SelectFields)
	if t.SelectAdditionalFields != "" {
		if t.SelectFields != "" {
			b.WriteString(",")
		}
		b.WriteString(t.SelectAdditionalFields)
	}
	b.WriteString(` FROM `)
	if t.Schema != "" {
		b.WriteString(t.Schema)
		b.WriteByte('.')
	}
	b.WriteString(t.Table)
	if t.Joins != "" {
		b.WriteString(" ")
		b.WriteString(t.Joins)
	}
	b.WriteString(` WHERE `)

	var idIndex int
	for _, field := range t.Fields {
		if field.ID {
			idIndex++
			if idIndex > 1 {
				b.WriteString(` AND `)
			}
			b.WriteString(t.Table)
			b.WriteString(".")
			b.WriteString(field.Name)
			b.WriteString(" = $")
			b.WriteString(strconv.Itoa(idIndex))
		}
	}
	return b.String()

}

func (t *Table[T]) GenerateGetByFieldsQuery(fields ...string) string {

	var b strings.Builder
	b.WriteString("SELECT ")
	b.WriteString(t.SelectFields)
	if t.SelectAdditionalFields != "" {
		if t.SelectFields != "" {
			b.WriteString(",")
		}
		b.WriteString(t.SelectAdditionalFields)
	}
	b.WriteString(` FROM `)
	if t.Schema != "" {
		b.WriteString(t.Schema)
		b.WriteByte('.')
	}
	b.WriteString(t.Table)
	if t.Joins != "" {
		b.WriteString(" ")
		b.WriteString(t.Joins)
	}
	b.WriteString(` WHERE `)

	var idIndex int
	for _, field := range fields {
		idIndex++
		if idIndex > 1 {
			b.WriteString(` AND `)
		}
		b.WriteString(t.Table)
		b.WriteString(".")
		b.WriteString(field)
		b.WriteString(" = $")
		b.WriteString(strconv.Itoa(idIndex))
	}
	return b.String()

}

func (t *Table[T]) GenerateDeleteByIDQuery() string {

	var b strings.Builder
	b.WriteString(`DELETE FROM `)
	if t.Schema != "" {
		b.WriteString(t.Schema)
		b.WriteByte('.')
	}
	b.WriteString(t.Table)
	b.WriteString(` WHERE `)

	var idIndex = 0
	for _, field := range t.Fields {
		if field.ID {
			idIndex++
			if idIndex > 1 {
				b.WriteString(` AND `)
			}
			b.WriteString(t.Table)
			b.WriteString(".")
			b.WriteString(field.Name)
			b.WriteString(" = $")
			b.WriteString(strconv.Itoa(idIndex))
		}
	}
	return b.String()

}

func (t *Table[T]) GenerateInsertQuery() string {

	var b strings.Builder
	var names []string
	var inserts []string
	var argCount int

	for _, field := range t.Fields {
		index := "$#"
		if field.Value != nil {
			argCount++
			index = "$" + strconv.Itoa(argCount)
		}
		if field.Insert != "" {
			names = append(names, field.Name)
			inserts = append(inserts, strings.ReplaceAll(field.Insert, Value, index))
		}
	}

	b.WriteString("WITH ")
	b.WriteString(t.Table)
	b.WriteString(" AS ( INSERT INTO ")
	if t.Schema != "" {
		b.WriteString(t.Schema)
		b.WriteString(".")
	}
	b.WriteString(t.Table)
	b.WriteString(" (")
	b.WriteString(strings.Join(names, ",")) // Fields
	b.WriteString(") VALUES(")
	b.WriteString(strings.Join(inserts, ",")) // Inserts
	b.WriteString(") RETURNING *")
	b.WriteString(") SELECT ")
	b.WriteString(t.SelectFields)
	if t.SelectAdditionalFields != "" {
		if t.SelectFields != "" {
			b.WriteString(",")
		}
		b.WriteString(t.SelectAdditionalFields)
	}
	b.WriteString(" FROM ")
	b.WriteString(t.Table)
	if t.Joins != "" {
		b.WriteString(" ")
		b.WriteString(t.Joins)
	}
	return b.String()

}

func (t *Table[T]) GenerateUpdateQuery() string {

	var b strings.Builder
	var updates []string
	var argCount int

	for _, field := range t.Fields {
		index := "$#"
		if field.Value != nil {
			argCount++
			index = "$" + strconv.Itoa(argCount)
		}
		if field.Update != "" {
			updates = append(updates, field.Name+" = "+strings.ReplaceAll(field.Update, Value, index))
		}
	}

	b.WriteString("WITH ")
	b.WriteString(t.Table)
	b.WriteString(" AS ( UPDATE ")
	if t.Schema != "" {
		b.WriteString(t.Schema)
		b.WriteString(".")
	}
	b.WriteString(t.Table)
	b.WriteString(" SET ")
	b.WriteString(strings.Join(updates, ",")) // Updates
	b.WriteString(` WHERE `)
	var idIndex = 0
	for _, field := range t.Fields {
		if field.ID {
			idIndex++
			if idIndex > 1 {
				b.WriteString(` AND `)
			}
			b.WriteString(t.Table)
			b.WriteString(".")
			b.WriteString(field.Name)
			b.WriteString(" = $")
			b.WriteString(strconv.Itoa(idIndex))
		}
	}
	b.WriteString(" RETURNING *")
	b.WriteString(") SELECT ")
	b.WriteString(t.SelectFields)
	if t.SelectAdditionalFields != "" {
		if t.SelectFields != "" {
			b.WriteString(",")
		}
		b.WriteString(t.SelectAdditionalFields)
	}
	b.WriteString(" FROM ")
	b.WriteString(t.Table)
	if t.Joins != "" {
		b.WriteString(" ")
		b.WriteString(t.Joins)
	}
	return b.String()

}

func (t *Table[T]) GenerateUpsertQuery() string {

	var b strings.Builder
	var names []string
	var inserts []string
	var updates []string
	var argCount int
	var ids []string

	for _, field := range t.Fields {
		index := "$#"
		if field.Value != nil {
			argCount++
			index = "$" + strconv.Itoa(argCount)
		}
		if field.Insert != "" {
			names = append(names, field.Name)
			inserts = append(inserts, strings.ReplaceAll(field.Insert, Value, index))
		}
		if field.Update != "" {
			updates = append(updates, field.Name+" = "+strings.ReplaceAll(field.Update, Value, index))
		}
		if field.ID {
			ids = append(ids, field.Name)
		}
	}

	b.WriteString("WITH ")
	b.WriteString(t.Table)
	b.WriteString(" AS ( INSERT INTO ")
	if t.Schema != "" {
		b.WriteString(t.Schema)
		b.WriteString(".")
	}
	b.WriteString(t.Table)
	b.WriteString(" (")
	b.WriteString(strings.Join(names, ",")) // Fields
	b.WriteString(") VALUES(")
	b.WriteString(strings.Join(inserts, ",")) // Inserts
	b.WriteString(") ON CONFLICT (")          // ID Fields
	b.WriteString(strings.Join(ids, ","))     // Inserts
	b.WriteString(") DO UPDATE SET ")
	b.WriteString(strings.Join(updates, ",")) // Updates
	b.WriteString(" RETURNING *")
	b.WriteString(") SELECT ")
	b.WriteString(t.SelectFields)
	if t.SelectAdditionalFields != "" {
		if t.SelectFields != "" {
			b.WriteString(",")
		}
		b.WriteString(t.SelectAdditionalFields)
	}
	b.WriteString(" FROM ")
	b.WriteString(t.Table)
	if t.Joins != "" {
		b.WriteString(" ")
		b.WriteString(t.Joins)
	}
	return b.String()

}

func (t *Table[T]) GenerateSelectorQuery() string {
	var b strings.Builder
	b.WriteString("SELECT ")
	b.WriteString(t.SelectFields)
	if t.SelectAdditionalFields != "" {
		if t.SelectFields != "" {
			b.WriteString(",")
		}
		b.WriteString(t.SelectAdditionalFields)
	}
	b.WriteString(` FROM `)
	b.WriteString(t.Table)
	if t.Joins != "" {
		b.WriteString(" ")
		b.WriteString(t.Joins)
	}
	return b.String()
}
