package data

import (
	"errors"
	up "github.com/upper/db/v4"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID        int       `db:"id,omitempty"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	Active    int       `db:"user_active"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Token     Token     `db:"-"`
	SuperUser int       `db:"is_superuser"`
}

func (u *User) Table() string {
	return "users"
}

func (u *User) GetAll() ([]*User, error) {
	collection := upper.Collection(u.Table())
	var all []*User
	res := collection.Find().OrderBy("last_name")
	err := res.All(&all)
	if err != nil {
		return nil, err
	}
	return all, err
}

func (u *User) GetByEmail(email string) (*User, error) {
	var user User
	collection := upper.Collection(u.Table())
	res := collection.Find(up.Cond{"email": email})
	err := res.One(&user)
	if err != nil {
		switch {
		case errors.Is(err, up.ErrNoMoreRows), errors.Is(err, up.ErrNilRecord):
			return nil, ErrorRecordNotFound
		default:
			return nil, err
		}
	}
	// Get user token
	var token Token
	collection = upper.Collection(token.Table())
	res = collection.Find(up.Cond{"user_id": user.ID, "expiry >": time.Now()}).OrderBy("created_at desc")
	err = res.One(&token)
	if err != nil {
		if err != up.ErrNilRecord && err != up.ErrNoMoreRows {
			return nil, ErrorRecordNotFound
		}
	}
	user.Token = token
	return &user, nil
}

func (u *User) GetById(id int) (*User, error) {
	var user User
	collection := upper.Collection(u.Table())
	res := collection.Find(up.Cond{"id": id})
	err := res.One(&user)
	if err != nil {
		switch {
		case errors.Is(err, up.ErrNoMoreRows), errors.Is(err, up.ErrNilRecord):
			return nil, ErrorRecordNotFound
		default:
			return nil, err
		}
	}
	// Get user token
	var token Token
	collection = upper.Collection(token.Table())
	res = collection.Find(up.Cond{"user_id": user.ID, "expiry >": time.Now()}).OrderBy("created_at desc")
	err = res.One(&token)
	if err != nil {
		if err != up.ErrNilRecord && err != up.ErrNoMoreRows {
			return nil, err
		}
	}
	user.Token = token
	return &user, nil
}

func (u *User) Update(theUser User) error {
	// update updated_at field
	theUser.UpdatedAt = time.Now()

	collection := upper.Collection(u.Table())
	res := collection.Find(theUser.ID)
	err := res.Update(&theUser)
	if err != nil {
		switch err.Error() {
		case ErrorDuplicateEmailMessage:
			return ErrorDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (u *User) Delete(id int) error {
	collection := upper.Collection(u.Table())
	res := collection.Find(id)
	err := res.Delete()
	if err != nil {
		return err
	}
	return nil
}

func (u *User) Insert(theUser User) (int, error) {
	// hash user password
	newHash, err := bcrypt.GenerateFromPassword([]byte(theUser.Password), 12)
	if err != nil {
		return 0, err
	}
	theUser.CreatedAt = time.Now()
	theUser.UpdatedAt = time.Now()
	theUser.Password = string(newHash)

	collection := upper.Collection(u.Table())
	res, err := collection.Insert(theUser)
	if err != nil {
		switch err.Error() {
		case ErrorDuplicateEmailMessage:
			return 0, ErrorDuplicateEmail
		default:
			return 0, err
		}
	}
	recordId := getInsertedId(res.ID())
	return recordId, nil
}

func (u *User) ResetPassword(id int, newPassword string) error {
	// hash password
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	// get user by id
	theUser, err := u.GetById(id)
	if err != nil {
		return err
	}
	// update the password
	u.Password = string(newHash)
	// update user
	err = theUser.Update(*u)
	if err != nil {
		return err
	}
	return nil
}

// PasswordMatches checks if password matches the user password
func (u *User) PasswordMatches(password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, ErrorPasswordMatch
		}
	}
	return true, nil
}

func (u *User) CheckForRememberToken(id int, token string) bool {
	var rememberToken RememberToken
	rt := RememberToken{}
	collection := upper.Collection(rt.Table())
	res := collection.Find(up.Cond{"user_id": id, "remember_token": token})
	err := res.One(&rememberToken)
	return err == nil
}
