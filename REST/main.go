package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"goMS1Assignment/REST/database"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/microcosm-cc/bluemonday"
	log "github.com/sirupsen/logrus"
)

var (
	courses      map[int]database.CourseInfo //used for storing courses on the REST API
	db           *sql.DB
	API_key      string
	codeRegExp   *regexp.Regexp
	detailRegExp *regexp.Regexp
	pol          = bluemonday.UGCPolicy() //pol for policy. Used for sanitization of input after input validation.
)

func init() {
	//below codes are for initializing third party logrus
	var filename string = "log/logfile.log"
	// Create the log file if doesn't exist. And append to it if it already exists.
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.SetFormatter(Formatter)
	log.SetLevel(log.WarnLevel) //only log the warning severity level or higher

	if err != nil {
		// Cannot open log file. Logging to stderr
		fmt.Println(err)
	} else {
		log.SetOutput(io.MultiWriter(file, os.Stdout)) //default logger will be writing to file and os.Stdout
	}
	codeRegExp = regexp.MustCompile(`^[0-9]*$`)                                              //code regexp to check for code pattern match
	detailRegExp = regexp.MustCompile(`^[\w'\-,.][^_!¡?÷?¿/\\+=$%ˆ&*(){}|~<>;:[\]]{0,250}$`) //regexp to check for Title, Dates, Lecturer and Description
}

func goDotEnvVariable(envVariable string) string {
	//load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")

	}
	return os.Getenv(envVariable)
}

func validKey(w http.ResponseWriter, r *http.Request) bool {
	v := r.URL.Query()
	if key, ok := v["key"]; ok {
		if key[0] == API_key {
			return true
		} else { //invalid key
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("401 - Invalid key"))
			return false
		}
	} else { //key is not provided
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("401 - Please supply access key"))
		return false
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the REST API!")
}

//func allcourses retrieves all courses from database and JSON encodes courses for http response writer
func allcourses(w http.ResponseWriter, r *http.Request) {
	if !validKey(w, r) {
		return
	}
	fmt.Fprintf(w, "List of all courses")

	courses = database.GetRecords(db)
	fmt.Println(courses)
	for _, v := range courses {
		validateAndSanitize(&v)
	}
	json.NewEncoder(w).Encode(courses)

}

//course handles the incoming console http request (Get, Post, Put, Delete) and handles the requests accordingly
func course(w http.ResponseWriter, r *http.Request) {
	if !validKey(w, r) {
		return
	}

	params := mux.Vars(r)
	//fmt.Println(params)

	code, err := strconv.Atoi(params["courseid"]) //code needs to be converted to int as it is set up as int in CourseInfo struct
	if err != nil {
		fmt.Println("Course ID is not an integer.")
		log.Error("Error at course function, received non-integer course ID")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 - Course data in wrong format, needs to be integer value."))
		return
	}

	course := database.GetRecord(db, code)
	validateAndSanitize(&course)

	if r.Method == "GET" {
		exist := database.RowExists(db, code)

		if exist {
			json.NewEncoder(w).Encode(course)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - No course found"))
		}
	}

	if r.Method == "DELETE" {

		exist := database.RowExists(db, code)

		if exist {
			database.DeleteRecord(db, code)
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("202 - Course deleted: " + params["courseid"]))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - No course found"))
		}
	}

	//only methods PUT and POST are with content-type "application/json"
	if r.Header.Get("Content-type") == "application/json" {
		//POST is for creating new course
		if r.Method == "POST" {
			// read the string sent to the service
			var newCourse database.CourseInfo
			reqBody, err := ioutil.ReadAll(r.Body)

			if err == nil {
				//convert JSON to object
				json.Unmarshal(reqBody, &newCourse)
				validateAndSanitize(&newCourse)

				// check if course exists; add only if course does not exist
				if _, ok := courses[code]; !ok {
					if newCourse.Title == "" || newCourse.Dates == "" || newCourse.Lecturer == "" || newCourse.Description == "" {
						w.WriteHeader(http.StatusUnprocessableEntity)
						w.Write([]byte("422 - Please supply course" + "information " + "in JSON format"))
						return
					}
					database.InsertRecord(db, newCourse.Code, newCourse.Title, newCourse.Dates, newCourse.Lecturer, newCourse.Description)
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("201 - Course added: " + params["courseid"]))
				} else {
					w.WriteHeader(http.StatusConflict)
					w.Write([]byte("409 - Duplicate course ID"))
					log.Error("Error at course function, 409 - Duplicate course ID")
				}
			} else {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte("422 - Please supply course information " + "in JSON format"))
				log.Error("Error at course function, 422 - Invalid course information.")
			}
		}

		//---PUT is for creating or updating existing course ---
		if r.Method == "PUT" {
			var newCourse database.CourseInfo
			reqBody, err := ioutil.ReadAll(r.Body)

			if err == nil {
				//convert JSON to object
				json.Unmarshal(reqBody, &newCourse)
				validateAndSanitize(&newCourse)

				// check if course exists; add only if course does not exist
				if exist := database.RowExists(db, newCourse.Code); !exist {
					if newCourse.Title == "" || newCourse.Dates == "" || newCourse.Lecturer == "" || newCourse.Description == "" {
						w.WriteHeader(http.StatusUnprocessableEntity)
						w.Write([]byte("422 - Please supply course" + "information " + "in JSON format"))
						log.Error("Error at course function, 422 - Invalid course information.")
						return
					}
					database.InsertRecord(db, newCourse.Code, newCourse.Title, newCourse.Dates, newCourse.Lecturer, newCourse.Description)
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("201 - Course added: " + params["courseid"]))
				} else {
					// update course
					if newCourse.Title == "" {
						newCourse.Title = course.Title
					}
					if newCourse.Dates == "" {
						newCourse.Dates = course.Dates
					}
					if newCourse.Lecturer == "" {
						newCourse.Lecturer = course.Lecturer
					}
					if newCourse.Description == "" {
						newCourse.Description = course.Description
					}
					database.EditRecord(db, newCourse.Code, newCourse.Title, newCourse.Dates, newCourse.Lecturer, newCourse.Description)
					w.WriteHeader(http.StatusAccepted)
					w.Write([]byte("202 - Course updated: " + params["courseid"]))
				}
			} else {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte("422 - Please supply course information " + "in JSON format"))
				log.Error("Error at course function, 422 - Invalid course information.")
			}
		}
	}
}

func main() {

	//the following variables are hidden using an environment variable so as not to expose security related data.
	API_key = goDotEnvVariable("API_KEY")
	db_password := goDotEnvVariable("PASSWORD")
	db_port := goDotEnvVariable("PORT")
	db_name := goDotEnvVariable("DB_NAME")

	var dataSourceName string = "root:" + db_password + "@tcp" + db_port + "/" + db_name
	var err error
	db, err = sql.Open("mysql", dataSourceName)

	if err != nil {
		log.Panic("Panic occured opening data base", err.Error())
	} else {
		fmt.Println("Database opened")
	}
	defer func() {
		db.Close()
		fmt.Println("Database closed")
	}()

	//instatiate courses
	courses = make(map[int]database.CourseInfo)

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/", home)
	router.HandleFunc("/api/v1/courses", allcourses)
	router.HandleFunc("/api/v1/courses/{courseid}", course).Methods("GET", "PUT", "POST", "DELETE")
	//register the methods with handler functions

	fmt.Println("Listening at port 5000")
	//log.Fatal(http.ListenAndServe(":5000", router))
	log.Fatal(http.ListenAndServeTLS(":5000", "./server.crt", "./server.key", router))
}

func validateAndSanitize(course *database.CourseInfo) error {
	var err error = errors.New("Error at Validate and Sanitization.")

	var codeString string
	codeString = strconv.Itoa(course.Code)

	if !codeRegExp.MatchString(codeString) {
		log.Error("Incorrect input format for course code detected during validation and sanitization.")
		return err
	}
	codeString = pol.Sanitize(codeString)
	course.Code, _ = strconv.Atoi(codeString)

	if course.Title != "" && !detailRegExp.MatchString(course.Title) {
		log.Error("Incorrect input format for Title input detected during validation and sanitization.")
		return err
	}
	course.Title = pol.Sanitize(course.Title) //pol.Sanitize is used to sanitize inputs after validation using BlueMonday

	if course.Dates != "" && !detailRegExp.MatchString(course.Dates) {
		log.Error("Incorrect input format for Dates input detected during validation and sanitization.")
		return err
	}
	course.Dates = pol.Sanitize(course.Dates) //pol.Sanitize is used to sanitize inputs after validation using BlueMonday

	if course.Lecturer != "" && !detailRegExp.MatchString(course.Lecturer) {
		log.Error("Incorrect input format for Lecturer input detected during validation and sanitization.")
		return err
	}
	course.Lecturer = pol.Sanitize(course.Lecturer) //pol.Sanitize is used to sanitize inputs after validation using BlueMonday

	if course.Description != "" && !detailRegExp.MatchString(course.Description) {
		log.Error("Incorrect input format for Dates input detected during validation and sanitization.")
		return err
	}
	course.Description = pol.Sanitize(course.Description) //pol.Sanitize is used to sanitize inputs after validation using BlueMonday

	return nil
}
