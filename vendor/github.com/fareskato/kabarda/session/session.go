package session

import (
	"database/sql"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Session Kabarda framework session type
type Session struct {
	CookieName     string
	CookieDomain   string
	CookieLifeTime string
	CookiePersist  string
	CookieSecure   string
	SessionType    string
	DBPool         *sql.DB
	RedisPool      *redis.Pool
}

// InitSession Kabarda session initialization: store session in different stores
func (s *Session) InitSession() *scs.SessionManager {
	var secure, persist bool
	// session duration
	minutes, err := strconv.Atoi(s.CookieLifeTime)
	if err != nil {
		minutes = 60
	}
	// persist
	if strings.ToLower(s.CookiePersist) == "true" {
		persist = true
	}
	// secure
	if strings.ToLower(s.CookieSecure) == "true" {
		secure = true
	}

	// create the session
	session := scs.New()
	session.Lifetime = time.Duration(minutes) + time.Minute
	session.Cookie.Persist = persist
	session.Cookie.Secure = secure
	session.Cookie.Name = s.CookieName
	session.Cookie.Domain = s.CookieDomain
	session.Cookie.SameSite = http.SameSiteLaxMode
	// session store: redis, mysql ...etc
	switch strings.ToLower(s.SessionType) {
	case "redis":
		session.Store = redisstore.New(s.RedisPool)
	case "mysql", "mariadb":
		session.Store = mysqlstore.New(s.DBPool)
	case "postgres", "postgresql":
		session.Store = postgresstore.New(s.DBPool)
	default:

	}

	return session
}
