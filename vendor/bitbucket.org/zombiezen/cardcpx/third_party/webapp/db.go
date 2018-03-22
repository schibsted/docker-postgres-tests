package webapp

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
)

// ScanOneStruct is equivalent to calling rows.Next, ScanStruct, then rows.Close.
// If there is no next row, sql.ErrNoRows is returned.
func ScanOneStruct(rows *sql.Rows, val interface{}) error {
	defer rows.Close()
	if !rows.Next() {
		return sql.ErrNoRows
	}
	return ScanStruct(rows, val)
}

// ScanStruct extracts a single row into the struct pointed to by val.
//
// Each exported struct field is converted to a column name by using the "sql"
// tag of that field, or by transforming the field name from upper camel case to
// lowercase underscore-separated words if there is no tag.  If the field tag is
// "-", then the field will be ignored.
//
// It is not an error to have extra columns or fields.
func ScanStruct(rows *sql.Rows, val interface{}) error {
	v := reflect.ValueOf(val)
	if v.Kind() != reflect.Ptr {
		return errors.New("db: val is not a pointer")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return errors.New("db: val is not a pointer to a struct")
	}
	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	dest := make([]interface{}, len(cols))
	var placeholder interface{}
	for i := range dest {
		dest[i] = &placeholder
	}

	for fi := 0; fi < v.NumField(); fi++ {
		name := fieldColumn(v.Type().Field(fi))
		if name == "" {
			continue
		}
		for i := range cols {
			if cols[i] == name {
				dest[i] = v.Field(fi).Addr().Interface()
				break
			}
		}
	}
	return rows.Scan(dest...)
}

// TransactionError is returned by RunInTransaction.
type TransactionError struct {
	Err   error // Error surrounding transaction
	TxErr error // Error during transaction
}

func (e *TransactionError) Error() string {
	if e.Err != nil {
		return "during transaction: " + e.Err.Error()
	}
	return "transaction: " + e.TxErr.Error()
}

// RunInTransaction executes a SQL operation in a transaction.  Any non-nil
// error will be wrapped in a TransactionError.
func RunInTransaction(db *sql.DB, f func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return &TransactionError{nil, err}
	}
	err = f(tx)
	var txErr error
	if err == nil {
		txErr = tx.Commit()
	} else {
		txErr = tx.Rollback()
	}
	if err != nil || txErr != nil {
		return &TransactionError{err, txErr}
	}
	return nil
}

// fieldColumn returns the SQL column name for a struct field.  An empty string
// means this field should not be persisted.
func fieldColumn(f reflect.StructField) string {
	if tag := f.Tag.Get("sql"); tag == "-" {
		return ""
	} else if tag != "" {
		return tag
	}
	return colname(f.Name)
}

// colname converts from upper camel case to underscore-separated words.
func colname(s string) string {
	words := make([]string, 0)
	var start int
	lastIdx, lastRune := -1, rune(0)
	for i, r := range s {
		if i == 0 {
			// don't create a new word, just grab the rune
		} else if isUpper(lastRune) && isLower(r) && start != lastIdx {
			words = append(words, strings.ToLower(s[start:lastIdx]))
			start = lastIdx
		} else if isLower(lastRune) && isUpper(r) && start != i {
			words = append(words, strings.ToLower(s[start:i]))
			start = i
		}
		lastIdx, lastRune = i, r
	}
	words = append(words, strings.ToLower(s[start:]))
	return strings.Join(words, "_")
}

func lower(r rune) rune {
	return (r - 'A') + 'a'
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}
