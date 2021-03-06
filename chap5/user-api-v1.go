package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"

	"github.com/steven-ferrer/go-web-services/chap5/password"
	"github.com/steven-ferrer/go-web-services/chap5/pseudoauth"
	apispec "github.com/steven-ferrer/go-web-services/chap5/specification"
)

var database *sql.DB

//User user
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	First    string `json:"first"`
	Last     string `json:"last"`
	Password string `json:"password"`
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

type DocMethod interface{}

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
	routes.ServeFiles("/src/*filepath", http.Dir("../"))
	routes.GET("/api", OptionsGet)
	routes.OPTIONS("/api/users", UsersInfo)
	routes.GET("/api/users", UsersGet)     //get all users
	routes.POST("/api/users", UsersCreate) //create user

	routes.GET("/api/users/:id", UserGet)    //get a specific user
	routes.PUT("/api/users/:id", UserUpdate) //update a specific user
	routes.DELETE("/api/users/:id", UserDelete)

	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", routes))
}

//UsersInfo options endpoint for users api
func UsersInfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Allow", "DELETE, GET,HEAD,OPTIONS,POST,PUT")
	UserDocumentation := []DocMethod{}
	UserDocumentation = append(UserDocumentation,
		apispec.UserPOST, apispec.UserOPTIONS, apispec.UserGET)
	output, _ := json.Marshal(UserDocumentation)
	fmt.Fprintln(w, string(output))
}

//OptionsGet default endpoint, displays the available endpoints
func OptionsGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintln(w, "Available Endpoing/s: \n * /api/users")
}

func CheckCredentials(w http.ResponseWriter, r *http.Request) {
	var Credentials string
	response := CreateResponse{}
	consumerKey := r.FormValue("consumer_key")
	fmt.Println("Consumer Key:", consumerKey)
	timestamp := r.FormValue("timestamp")
	signature := r.FormValue("signature")
	nonce := r.FormValue("nonce")
	err := database.QueryRow("SELECT consumer_secret FROM api_credentials WHERE consumer_key=?",
		consumerKey).Scan(&Credentials)

	if err != nil {
		errMsg := ErrorMessages(404)
		log.Println(errMsg.ErrCode)
		log.Println(w, errMsg.Msg, errMsg.StatusCode)
		response.Error = errMsg.Msg
		response.ErrorCode = string(errMsg.StatusCode)
		http.Error(w, errMsg.Msg, errMsg.StatusCode)
		return
	}

	token, err := pseudoauth.ValidateSignature(consumerKey, Credentials,
		timestamp, nonce, signature, 0)
	if err != nil {
		errMsg := ErrorMessages(401)
		log.Println(errMsg.ErrCode)
		log.Println(w, errMsg.Msg, errMsg.StatusCode)
		response.Error = errMsg.Msg
		response.ErrorCode = string(errMsg.StatusCode)
		http.Error(w, errMsg.Msg, errMsg.StatusCode)
	}
}

//UsersGet endpoint for getting all users
func UsersGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Println("starting retrieval")
	start := 0
	limit := 10

	next := start + limit

	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Link", "<http://localhost:8080/api/users?start="+string(next)+
		"; rel=\"next\"")

	rows, _ := database.Query("SELECT user_id, user_nickname, user_first, user_last, " +
		"user_email FROM users ORDER BY user_id DESC LIMIT 10")

	users := Users{}

	for rows.Next() {
		user := User{}
		rows.Scan(&user.ID, &user.Username, &user.First, &user.Last, &user.Email)
		users.Users = append(users.Users, user)
	}

	output, err := json.Marshal(users)
	if err != nil {
		fmt.Fprintln(w, "Something went wrong while processing your request: ", err.Error())
	}

	fmt.Fprintln(w, string(output))
}

//UsersCreate endpoint for creating a user
func UsersCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	//freakin' Access-Control-Allow-Origin
	//Still looking for a fix for this
	//w.Header().Set("Access-Control-Allow-Origin", "http://localhost:9000")
	//r.Header.Set("Origin", "http://localhost:9000")
	//fmt.Println(r.Header.Get("Origin"))
	NewUser := User{}
	/*fmt.Println(r.FormFile("userImage"))
	f, _, err := r.FormFile("userImage")
	if err != nil {
		fmt.Println(err.Error())
	}*/

	NewUser.Username = r.FormValue("username")
	NewUser.Email = r.FormValue("email")
	NewUser.First = r.FormValue("first")
	NewUser.Last = r.FormValue("last")
	NewUser.Password = r.FormValue("password")

	//returns a byte
	//fileData, _ := ioutil.ReadAll(f)
	//encode to string
	//fileString := base64.StdEncoding.EncodeToString(fileData)

	//fmt.Println("****************************")
	//fmt.Println(fileString)
	//fmt.Println("****************************")
	resp := CreateResponse{}
	resp.Error = ""

	output, err := json.Marshal(NewUser)
	fmt.Println(string(output))
	if err != nil {
		resp.Error += "\n*" + err.Error()
	}

	pwdSalt, pwdhash := password.ReturnPassword(NewUser.Password)
	fmt.Println(pwdSalt)
	fmt.Println(pwdhash)

	sql := "INSERT INTO users SET user_nickname=?, user_first=?, user_last=?, " +
		"user_email=?, user_password=?, user_salt=?"
	q, err := database.Exec(sql, NewUser.Username, NewUser.First, NewUser.Last,
		NewUser.Email, pwdhash, pwdSalt)
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
	id := ps.ByName("id")
	var user User

	userID, err := strconv.Atoi(id)

	if err != nil {
		fmt.Fprintln(w, "Error parsing id: ", err)
		return
	}

	sql := "SELECT user_id, user_nickname, user_first, user_last FROM users WHERE user_id=?"
	err = database.QueryRow(sql, userID).Scan(&user.ID, &user.Username, &user.First, &user.Last)
	if err != nil {
		fmt.Fprintln(w, "Error getting user: ", err.Error())
		return
	}

	output, err := json.Marshal(user)
	if err != nil {
		fmt.Fprintln(w, "Error encoding user: ", err.Error())
		return
	}

	fmt.Fprintln(w, string(output))
}

//UserUpdate endpoint for updating a specific user
func UserUpdate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	response := CreateResponse{}
	id := ps.ByName("id")
	email := r.FormValue("email")

	var userCount int

	err := database.QueryRow("SELECT COUNT(user_id) FROM users WHERE user_id=?", id).Scan(&userCount)

	if userCount == 0 {
		errMsg := ErrorMessages(404)
		log.Println(errMsg.ErrCode)
		log.Println(w, errMsg.Msg, errMsg.StatusCode)
		fmt.Fprintln(w, errMsg.Msg, " : ", errMsg.StatusCode)
		return
	} else if err != nil {
		log.Println(err)
		return
	} else {
		_, err = database.Exec("UPDATE users SET user_email=? WHERE user_id=?", email, id)

		if err != nil {
			_, errorCode := dbErrorParse(err.Error())
			errMsg := ErrorMessages(errorCode)

			http.Error(w, errMsg.Msg, errMsg.StatusCode)
		} else {
			response.Error = "success"
			response.ErrorCode = "0"
			output, _ := json.Marshal(response)
			fmt.Fprintln(w, string(output))
		}
	}
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
	var errorMessage string
	var statusCode int
	var errorCode int
	switch err {
	case 1062:
		errorMessage = http.StatusText(http.StatusConflict)
		errorCode = 10
		statusCode = http.StatusConflict
	default:
		errorMessage = http.StatusText(int(err))
		errorCode = 0
		statusCode = int(err)
	}

	em.ErrCode = errorCode
	em.StatusCode = statusCode
	em.Msg = errorMessage

	return em
}
