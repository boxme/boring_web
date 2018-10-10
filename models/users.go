package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
	"lenslocked.com/hash"
	"lenslocked.com/rand"
	"regexp"
	"strings"
)

const (
	userPwPepper  = "secret-random-string"
	hmacSecretKey = "secret-hmac-key"
)

const (
	ErrNotFound          modelError = "models: resource not found"
	ErrIDInvalid         modelError = "models: ID provided was invalid"
	ErrPasswordIncorrect modelError = "models: incorrect " +
		"password provided"
	ErrPasswordTooShort modelError = "models: password must " +
		"be at least 8 characters long"
	ErrPasswordRequired modelError = "models: password is required"
	ErrEmailRequired    modelError = "models: email address is " +
		"required"
	ErrEmailInvalid modelError = "models: email address is " +
		"not valid"
	ErrEmailTaken modelError = "models: email address is " +
		"already taken"
	ErrRememberRequired modelError = "models: remember token " +
		"is required"
	ErrRememberTooShort modelError = "models: remember token " +
		"must be at least 32 bytes"
)

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
}

type userGorm struct {
	db *gorm.DB
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
	hmac       hash.HMAC
	emailRegex *regexp.Regexp
}

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1)
	split := strings.Split(s, " ")
	split[0] = strings.Title(split[0])
	return strings.Join(split, " ")
}

// Declare function type
type userValFn func(*User) error

func NewUserService(db *gorm.DB) UserService {
	ug := &userGorm{
		db: db,
	}
	hmac := hash.NewHMAC(hmacSecretKey)
	uv := newUserValidator(ug, hmac)
	return &userService{
		UserDB: uv,
	}
}

func newUserValidator(udb UserDB, hmac hash.HMAC) *userValidator {
	return &userValidator{
		UserDB:     udb,
		hmac:       hmac,
		emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`),
	}
}

func (uv *userValidator) Create(user *User) error {
	if user.Remember != "" {
		panic(errors.New("user's remember is not empty'"))
	}

	err := runUserValFns(
		user,
		uv.passwordRequired,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.setRememberIfUnset,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.requireEmail,
		uv.normalizeEmail,
		uv.emailFormat,
		uv.emailIsAvail)
	if err != nil {
		return err
	}

	return uv.UserDB.Create(user)
}

func (uv *userValidator) setRememberIfUnset(user *User) error {
	if user.Remember != "" {
		return nil
	}

	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}

func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		// Do not run this if password hasn't been changed
		return nil
	}

	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""
	return nil
}

func (us *userGorm) Create(user *User) error {
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
		return nil, ErrPasswordIncorrect
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

func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := User{
		Email: email,
	}
	err := runUserValFns(&user, uv.normalizeEmail)
	if err != nil {
		return nil, err
	}
	return uv.UserDB.ByEmail(user.Email)
}

func (us *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := User{
		Remember: token,
	}
	if err := runUserValFns(&user, uv.hmacRemember); err != nil {
		return nil, err
	}
	return uv.UserDB.ByRemember(user.RememberHash)
}

func (us *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User
	err := first(us.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (uv *userValidator) Update(user *User) error {
	err := runUserValFns(
		user,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.requireEmail,
		uv.normalizeEmail,
		uv.emailFormat,
		uv.emailIsAvail)
	if err != nil {
		return err
	}
	return uv.UserDB.Update(user)
}

func (us *userGorm) Update(user *User) error {
	return us.db.Save(user).Error
}

func (uv *userValidator) Delete(id uint) error {
	var user User
	user.ID = id
	err := runUserValFns(&user, uv.idGreaterThan(0))
	if err != nil {
		return err
	}

	return uv.UserDB.Delete(id)
}

// Cast closure to userValFn type
func (uv *userValidator) idGreaterThan(n uint) userValFn {
	return userValFn(func(user *User) error {
		if user.ID <= n {
			return ErrIDInvalid
		}
		return nil
	})
}

func (us *userGorm) Delete(id uint) error {
	user := User{Model: gorm.Model{ID: id}}
	return us.db.Delete(&user).Error
}

func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}

	return err
}

func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

func (uv *userValidator) emailFormat(user *User) error {
	if user.Email == "" {
		return nil
	}

	if !uv.emailRegex.MatchString(user.Email) {
		return ErrEmailInvalid
	}
	return nil
}

func (uv *userValidator) emailIsAvail(user *User) error {
	existing, err := uv.ByEmail(user.Email)
	if err == ErrNotFound {
		return nil
	}

	if err != nil {
		return err
	}

	if user.ID != existing.ID {
		return ErrEmailTaken
	}

	return nil
}

func (uv *userValidator) passwordMinLength(user *User) error {
	if user.Password == "" {
		return nil
	}

	if len(user.Password) < 8 {
		return ErrPasswordTooShort
	}

	return nil
}

func (uv *userValidator) passwordRequired(user *User) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}

	return nil
}

func (uv *userValidator) passwordHashRequired(user *User) error {
	if user.PasswordHash == "" {
		return ErrPasswordRequired
	}

	return nil
}

func (uv *userValidator) rememberMinBytes(user *User) error {
	if user.Remember == "" {
		return nil
	}

	n, err := rand.NBytes(user.Remember)
	if err != nil {
		return err
	}

	if n < 32 {
		return ErrRememberTooShort
	}

	return nil
}

func (uv *userValidator) rememberHashRequired(user *User) error {
	if user.RememberHash == "" {
		return ErrRememberRequired
	}

	return nil
}

func runUserValFns(user *User, fns ...userValFn) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}
