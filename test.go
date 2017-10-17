package main

//import (
//	"database/sql"
//	_ "github.com/go-sql-driver/mysql"
//	"fmt"
//)
//
//func main()  {
//	db, err := sql.Open("mysql", "madhav:password@/Auditbot")
//	fmt.Println(db)
//	if err != nil {
//		fmt.Errorf(err.Error())
//	}
//	rows, err := db.Query(fmt.Sprintf("SELECT answer FROM qrn555"))
//	if err != nil {
//		panic(err)
//	}
//	var allAnswers string = ""
//	for rows.Next() {
//		var answer string
//		err = rows.Scan(&answer)
//		if err != nil {
//			panic(err)
//		}
//		if len(allAnswers) > 0 {
//			allAnswers = fmt.Sprintf("%s,%s", allAnswers, answer)
//		} else {
//			allAnswers = fmt.Sprintf("%s", answer)
//		}
//		//fmt.Println(answer)
//	}
//	fmt.Println(allAnswers)
//}
