package main

import (
    "database/sql"
    "fmt"
    "log"

    "encoding/json"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
)

type App struct {
    Router *mux.Router
    DB     *sql.DB
}

func (a *App) Initialize(user, password, dbname string) {
    connectionString :=
        fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

    var err error
    a.DB, err = sql.Open("postgres", connectionString)
    if err != nil {
        log.Fatal(err)
    }

    a.Router = mux.NewRouter()

    a.initializeRoutes()
}

func (a *App) Run(addr string) {
    log.Fatal(http.ListenAndServe(":8010", a.Router))
}

func (a *App) getTimeSpent(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid user ID")
        return
    }

    u := user{ID: id}
    if err := u.getTimeSpent(a.DB); err != nil {
        switch err {
        case sql.ErrNoRows:
            respondWithError(w, http.StatusNotFound, "User not found")
        default:
            respondWithError(w, http.StatusInternalServerError, err.Error())
        }
        return
    }

    respondWithJSON(w, http.StatusOK, u)
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid user ID")
        return
    }

    u := user{ID: id}
    if err := u.getUser(a.DB); err != nil {
        switch err {
        case sql.ErrNoRows:
            respondWithError(w, http.StatusNotFound, "User not found")
        default:
            respondWithError(w, http.StatusInternalServerError, err.Error())
        }
        return
    }

    respondWithJSON(w, http.StatusOK, u)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
    respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    response, _ := json.Marshal(payload)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(response)
}

func (a *App) getUsers(w http.ResponseWriter, r *http.Request) {
    count, _ := strconv.Atoi(r.FormValue("count"))
    start, _ := strconv.Atoi(r.FormValue("start"))

    if count > 10 || count < 1 {
        count = 10
    }
    if start < 0 {
        start = 0
    }

    users, err := getUsers(a.DB, start, count)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, users)
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
    var u user
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&u); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }
    defer r.Body.Close()

    if err := u.createUser(a.DB); err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusCreated, u)
}


func (a *App) updateTimeSpent(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid product ID")
        return
    }

    var u user
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&u); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
        return
    }
    defer r.Body.Close()
    u.ID = id

    if err := u.updateTimeSpent(a.DB); err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, u)
}

func (a *App) updateUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid product ID")
        return
    }

    var u user
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&u); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
        return
    }
    defer r.Body.Close()
    u.ID = id

    if err := u.updateUser(a.DB); err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, u)
}

func (a *App) deleteUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid Product ID")
        return
    }

    u := user{ID: id}
    if err := u.deleteUser(a.DB); err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}


func (a *App) initializeRoutes() {
    a.Router.HandleFunc("/user", a.createUser).Methods("POST")
    a.Router.HandleFunc("/users", a.getUsers).Methods("GET")
    a.Router.HandleFunc("/user/{id:[0-9]+}", a.getUser).Methods("GET")
    a.Router.HandleFunc("/timespent/{id:[0-9]+}", a.getTimeSpent).Methods("GET")
    a.Router.HandleFunc("/user/{id:[0-9]+}", a.updateUser).Methods("PUT")
    a.Router.HandleFunc("/timespent/{id:[0-9]+}", a.updateTimeSpent).Methods("PUT")
    a.Router.HandleFunc("/user/{id:[0-9]+}", a.deleteUser).Methods("DELETE")
}