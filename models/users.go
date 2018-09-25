package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
	"lenslocked.com/hash"
	"lenslocked.com/rand"
)

var (
	ErrNotFound        = errors.New("models: resource not found")
	ErrInvalidID       = errors.New("models: ID provided was invalid")
	ErrInvalidPassword = errors.New("models: incorrect password provided")
	userPwPepper       = "secret-random-string"
)

const hmacSecretKey = "secret-hmac-key"

type User struct {
	gorm.Model
	Name         string
	Email        string `gorm: "not null; unique_index"`
	Password     string `gorm: "-"`
	PasswordHash string `gorm: "not null"`
	Remember     string `gorm: "-"`
	RememberHash string `gorm: "not null; unqiue_index"`
}

type UserDB interface {
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)

	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error

	Close() error

	AutoMigrate() error
}

type userGorm struct {
	db   *gorm.DB
	hmac hash.HMAC
}

type UserService interface {
	UserDB
	Authenticate(email, password string) (*User, error)
}

type userService struct {
	UserDB
}

type userValidator struct {
	UserDB
}

func NewUserService(connectionInfo string) (UserService, error) {
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}

	return &userService{
		UserDB: userValidator{
			UserDB: ug,
		},
	}, nil
}

func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}

	db.LogMode(true)
	hmac := hash.NewHMAC(hmacSecretKey)
	return &userGorm{
		db:   db,
		hmac: hmac,
	}, nil
}

func (us *userGorm) Create(user *User) error {
	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""

	if user.Remember != "" {
		panic(errors.New("user's remember is not empty'"))
	}

	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	user.RememberHash = us.hmac.Hash(token)

	return us.db.Create(user).Error
}

// struct userService implements this function, which is declared by interface UserService
// By Golang implicit interface implementation, this means userService has implemented interface UserService
func (us *userService) Authenticate(email string, password string) (*User, error) {
	foundUser, err := us.UserDB.ByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password+userPwPepper))
	switch err {
	case nil:
		return foundUser, nil
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, ErrInvalidPassword
	default:
		return nil, err
	}
}

func (us *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (us *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

func (us *userGorm) ByRemember(token string) (*User, error) {
	var user User
	rememberHash := us.hmac.Hash(token)
	err := first(us.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (us *userGorm) Update(user *User) error {
	if user.Remember != "" {
		user.RememberHash = us.hmac.Hash(user.Remember)
	}
	return us.db.Save(user).Error
}

func (us *userGorm) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}

	user := User{Model: gorm.Model{ID: id}}
	return us.db.Delete(&user).Error
}

func (us *userGorm) DestructiveReset() error {
	err := us.db.DropTableIfExists(&User{}).Error
	if err != nil {
		return nil
	}

	return us.AutoMigrate()
}

func (us *userGorm) AutoMigrate() error {
	if err := us.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}

	return nil
}

func (us *userGorm) Close() error {
	return us.db.Close()
}

func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}

	return err
}
