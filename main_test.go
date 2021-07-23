package main

import (
    "log"
    "os"
    "testing"

    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strconv"
)

var a App

func TestMain(m *testing.M) {
    a.Initialize(
        os.Getenv("APP_DB_USERNAME"),
        os.Getenv("APP_DB_PASSWORD"),
        os.Getenv("APP_DB_NAME"))

    ensureTableExists()
    code := m.Run()
    clearTable()
    os.Exit(code)
}

func ensureTableExists() {
    if _, err := a.DB.Exec(tableCreationQuery); err != nil {
        log.Fatal(err)
    }
}

func clearTable() {
    a.DB.Exec("DELETE FROM users")
    a.DB.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS users
(
    id SERIAL,
    name TEXT NOT NULL,
    timespent NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    CONSTRAINT users_pkey PRIMARY KEY (id)
)`


func TestEmptyTable(t *testing.T) {
    clearTable()

    req, _ := http.NewRequest("GET", "/users", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

    if body := response.Body.String(); body != "[]" {
        t.Errorf("Expected an empty array. Got %s", body)
    }
}


func executeRequest(req *http.Request) *httptest.ResponseRecorder {
    rr := httptest.NewRecorder()
    a.Router.ServeHTTP(rr, req)

    return rr
}


func checkResponseCode(t *testing.T, expected, actual int) {
    if expected != actual {
        t.Errorf("Expected response code %d. Got %d\n", expected, actual)
    }
}


func TestGetNonExistentTimeSpent(t *testing.T) {
    clearTable()

    req, _ := http.NewRequest("GET", "/timespent/11", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusNotFound, response.Code)

    var m map[string]string
    json.Unmarshal(response.Body.Bytes(), &m)
    if m["error"] != "User not found" {
        t.Errorf("Expected the 'error' key of the response to be set to 'User not found'. Got '%s'", m["error"])
    }
}


func TestCreateUser(t *testing.T) {

    clearTable()

    var jsonStr = []byte(`{"name":"test user", "timespent": 11.22}`)
    req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonStr))
    req.Header.Set("Content-Type", "application/json")

    response := executeRequest(req)
    checkResponseCode(t, http.StatusCreated, response.Code)

    var m map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &m)

    if m["name"] != "test user" {
        t.Errorf("Expected product name to be 'test product'. Got '%v'", m["name"])
    }

    if m["timespent"] != 11.22 {
        t.Errorf("Expected user timespent to be '11.22'. Got '%v'", m["timespent"])
    }

    // the id is compared to 1.0 because JSON unmarshaling converts numbers to
    // floats, when the target is a map[string]interface{}
    if m["id"] != 1.0 {
        t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
    }
}


func TestGetUser(t *testing.T) {
    clearTable()
    addUsers(1)

    req, _ := http.NewRequest("GET", "/user/1", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)
}

// main_test.go

func addUsers(count int) {
    if count < 1 {
        count = 1
    }

    for i := 0; i < count; i++ {
        a.DB.Exec("INSERT INTO users(name, timespent) VALUES($1, $2)", "User "+strconv.Itoa(i), (i+1.0)*10)
    }
}


func TestUpdateUser(t *testing.T) {

    clearTable()
    addUsers(1)

    req, _ := http.NewRequest("GET", "/user/1", nil)
    response := executeRequest(req)
    var originalProduct map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &originalProduct)

    var jsonStr = []byte(`{"name":"test product - updated name", "timespent": 11.22}`)
    req, _ = http.NewRequest("PUT", "/user/1", bytes.NewBuffer(jsonStr))
    req.Header.Set("Content-Type", "application/json")

    response = executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

    var m map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &m)

    if m["id"] != originalProduct["id"] {
        t.Errorf("Expected the id to remain the same (%v). Got %v", originalProduct["id"], m["id"])
    }

    if m["name"] == originalProduct["name"] {
        t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalProduct["name"], m["name"], m["name"])
    }

    if m["timespent"] == originalProduct["timespent"] {
        t.Errorf("Expected the timespent to change from '%v' to '%v'. Got '%v'", originalProduct["timespent"], m["timespent"], m["timespent"])
    }
}

func TestUpdateTimestamp(t *testing.T) {

    clearTable()
    addUsers(1)

    req, _ := http.NewRequest("GET", "/timespent/1", nil)
    response := executeRequest(req)
    var originalProduct map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &originalProduct)

    var jsonStr = []byte(`{"name":"test product - updated name", "timespent": 11.222}`)
    req, _ = http.NewRequest("PUT", "/timespent/1", bytes.NewBuffer(jsonStr))
    req.Header.Set("Content-Type", "application/json")

    response = executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

    var m map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &m)

    if m["id"] != originalProduct["id"] {
        t.Errorf("Expected the id to remain the same (%v). Got %v", originalProduct["id"], m["id"])
    }

    if m["name"] == originalProduct["name"] {
        t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalProduct["name"], m["name"], m["name"])
    }

    if m["timespent"] == originalProduct["timespent"] {
        t.Errorf("Expected the timespent to change from '%v' to '%v'. Got '%v'", originalProduct["timespent"], m["timespent"], m["timespent"])
    }
}

func TestDeleteUser(t *testing.T) {
    clearTable()
    addUsers(1)

    req, _ := http.NewRequest("GET", "/user/1", nil)
    response := executeRequest(req)
    checkResponseCode(t, http.StatusOK, response.Code)

    req, _ = http.NewRequest("DELETE", "/user/1", nil)
    response = executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

    req, _ = http.NewRequest("GET", "/user/1", nil)
    response = executeRequest(req)
    checkResponseCode(t, http.StatusNotFound, response.Code)
}
