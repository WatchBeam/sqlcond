package sqlcond_test

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/WatchBeam/sqlcond"
	_ "github.com/go-sql-driver/mysql"
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

func ExampleSQLCond() {
	db := getDatabase()
	defer db.Close()

	go func() {
		time.Sleep(time.Second)
		mustExec(db, "INSERT INTO `tt` VALUES (1);")
	}()

	cond := sqlcond.New(db, sqlcond.Exists("tt", "id = ?", 1))
	defer cond.Close()

	select {
	case err := <-cond.Errors:
		fmt.Println("An error occurred: " + err.Error())
	case <-cond.C:
		fmt.Println("Row with ID 1 appeared")
	case <-time.After(time.Second * 5):
		fmt.Println("Timeout")
	}
	// Output: Row with ID 1 appeared
}
