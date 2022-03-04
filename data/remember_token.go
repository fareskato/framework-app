package data

import (
	up "github.com/upper/db/v4"
	"time"
)

type RememberToken struct {
	ID            int       `db:"id,omitempty"`
	UserID        int       `db:"user_id"`
	RememberToken string    `db:"remember_token"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func (t *RememberToken) Table() string {
	return "remember_tokens"
}

func (t *RememberToken) Insert(userId int, token string) error {
	collection := upper.Collection(t.Table())
	// create remember token instance
	rt := RememberToken{
		UserID:        userId,
		RememberToken: token,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_, err := collection.Insert(rt)
	if err != nil {
		return err
	}
	return nil
}

func (t *RememberToken) Delete(rt string) error {
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"remember_token": rt})
	err := res.Delete()
	if err != nil {
		return err
	}
	return nil
}