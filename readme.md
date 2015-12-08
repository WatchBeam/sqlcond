# sqlcond [![Build Status](https://travis-ci.org/WatchBeam/sqlcond.svg?branch=master)](https://travis-ci.org/WatchBeam/sqlcond) [![godoc reference](https://godoc.org/github.com/WatchBeam/sqlcond?status.png)](https://godoc.org/github.com/WatchBeam/sqlcond)

Little utility to wait for the database to match a state.

```go

cond := sqlcond.New(db, sqlcond.Exists("tt", "id = ?", 1))
defer cond.Close()

select {
case <-cond.C:
    fmt.Println("Row with ID 1 appeared! Hurrah!")
case err := <-cond.Errors:
    fmt.Println("An error occurred: " + err.Error())
case <-time.After(time.Second * 5):
    fmt.Println("Timeout")
}

```
