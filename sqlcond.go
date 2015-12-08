// SQLCond provides functionality to wait for a condition to be met in a SQL
// database. It's a small structure that wraps a query until it yields
// some result.
package sqlcond

import (
	"database/sql"
	"time"
)

const (
	// Duration between query times.
	defaultPoll = time.Millisecond * 500
)

type SQLCond struct {
	// Onces rows that satisfy out requirements are found, they'll be
	// sent down this buffered channel. Once a result is sent, the SQLCond
	// will close.
	C chan *sql.Row
	// Buffered channel of errors that can be a result of query operations.
	// Note that when an error is thrown, the cond will NOT close; you are
	// responsible for doing that.
	Errors chan error

	closer chan bool
	db     *sql.DB
	poll   time.Duration
	query  Query
}

func New(db *sql.DB, query Query) *SQLCond {
	cond := &SQLCond{
		C:      make(chan *sql.Row, 1),
		Errors: make(chan error, 4),
		closer: make(chan bool, 1),
		db:     db,
		poll:   defaultPoll,
		query:  query,
	}
	go cond.run()

	return cond
}

// Halts the SQLCond. The SQLCond only closes by itself when it gets a result
// (which is sent down SQLCond.C). In all other cases, you must call this
// function manually.
func (s *SQLCond) Close() {
	s.closer <- true
}

func (s *SQLCond) run() {
	defer func() {
		close(s.C)
		close(s.Errors)
	}()

	if err := s.query.Prepare(s.db); err != nil {
		s.Errors <- err
		return
	}

	t := time.NewTicker(s.poll)
	defer func() {
		t.Stop()
		s.query.Close()
	}()

	for {
		select {
		case <-s.closer:
			return
		case <-t.C:
			satisfied, row, err := s.query.Attempt(s.db)
			if err != nil {
				s.Errors <- err
			} else if satisfied {
				s.C <- row
			}
		}
	}
}
