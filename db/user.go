package db

import (
	"database/sql"
	"fmt"
)

type User struct {
	Pk       int64      `json:"-"`         // why we used '-' here?
	Nickname string     `json:"nickname"`
	Name     string     `json:"fullname"`
	About    string     `json:"about"`
	Email    string     `json:"email"`
}


func InsertIntoUser(userData User) ([]User, error) {
	var users []User
	sqlStatement := `SELECT full_name, nickname, email, about FROM profile WHERE nickname = '$1' OR email = '$2';`
	fmt.Println(sqlStatement)
	rows, err := DB.Query(sqlStatement, userData.Nickname, userData.Email)
	fmt.Println("i am here")
	if err != nil {
		sqlStatement = `INSERT INTO profile VALUES (default, $1, $2, $3, $4);`
		fmt.Println("request error")
		fmt.Println(sqlStatement)

		_, err = DB.Exec(sqlStatement, userData.Nickname, userData.Name, userData.About, userData.Email)

		if err != nil {
			fmt.Println("error with insertion")
			fmt.Println(err.Error())
			return nil, err
		}
		fmt.Println("insertion complete")
		fmt.Println(rows)
		users = append(users, userData)
		return users, nil
	}

	defer rows.Close()
	fmt.Println("and here")

	for rows.Next() {
		newUser := User{}
		err = rows.Scan(
			&newUser.Pk,
			&newUser.Name,
			&newUser.Nickname,
			&newUser.Email,
			&newUser.About)
		fmt.Println("i am alive")
		if err != nil {
			// handle this error
			// but i did't know how to do this .-.
			return nil, err
		}
		users = append(users, newUser)
	}
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func SelectUser(nickname string) (User, error) {
	fmt.Println("SELECT uid, full_name, nickname, email, about FROM profile WHERE nickname = ", nickname)
	sqlStatement := `SELECT uid, full_name, nickname, email, about FROM profile WHERE nickname = $1`
	row := DB.QueryRow(sqlStatement, nickname)
	newUser := User{}
	err := row.Scan(
		&newUser.Pk,
		&newUser.Name,
		&newUser.Nickname,
		&newUser.Email,
		&newUser.About)
	if err == sql.ErrNoRows {
		fmt.Println("No rows were returned!")
		return User{}, err
	} else if err != nil {
		fmt.Println("errors occurred")
		return User{}, err
	}
	fmt.Println("request complete, user is", newUser)
	return newUser, nil
}

func UpdateUser(updUser User) (User, error) {
	sqlStatement := `
  SELECT full_name, nickname, email FROM profile WHERE email = $2;`
	row := DB.QueryRow(sqlStatement, updUser.Email)
	user := User{}
	err := row.Scan(
		&user.Name,
		&user.Nickname,
		&user.Email)
	if err == sql.ErrNoRows || user.Nickname == updUser.Nickname {
		sqlStatement = `UPDATE profile u SET u.full_name = $1, u.email = $2, u.about = $3 WHERE u.nickname = $4;`
		_, err = DB.Exec(sqlStatement, updUser.Name, updUser.Email, updUser.About, updUser.Nickname)
		if err != nil {
			return User{}, err
		}
		return updUser, nil
	}
	return User{}, err
}