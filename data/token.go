package data

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	up "github.com/upper/db/v4"
	"net/http"
	"strings"
	"time"
)

type Token struct {
	ID        int       `db:"id" json:"id"`
	UserId    int       `db:"user_id" json:"user_id"`
	FirstName string    `db:"first_name" json:"first-name"`
	Email     string    `db:"email" json:"email"`
	PlainText string    `db:"token" json:"token"`
	Hash      []byte    `db:"token_hash" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Expires   time.Time `db:"expiry" json:"expiry"`
}

func (t *Token) Table() string {
	return "tokens"
}

// GetUserByToken fetch user by token
func (t *Token) GetUserByToken(token string) (*User, error) {
	var user User
	var theToken Token

	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"token": token})
	err := res.One(&token)
	if err != nil {
		return nil, err
	}
	// fetch user for token
	collection = upper.Collection("users")
	res = collection.Find(up.Cond{"id": theToken.UserId})
	err = res.One(&user)
	if err != nil {
		return nil, err
	}
	user.Token = theToken
	return &user, nil
}

// GetTokensForUser fetch all user tokens  by user id
func (t *Token) GetTokensForUser(userId int) ([]*Token, error) {
	var tokens []*Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"user_id": userId})
	err := res.All(&tokens)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// GetTokenById fetch token by id
func (t *Token) GetTokenById(tokenId int) (*Token, error) {
	var token Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"id": tokenId})
	err := res.One(&token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetByToken fetch token by token string
func (t *Token) GetByToken(tokenStr string) (*Token, error) {
	var token Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"token": tokenStr})
	err := res.One(&token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteById delete token by id: uses when user logout
func (t *Token) DeleteById(id int) error {
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"id": id})
	err := res.Delete()
	if err != nil {
		return err
	}
	return nil
}

// DeleteByToken delete token by token plain test: uses when user logout
func (t *Token) DeleteByToken(tokenStr string) error {
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"token": tokenStr})
	err := res.Delete()
	if err != nil {
		return err
	}
	return nil
}

// Insert create new token
func (t *Token) Insert(token Token, user User) error {
	collection := upper.Collection(t.Table())
	// delete all existing tokens for user
	res := collection.Find(up.Cond{"user_id": user.ID})
	err := res.Delete()
	if err != nil {
		return err
	}
	// populate new token fields(other fields will be populated by frontend user)
	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()
	token.FirstName = user.FirstName
	token.Email = user.Email

	// insert the new populated token
	_, err = collection.Insert(token)
	if err != nil {
		return err
	}
	return nil
}

// GenerateToken will generate token for user and expiry date ir depends on the td(token life time)
func (t *Token) GenerateToken(userId int, td time.Duration) (*Token, error) {
	token := &Token{
		UserId:  userId,
		Expires: time.Now().Add(td),
	}
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	// generate the plain text
	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	// generate the hash
	hash := sha256.Sum256([]byte(token.PlainText))
	// Note that by default base-32 strings may be padded at the end with the =
	// character. We don't need this padding character for the purpose of our tokens, so
	// we use the WithPadding(base32.NoPadding) method in the line below to omit them.
	token.Hash = hash[:]
	return token, nil
}

// AuthenticateToken checks if token exists for certain user
func (t *Token) AuthenticateToken(r *http.Request) (*User, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("no authorization header received")
	}
	headerParts := strings.Split(authHeader, " ")
	// Bearer tokens enable requests to authenticate using an access key, such as a JSON Web Token (JWT)
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, errors.New("no authorization header received")
	}
	// now we have the token
	token := headerParts[1]
	// every token must has length of 26 character
	if len(token) != 26 {
		return nil, errors.New("token wrong size")
	}
	// get token from DB
	theToken, err := t.GetByToken(token)
	if err != nil {
		return nil, errors.New(" no matching token found")
	}
	// if the token is expired
	if theToken.Expires.Before(time.Now()) {
		return nil, errors.New("expired token")
	}
	// get the user
	user, err := theToken.GetUserByToken(token)
	if err != nil {
		return nil, errors.New("no user found for this token")
	}
	// all good
	return user, nil
}

func (t *Token) ValidateToken(token string) (bool, error) {
	// get the user
	user, err := t.GetUserByToken(token)
	if err != nil {
		return false, errors.New("no user found for this token")
	}
	// if token is empty string
	if user.Token.PlainText == "" {
		return false, errors.New("no matching token found")
	}
	// check token expiry
	if user.Token.Expires.Before(time.Now()) {
		return false, errors.New("expired token")
	}
	return true, nil
}
