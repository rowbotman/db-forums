package db

import (
	"database/sql"
	"errors"
	"fmt"
)

type User struct {
	Pk       int64      `json:"-"`         // why we used '-' here?
	Nickname string     `json:"nickname,omitempty"`
	Name     string     `json:"fullname,omitempty"`
	About    string     `json:"about,omitempty"`
	Email    string     `json:"email,omitempty"`
}

func (us *User)IsEmpty() bool {
	if len(us.Email) == 0 &&
		len(us.Name) == 0 && len(us.About) == 0 {
		return true
	}
	return false
}

func InsertIntoUser(userData User) ([]User, error) {
	var users []User
	sqlStatement := `SELECT full_name, nickname, email, about FROM profile WHERE LOWER(nickname) = LOWER($1) OR LOWER(email) = LOWER($2);`
	fmt.Println(sqlStatement)
	rows, err := DB.Query(sqlStatement, userData.Nickname, userData.Email)
	if err != nil && !rows.Next() {
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
	for rows.Next() {
		newUser := User{}
		err = rows.Scan(
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

	if len(users) == 0 {
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

	return users, errors.New("multiple rows")
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
		return User{}, errors.New("no rows")
	} else if err != nil {
		fmt.Println("errors occurred")
		return User{}, err
	}
	fmt.Println("request complete, user is", newUser)
	return newUser, nil
}

func UpdateUser(updUser User) (User, error) {
	fmt.Println("update user is starting...")
	sqlStatement := `
  SELECT full_name, nickname, email FROM profile WHERE LOWER(email) = LOWER($1);`
	row := DB.QueryRow(sqlStatement, updUser.Email)
	user := User{}
	err := row.Scan(
		&user.Name,
		&user.Nickname,
		&user.Email)
	fmt.Println(user)
	if err == sql.ErrNoRows || user.Nickname == updUser.Nickname {
		if updUser.IsEmpty() {
			userInfo, err := SelectUser(updUser.Nickname)
			if err != nil {
				return User{}, err
			}
			return userInfo, nil
		}
		userInfo, err := SelectUser(updUser.Nickname)
		if err != nil {
			return User{}, err
		}
		if len(updUser.About) == 0 {
			updUser.About = userInfo.About
		}
		if len(updUser.Name) == 0 {
			updUser.Name = userInfo.Name
		}
		if len(updUser.Email) == 0 {
			updUser.Email = userInfo.Email
		}
		sqlStatement = `UPDATE profile SET full_name = $1, email = $2, about = $3 WHERE nickname = $4;`
		_, err = DB.Exec(sqlStatement, updUser.Name, updUser.Email, updUser.About, updUser.Nickname)
		if err != nil {
			return User{}, err
		}
		return updUser, nil
	}
	fmt.Println("i am return exist user")
	return updUser, errors.New("This email is already registered by user: " + user.Nickname)
}