package main

import (
	"fmt"
	_ "github.com/lib/pq"
	"lenslocked.com/models"
)

const (
	host   = "localhost"
	post   = 5432
	user   = "desmond"
	dbname = "lenslocked_dev"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, post, user, dbname)
	us, err := models.NewUserService(psqlInfo)
	if err != nil {
		panic(err)
	}

	defer us.Close()
	us.DestructiveReset()

	user := models.User{
		Name:  "Michael Scott",
		Email: "michael@dundermifflin.com",
	}
	if err := us.Create(&user); err != nil {
		panic(err)
	}

	foundUser, err := us.ByID(1)
	if err != nil {
		panic(err)
	}

	fmt.Println(foundUser)

	user.Name = "Updated Name"
	if err := us.Update(&user); err != nil {
		panic(err)
	}

	emailUser, err := us.ByEmail("michael@dundermifflin.com")
	if err != nil {
		panic(err)
	}

	fmt.Println(emailUser)

	if err := us.Delete(foundUser.ID); err != nil {
		panic(err)
	}

	_, err = us.ByID(foundUser.ID)
	if err != models.ErrNotFound {
		panic("user was not deleted!")
	}
}
