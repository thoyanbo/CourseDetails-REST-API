package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

type CourseInfo struct {
	Code        int    `json:"Code"`
	Title       string `json:"Title"`
	Dates       string `json:"Dates"`
	Lecturer    string `json:"Lecturer"`
	Description string `json:"Description"`
}

//InsertRecord queries the database to delete existing course
func DeleteRecord(db *sql.DB, Code int) {
	ctx := context.Background()
	query := "DELETE FROM CourseInfo WHERE Code = ?"
	_, err := db.QueryContext(ctx, query, Code)
	if err != nil {
		log.Panic("Error deleting record.", err.Error())
	}

	// query := fmt.Sprintf("DELETE FROM CourseInfo WHERE Code='%d'", Code)
	// _, err := db.Query(query)
	// if err != nil {
	// 	panic(err.Error())
	// }
}

//InsertRecord queries the database to update existing course
func EditRecord(db *sql.DB, Code int, Title string, Dates string, Lecturer string, Description string) {

	ctx := context.Background()
	query := "UPDATE CourseInfo SET Title=?, Dates=?, Lecturer=?, Description=? WHERE Code=?"
	_, err := db.QueryContext(ctx, query, Title, Dates, Lecturer, Description, Code)
	if err != nil {
		log.Panic("Panic at Update Record.", err.Error())
	}

}

//InsertRecord queries the database to create new course
func InsertRecord(db *sql.DB, Code int, Title string, Dates string, Lecturer string, Description string) {

	ctx := context.Background()
	query := "INSERT INTO CourseInfo (Code, Title, Dates, Lecturer, Description) VALUES (?, ?, ?, ?, ?)"
	_, err := db.QueryContext(ctx, query, Code, Title, Dates, Lecturer, Description)
	if err != nil {
		log.Panic("Panic at Insert Record.", err.Error())
	}

}

//GetRecord queries the database to return all courses
func GetRecords(db *sql.DB) map[int]CourseInfo {
	courses := make(map[int]CourseInfo)

	results, err := db.Query("Select * FROM goMS1_db.CourseInfo")
	if err != nil {
		log.Panic("Panic at Get Records.", err.Error())
	}

	for results.Next() {
		// map this type to the record in the table
		var course CourseInfo
		err = results.Scan(&course.Code, &course.Title, &course.Dates, &course.Lecturer, &course.Description)
		if err != nil {
			log.Panic("Panic at Get Records.", err.Error())
		}

		//fmt.Println(course.Code, course.Title, course.Dates, course.Lecturer, course.Description)
		courses[course.Code] = CourseInfo{course.Code, course.Title, course.Dates, course.Lecturer, course.Description}
	}
	//fmt.Println(courses)
	return courses
}

//GetRecord queries the SQL database and returns a course
func GetRecord(db *sql.DB, Code int) CourseInfo {

	ctx := context.Background()
	var course CourseInfo
	query := "SELECT * FROM goMS1_db.CourseInfo WHERE Code =?"
	results, err := db.QueryContext(ctx, query, Code)
	if err != nil {
		log.Panic("Panic at Get Record.", err.Error())
	}

	for results.Next() {
		// map this type to the record in the table
		err = results.Scan(&course.Code, &course.Title, &course.Dates, &course.Lecturer, &course.Description)
		if err != nil {
			log.Panic("Panic at Get Record.", err.Error())
		}
	}

	return course

}

//RowExists queries table CourseInfo with code and returns a bool if code exists
func RowExists(db *sql.DB, code int) bool {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM CourseInfo WHERE code = '%d')", code)
	err := db.QueryRow(query).Scan(&exists)

	if err != nil {
		log.Error("Error at Row Exists", err.Error())
	}

	if exists == false {
		log.Warning("Code ", code, " does not exist. Warning triggered at function RowExists.")
	}
	return exists

}
