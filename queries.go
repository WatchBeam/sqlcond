package sqlcond

import (
	"database/sql"
)

// Queries are passed into the SQLCond and used to notify it when the
// conditions are satisfied, and provide it with results.
type Query interface {
	// Can be used to prepare a statement for multiple executions.
	Prepare(db *sql.DB) error
	// Returns whether the condition is satisfied. If so, there should be
	// a non-nil "results" returned.
	Attempt(db *sql.DB) (satisfied bool, results *sql.Row, err error)
	// Called when the SQLCond has no more need of the query.
	Close()
}

type baseQuery struct {
	query    string
	params   []interface{}
	tester   func(*sql.Row) (bool, error)
	prepared *sql.Stmt
}

func (b *baseQuery) Prepare(db *sql.DB) error {
	var err error
	b.prepared, err = db.Prepare(b.query)
	return err
}

func (b *baseQuery) Attempt(db *sql.DB) (satisfied bool, results *sql.Row, err error) {
	results = b.prepared.QueryRow(b.params...)
	satisfied, err = b.tester(results)
	return
}

func (b *baseQuery) Close() {
	b.prepared.Close()
}

func testExistence(table, cols, where string, args []interface{}) Query {
	return &baseQuery{
		query:  "SELECT " + cols + " FROM " + table + " WHERE " + where + " LIMIT 1",
		params: args,
		tester: func(r *sql.Row) (bool, error) {
			var target interface{}
			err := r.Scan(&target)
			if err == nil {
				return true, nil
			} else if err == sql.ErrNoRows {
				return false, nil
			} else {
				return false, err
			}
		},
	}
}

// This checks for the existence of at least one record in the table matching
// the "where" query. No useful data will be returned in the result rows;
// if you need the record, use Once, or your own query.
func Exists(table, where string, args ...interface{}) Query {
	return testExistence(table, "1", where, args)
}

// This is almost the same as Exists, except that it includes columns
// of the provided table in the results.
func Once(table string, columns []string, where string, args ...interface{}) Query {
	cols := ""
	for _, column := range columns {
		cols += ",`" + column + "`"
	}

	return testExistence(table, cols[1:], where, args)
}
