package sqlcond

import (
	"database/sql"
	"testing"
	"time"
)

func mustExec(db *sql.DB, query string) {
	if _, err := db.Exec(query); err != nil {
		panic(err)
	}
}

func getDatabase() *sql.DB {
	db, err := sql.Open("mysql", "root@/sqlcond")
	if err != nil {
		panic(err)
	}

	mustExec(db, "DROP TABLE IF EXISTS tt;")
	mustExec(db, "CREATE TABLE `tt` (`id` INT NOT NULL);")

	return db
}

func TestExistsWorks(t *testing.T) {
	db := getDatabase()
	defer db.Close()
	mustExec(db, "INSERT INTO `tt` VALUES (1);")

	cond := New(db, Exists("tt", "id = ?", 1))
	defer cond.Close()

	select {
	case <-cond.C:
		break
	case err := <-cond.Errors:
		t.Fatal(err)
	case <-time.After(time.Second * 5):
		t.Fatal("timed out")
	}
}

func TestAbortsEarly(t *testing.T) {
	db := getDatabase()
	defer db.Close()

	cond := New(db, Exists("tt", "id = ?", 1))
	go func() {
		time.Sleep(time.Second * 2)
		cond.Close()
	}()

	select {
	case <-cond.C:
		break
	case err := <-cond.Errors:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second * 5):
		t.Fatal("timed out")
	}
}
