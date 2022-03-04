package signer

import (
	"fmt"
	goalone "github.com/bwmarrin/go-alone"
	"strings"
	"time"
)

type Signer struct {
	Secret []byte
}

func (s *Signer) GenerateTokenFromString(data string) string {
	var urlToSign string
	// Create a new Signer using our secret
	sr := goalone.New(s.Secret, goalone.Timestamp)
	// if url contains token: check if the url has query params after ?
	if strings.Contains(data, "?") {
		urlToSign = fmt.Sprintf("%s&hash=", data)
	} else {
		urlToSign = fmt.Sprintf("%s?hash=", data)
	}
	// Sign and return a token in the form of `data.signature`
	// You can reuse this struct as many times as you wish
	tokenBytes := sr.Sign([]byte(urlToSign))
	token := string(tokenBytes)

	return token
}

func (s *Signer) VerifyToken(token string) bool {
	// Create a new Signer using our secret
	sr := goalone.New(s.Secret, goalone.Timestamp)
	// You can easily Unsign a token, which will verify the signature is valid
	// then return signed data of the token
	_, err := sr.Unsign([]byte(token))
	if err != nil {
		return false
	}
	return true
}

func (s *Signer) Expired(token string, periodToExpire int) bool {
	// Create a new Signer using our secret
	sr := goalone.New(s.Secret, goalone.Timestamp)

	// You can parse out a token into a struct that separates the payload and
	// timestamp for you.
	ts := sr.Parse([]byte(token))

	// We can even check how old our timestamp is.  You can do a lot of other
	// things with the Timestamp too, since it's a standard time.Time value.
	return time.Since(ts.Timestamp) > time.Duration(periodToExpire)*time.Minute
}
