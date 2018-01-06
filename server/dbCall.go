package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := connectDB(
		`db_name`,
		"user",
		"password",
		"127.0.0.1",
		"3306")
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()

	for i := 1; i <= 65536; i++ {
		name := getASName(i, db)
		fmt.Printf("AS%v = %v\n", i, name)
	}
}

func connectDB(dbName, user, password, host, port string) (*sql.DB, error) {
	conn := fmt.Sprintf(`%v:%v@tcp(%v:%v)/%v`, user, password, host, port, dbName)
	db, err := sql.Open("mysql", conn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func getASName(asNumber int, db *sql.DB) string {
	var name string
	err := db.QueryRow("select AS_NAME from ASN where AS_NUM = ?", asNumber).Scan(&name)
	if err != nil {
		return ""
	}
	return name

}
