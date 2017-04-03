package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "net/http"
    "log"

    "github.com/gorilla/mux"
    _ "github.com/go-sql-driver/mysql"
)



type User struct {
    ID int `json:id`
    Username string `json:username`
    Email string `json:email`
    First string `json:first`
    Last string `json:last`
}

var Db *sql.DB

func main(){
    //initialize the database
    db, err := sql.Open("mysql", "root:password@/social_network")
    if err != nil {
        fmt.Println(err)
    }

    Db = db

    routes := mux.NewRouter()
    routes.HandleFunc("/api/user/create", CreateUser).Methods("GET")
    routes.HandleFunc("/api/user/{id:[0-9]+}", GetUser).Methods("GET")
    http.Handle("/", routes)
    log.Fatal(http.ListenAndServe(":6060", nil))
}

func CreateUser(w http.ResponseWriter, r *http.Request){
    NewUser := User{}
    NewUser.Username = r.FormValue("username")
    NewUser.Email = r.FormValue("email")
    NewUser.First = r.FormValue("first")
    NewUser.Last = r.FormValue("last")
    output, err := json.Marshal(NewUser)
    fmt.Println(string(output))
    if err != nil {
        fmt.Println("Something went wrong...")
    }

    sql := "INSERT INTO users set user_nickname='" + NewUser.Username +
           "', user_first='" + NewUser.First + "', user_last='" +
           NewUser.Last + "', user_email='" + NewUser.Email + "'"

    q, err := Db.Exec(sql)
    if err != nil {
        fmt.Println(err)
    }

    fmt.Println(q)
}

func GetUser(w http.ResponseWriter, r *http.Request){
    w.Header().Set("Pragma", "no-cache")
    
    urlParams := mux.Vars(r)
    id := urlParams["id"]
    User := User{}
    err := Db.QueryRow("SELECT * FROM users WHERE " +
        "user_id=?",id).Scan(&User.ID, &User.Username,
        &User.First, &User.Last, &User.Email)

    switch  {
    case err == sql.ErrNoRows:
        fmt.Fprintf(w, "No such user")
    case err != nil :
        log.Fatal(err)
        fmt.Fprintf(w, "Error")
    default:
        output, _ := json.Marshal(User)
        fmt.Println(string(output))
        fmt.Fprintf(w, string(output))
    }

}
