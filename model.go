package main

import (
    "database/sql"
)

type user struct {
    ID    int     `json:"id"`
    Name  string  `json:"name"`
    TimeSpent float64 `json:"timespent"`
}

func (u *user) getTimeSpent(db *sql.DB) error {
    return db.QueryRow("SELECT name, timespent FROM users WHERE id=$1",
        u.ID).Scan(&u.Name, &u.TimeSpent)
}

func (u *user) getUser(db *sql.DB) error {
    return db.QueryRow("SELECT name, timespent FROM users WHERE id=$1",
        u.ID).Scan(&u.Name, &u.TimeSpent)
}

func (u *user) updateTimeSpent(db *sql.DB) error {
    _, err :=
        db.Exec("UPDATE users SET timespent=$1 WHERE id=$2",
            u.TimeSpent, u.ID)

    return err
}

func (u *user) updateUser(db *sql.DB) error {
    _, err :=
        db.Exec("UPDATE users SET timespent=$1 WHERE id=$2",
            u.TimeSpent, u.ID)

    return err
}

func (u *user) deleteUser(db *sql.DB) error {
    _, err := db.Exec("DELETE FROM users WHERE id=$1", u.ID)

    return err
}

func (u *user) createUser(db *sql.DB) error {
    err := db.QueryRow(
        "INSERT INTO users(name, timespent) VALUES($1, $2) RETURNING id",
        u.Name, u.TimeSpent).Scan(&u.ID)

    if err != nil {
        return err
    }

    return nil
}

func getUsers(db *sql.DB, start, count int) ([]user, error) {
    rows, err := db.Query(
        "SELECT id, name,  timespent FROM users LIMIT $1 OFFSET $2",
        count, start)

    if err != nil {
        return nil, err
    }

    defer rows.Close()

    users := []user{}

    for rows.Next() {
        var u user
        if err := rows.Scan(&u.ID, &u.Name, &u.TimeSpent); err != nil {
            return nil, err
        }
        users = append(users, u)
    }

    return users, nil
}