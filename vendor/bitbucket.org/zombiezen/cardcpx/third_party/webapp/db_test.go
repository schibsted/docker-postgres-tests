package webapp

import (
	"database/sql"
	"log"
	"testing"
)

func TestColname(t *testing.T) {
	tests := []struct {
		s   string
		out string
	}{
		{"", ""},
		{"a", "a"},
		{"A", "a"},
		{"ID", "id"},
		{"Foo", "foo"},
		{"FooB", "foo_b"},
		{"FirstName", "first_name"},
		{"StudentID", "student_id"},
		{"studentID", "student_id"},
		{"FooIDBar", "foo_id_bar"},
		{"FooID4Bar", "foo_id4_bar"},
	}
	for _, test := range tests {
		result := colname(test.s)
		if result != test.out {
			t.Errorf("colname(%q) = %v; want %v", test.s, result, test.out)
		}
	}
}

func ExampleScanStruct() {
	type Person struct {
		ID        int    `sql:"ID"`
		FirstName string // implicitly "first_name"
		LastName  string // implicitly "last_name"
		Ignored   bool   `sql:"-"`
	}

	var person Person
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE person ( ID integer, first_name text, last_name text );")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO person ( ID, first_name, last_name ) VALUES (1, 'John', 'Doe');")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query("SELECT * FROM person;")
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	if err := ScanStruct(rows, &person); err != nil {
		log.Fatal(err)
	}
}
