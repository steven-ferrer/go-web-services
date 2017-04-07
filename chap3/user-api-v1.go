package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
)

var database *sql.DB

//User user
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	First    string `json:"first"`
	Last     string `json:"last"`
}

//Users list of users
type Users struct {
	Users []User `json:"users"`
}

//CreateResponse response if any error occurs
type CreateResponse struct {
	Error     string `json:"error"`
	ErrorCode string `json:"code"`
}

//ErrMsg error object
type ErrMsg struct {
	ErrCode    int
	StatusCode int
	Msg        string
}

func main() {

	//initialize the database
	db, err := sql.Open("mysql", "root:password@/social_network")
	if err != nil {
		fmt.Println(err)
	}
	//initialize the global database
	database = db

	routes := httprouter.New()
	routes.GET("/api", OptionsGet)
	routes.GET("/api/users", UsersGet)     //get all users
	routes.POST("/api/users", UsersCreate) //create user

	routes.GET("/api/users/:id", UserGet)     //get a specific user
	routes.POST("/api/users/:id", UserUpdate) //update a specific user

	routes.DELETE("/api/users/:id", UserDelete) //delete a user
	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", routes))
}

//OptionsGet default endpoint, displays the available endpoints
func OptionsGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintln(w, "Available Endpoing/s: \n * /api/users")
}

//UsersGet endpoint for getting all users
func UsersGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Println("starting retrieval")
	start := 0
	limit := 10

	next := start + limit

	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Link", "<http://localhost:8080/api/users?start="+string(next)+"; rel=\"next\"")

	rows, _ := database.Query("SELECT * FROM users LIMIT 10")

	users := Users{}

	for rows.Next() {
		user := User{}
		rows.Scan(&user.ID, &user.Username, &user.First, &user.Last, &user.Email)
		users.Users = append(users.Users, user)
	}

	output, err := json.Marshal(users)
	if err != nil {
		fmt.Fprintln(w, "Something went wrong while processing your request: ", err)
	}

	fmt.Fprintln(w, string(output))
}

//UsersCreate endpoint for creating a user
func UsersCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	//freakin' Access-Control-Allow-Origin
	//Still looking for a fix for this
	//w.Header().Set("Access-Control-Allow-Origin", "http://localhost:9000")
	//r.Header.Set("Origin", "http://localhost:9000")
	fmt.Println(r.Header.Get("Origin"))
	NewUser := User{}
	fmt.Println(r.FormFile("userImage"))
	f, _, err := r.FormFile("userImage")
	if err != nil {
		fmt.Println(err.Error())
	}

	NewUser.Username = r.FormValue("username")
	NewUser.Email = r.FormValue("email")
	NewUser.First = r.FormValue("first")
	NewUser.Last = r.FormValue("last")

	//returns a byte
	fileData, _ := ioutil.ReadAll(f)
	//encode to string
	fileString := base64.StdEncoding.EncodeToString(fileData)

	fmt.Println("****************************")
	fmt.Println(fileString)
	fmt.Println("****************************")
	resp := CreateResponse{}
	resp.Error = ""

	output, err := json.Marshal(NewUser)
	fmt.Println(string(output))
	if err != nil {
		resp.Error += "\n*" + err.Error()
	}
	sql := fmt.Sprintf("INSERT INTO users SET user_nickname='%s', "+
		"user_first='%s', user_last='%s', user_email='%s', user_image='%s'",
		NewUser.Username, NewUser.First, NewUser.Last, NewUser.Email, fileString)
	q, err := database.Exec(sql)
	if err != nil {
		errorMessage, errorCode := dbErrorParse(err.Error())
		fmt.Println(errorMessage)
		errMsg := ErrorMessages(errorCode)
		resp.Error = errMsg.Msg
		resp.ErrorCode = string(errMsg.ErrCode)
		fmt.Println(errMsg.StatusCode)
		http.Error(w, "Conflict", errMsg.StatusCode)
	}
	fmt.Println(q)
	createOutput, _ := json.Marshal(resp)
	fmt.Fprintln(w, string(createOutput))
}

//UserGet endpoint for getting a specific user
func UserGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

//UserUpdate endpoint for updating a specific user
func UserUpdate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

//UserDelete endpoint for deleting a specific user
func UserDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func dbErrorParse(err string) (string, int64) {
	Parts := strings.Split(err, ":")
	errorMessage := Parts[1]
	Code := strings.Split(Parts[0], "Error ")
	errorCode, _ := strconv.ParseInt(Code[1], 10, 32)
	return errorMessage, errorCode
}

//ErrorMessages will return an ErrMsg object containing
//all the details about an error
func ErrorMessages(err int64) ErrMsg {
	var em ErrMsg
	errorMessage := ""
	statusCode := 200
	errorCode := 0
	switch err {
	case 1062:
		errorMessage = "Duplicate Entry"
		errorCode = 10
		statusCode = 409
	}

	em.ErrCode = errorCode
	em.StatusCode = statusCode
	em.Msg = errorMessage

	return em
}
