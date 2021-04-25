package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/microcosm-cc/bluemonday"
	log "github.com/sirupsen/logrus"
)

const baseURL = "https://localhost:5000/api/v1/courses"

type CourseInfo struct {
	Code        int    `json:"Code"`
	Title       string `json:"Title"`
	Dates       string `json:"Dates"`
	Lecturer    string `json:"Lecturer"`
	Description string `json:"Description"`
}

var (
	key          string
	codeRegExp   *regexp.Regexp
	detailRegExp *regexp.Regexp
	pol          = bluemonday.UGCPolicy() //pol for policy. Used for sanitization of input after input validation.
	client       = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: loadCA("cert/ca.crt")},
		},
	}
)

func init() {
	//godotenv package
	key = goDotEnvVariable("API_KEY")

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

// LoadCAFile loads a single PEM-encoded file from the path specified.
func loadCA(caFile string) *x509.CertPool {
	pool := x509.NewCertPool()

	if ca, err := ioutil.ReadFile(caFile); err != nil {
		log.Fatal("Fatal Error at Certification ReadFile: ", err)
	} else {
		pool.AppendCertsFromPEM(ca)
	}
	return pool
}

//use godot package to load/read the .env file and return the value of the key
func goDotEnvVariable(key string) string {
	//load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")

	}
	return os.Getenv(key)
}

//addCourse sends a http request with Method Get and awaits a response
func getCourse(code string) {
	url := baseURL + "?key=" + key
	if code != "" {
		url = baseURL + "/" + code + "?key=" + key
	}
	response, err := client.Get(url)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		log.Error("Error at get course function", err.Error())
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(response.StatusCode)
		fmt.Println(string(data))
		response.Body.Close()
	}
}

//addCourse sends a http request with Method post and awaits a response
func addCourse(code string, jsonData CourseInfo) {
	jsonValue, _ := json.Marshal(jsonData)

	response, err := client.Post(baseURL+"/"+code+"?key="+key, "application/json", bytes.NewBuffer(jsonValue))

	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		log.Error("Error at add course function", err.Error())
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(response.StatusCode)
		fmt.Println(string(data))
		response.Body.Close()
	}
}

//updateCourse sends a http request with Method put and awaits a response
func updateCourse(code string, jsonData CourseInfo) {
	jsonValue, _ := json.Marshal(jsonData)

	request, err := http.NewRequest(http.MethodPut, baseURL+"/"+code+"?key="+key, bytes.NewBuffer(jsonValue))

	request.Header.Set("Content-Type", "application/json")

	//client := &http.Client{}
	response, err := client.Do(request) //this is to send the request
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		log.Error("Error at update course function", err.Error())
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(response.StatusCode)
		fmt.Println(string(data))
		response.Body.Close()
	}
}

//deleteCourse sends a http request with Method delete and awaits a response
func deleteCourse(code string) {
	request, err := http.NewRequest(http.MethodDelete, baseURL+"/"+code+"?key="+key, nil)

	//client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		log.Error("Error at delete course function", err.Error())
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(response.StatusCode)
		fmt.Println(string(data))
		response.Body.Close()
	}
}

func menu() {
	var choice string

	for choice != "e" {

		fmt.Println("===============================")
		fmt.Println("Welcome to API client console")
		fmt.Println("===============================")
		fmt.Println("Please select 1 of the following URL queries:")
		fmt.Println(" c - Create")
		fmt.Println(" r - Read")
		fmt.Println(" u - Update")
		fmt.Println(" d - Delete")
		fmt.Println(" e - Exit Console")
		fmt.Println(" Please make a selection and hit enter.")

		fmt.Scanln(&choice)

		switch choice {
		case "c":
			createFunc()
		case "r":
			readFunc()
		case "u":
			updateFunc()
		case "d":
			deleteFunc()
		case "e":
			fmt.Println("Exiting the program..")
		default:
			fmt.Println("You did not make a valid section. Please try again.")
		}
	}
}

//console function to create new course
func createFunc() {
	var code int
	fmt.Println("Please enter the following details:")
	fmt.Println("Please enter course code")
	fmt.Scanln(&code)

	var codeString string
	codeString = strconv.Itoa(code)

	if !codeRegExp.MatchString(codeString) {
		log.Error("Incorrect input format for course code detected at create function.")
		return
	}
	codeString = pol.Sanitize(codeString)

	fmt.Println("Please enter course title")
	reader := bufio.NewReader(os.Stdin) //create new reader, assuming bufio imported
	title, _ := reader.ReadString('\n')
	title = strings.TrimRight(title, "\n") // this removes the \n at end of scan function
	if !detailRegExp.MatchString(title) {
		log.Error("Incorrect input format for title input detected at create function.")
		return
	}
	title = pol.Sanitize(title) //pol.Sanitize is used to sanitize inputs after validation using BlueMonday

	fmt.Println("Please enter course dates")
	reader = bufio.NewReader(os.Stdin) //create new reader, assuming bufio imported
	dates, _ := reader.ReadString('\n')
	dates = strings.TrimRight(dates, "\n") // this removes the \n at end of scan function
	if !detailRegExp.MatchString(dates) {
		log.Error("Incorrect input format for dates input detected at create function.")
		return
	}
	dates = pol.Sanitize(dates)

	fmt.Println("Please enter lecturer name")
	reader = bufio.NewReader(os.Stdin) //create new reader, assuming bufio imported
	lecturer, _ := reader.ReadString('\n')
	lecturer = strings.TrimRight(lecturer, "\n") // this removes the \n at end of scan function
	if !detailRegExp.MatchString(lecturer) {
		log.Error("Incorrect input format for title input detected at create function.")
		return
	}
	lecturer = pol.Sanitize(lecturer)

	fmt.Println("Please enter course description")
	reader = bufio.NewReader(os.Stdin) //create new reader, assuming bufio imported
	description, _ := reader.ReadString('\n')
	description = strings.TrimRight(description, "\n") // this removes the \n at end of scan function
	if !detailRegExp.MatchString(description) {
		log.Error("Incorrect input format for description input detected at create function.")
		return
	}
	description = pol.Sanitize(description)

	newCourse := CourseInfo{code, title, dates, lecturer, description}
	addCourse(codeString, newCourse)
}

//console function to read input course code
func readFunc() {
	fmt.Println("Please select from the following")
	fmt.Println("1. Get all courses.")
	fmt.Println("2. Get specific course.")
	var choice int
	var code string
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		code = ""
		getCourse(code)
	case 2:
		fmt.Println("Please enter course code:")
		fmt.Scanln(&code)
		if !codeRegExp.MatchString(code) {
			log.Error("Incorrect input format for course code detected at read function.")
			return
		}
		code = pol.Sanitize(code)
		getCourse(code)
	default:
		fmt.Println("You did not make a valid selection, please try again. Returning to main menu")
		fmt.Println("==========================================")
		menu()
	}
}

//console function to update input course code
func updateFunc() {
	var code int
	fmt.Println("Please enter the following details:")
	fmt.Println("Please enter course code")
	fmt.Scanln(&code)

	var codeString string
	codeString = strconv.Itoa(code)

	if !codeRegExp.MatchString(codeString) {
		log.Error("Incorrect input format for course code detected at update function.")
		return
	}
	codeString = pol.Sanitize(codeString)

	fmt.Println("Please enter course title. Press enter if no change.")
	reader := bufio.NewReader(os.Stdin) //create new reader, assuming bufio imported
	title, _ := reader.ReadString('\n')
	title = strings.TrimRight(title, "\n") // this removes the \n at end of scan function
	if title != "" && !detailRegExp.MatchString(title) {
		log.Error("Incorrect input format for title input detected at update function.")
		return
	}
	title = pol.Sanitize(title)

	fmt.Println("Please enter course dates. Press enter if no change.")
	reader = bufio.NewReader(os.Stdin) //create new reader, assuming bufio imported
	dates, _ := reader.ReadString('\n')
	dates = strings.TrimRight(dates, "\n") // this removes the \n at end of scan function
	if dates != "" && !detailRegExp.MatchString(dates) {
		log.Error("Incorrect input format for dates input detected at update function.")
		return
	}
	dates = pol.Sanitize(dates)

	fmt.Println("Please enter lecturer name. Press enter if no change.")
	reader = bufio.NewReader(os.Stdin) //create new reader, assuming bufio imported
	lecturer, _ := reader.ReadString('\n')
	lecturer = strings.TrimRight(lecturer, "\n") // this removes the \n at end of scan function
	if lecturer != "" && !detailRegExp.MatchString(lecturer) {
		log.Error("Incorrect input format for lecturer input detected at create function.")
		return
	}
	lecturer = pol.Sanitize(lecturer)

	fmt.Println("Please enter course description. Press enter if no change.")
	reader = bufio.NewReader(os.Stdin) //create new reader, assuming bufio imported
	description, _ := reader.ReadString('\n')
	description = strings.TrimRight(description, "\n") // this removes the \n at end of scan function
	if description != "" && !detailRegExp.MatchString(description) {
		log.Error("Incorrect input format for description input detected at create function.")
		return
	}
	description = pol.Sanitize(description)

	updatedCourse := CourseInfo{code, title, dates, lecturer, description}
	updateCourse(strconv.Itoa(code), updatedCourse)
}

//console function to delete input course code
func deleteFunc() {
	fmt.Println("Please enter course code to delete.")
	var code string
	fmt.Scanln(&code)

	if !codeRegExp.MatchString(code) {
		log.Error("Incorrect input format for course code detected at delete function.")
		return
	}
	code = pol.Sanitize(code)
	deleteCourse(code)
}

func main() {

	menu()

}
