package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
)

var database *sql.DB

type User struct {
    ID int `json:"id"`
    Username string `json:"username"`
    Email string `json:"email"`
    First string `json:"first"`
    Last string `json:"last"`
}

type Users struct {
    Users []User `json:"users"`
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
	routes.GET("/api/users", UsersGet)   //get all users
	routes.POST("/api/users", UsersCreate) //create user

	routes.GET("/api/users/:id", UserGet)   //get a specific user
	routes.POST("/api/users/:id", UserUpdate) //update a specific user

	routes.DELETE("/api/users/:id", UserDelete) //delete a user

	log.Fatal(http.ListenAndServe(":8080", routes))
}

func OptionsGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprintln(w, "Available Endpoing/s: \n * /api/users")
}

func UsersGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    log.Println("starting retrieval")
    start := 0
    limit := 10

    next := start + limit

    w.Header().Set("Pragma", "no-cache")
    w.Header().Set("Link", "<http://localhost:8080/api/users?start="+string(next)+"; rel=\"next\"")

    rows, _ := database.Query("SELECT * FROM users LIMIT 10")

    users := Users{}

    for rows.Next(){
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

func UsersCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    r.ParseForm()
    NewUser := User{}
    NewUser.Username = r.Form.Get("username")
    NewUser.Email = r.Form.Get("email")
    NewUser.First = r.Form.Get("first")
    NewUser.Last = r.Form.Get("last")

    output, err := json.Marshal(NewUser)
    fmt.Println(string(output))
    if err != nil {
        fmt.Fprintln(w, "Something went wrong while processing your request...")
    }

    sql := fmt.Sprintf("INSERT INTO users SET user_nickname='%s', " +
                       "user_first='%s', user_last='%s', user_email='%s'",
                     NewUser.Username, NewUser.First, NewUser.Last, NewUser.Email)

    _, err = database.Exec(sql)
    if err != nil {
        fmt.Fprintln(w, "Something went wrong while we process your request: ", err)
    }
    fmt.Fprintf(w, "Things went fine...")
}

func UserGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func UserUpdate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func UserDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}
