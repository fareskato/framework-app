package middlewares

import (
	"fmt"
	"myapp/data"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (m *Middleware) CheckRemember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get userID from the session(if exists)
		if !m.App.Session.Exists(r.Context(), "userID") {
			// user is not logged in
			// check for cookie: the cookie name is the app name
			cookie, err := r.Cookie(fmt.Sprintf("_%s_remember", m.App.AppName))
			if err != nil {
				next.ServeHTTP(w, r)
			} else {
				// cookie found
				val := cookie.Value
				// declare user instance
				var u data.User
				if len(val) > 0 {
					// there are some date in the cookie, so we need to  validate it
					split := strings.Split(val, "|")
					uidStr, hash := split[0], split[1]
					uid, _ := strconv.Atoi(uidStr)
					isValidToken := u.CheckForRememberToken(uid, hash)
					// if not valid cookie, so delete it
					if !isValidToken {
						m.deleteRememberCookie(w, r)
						m.App.Session.Put(r.Context(), "error", "you,ve been logged out already !")
						next.ServeHTTP(w, r)
					} else {
						// valid, so log the user in
						user, _ := u.GetById(uid)
						m.App.Session.Put(r.Context(), "userID", user.ID)
						m.App.Session.Put(r.Context(), "remember_token", hash)
						next.ServeHTTP(w, r)
					}
				} else {
					m.deleteRememberCookie(w, r)
					next.ServeHTTP(w, r)
				}
			}

		} else {
			// user is logged in
			next.ServeHTTP(w, r)
		}
	})
}

func (m *Middleware) deleteRememberCookie(w http.ResponseWriter, r *http.Request) {
	// Best practice renewing the session
	_ = m.App.Session.RenewToken(r.Context())
	// delete cookie
	ck := http.Cookie{
		Name:     fmt.Sprintf("_%s_remember", m.App.AppName),
		Value:    "",
		Path:     "/",
		Domain:   m.App.Session.Cookie.Domain,
		Expires:  time.Now().Add(-100 * time.Hour),
		MaxAge:   -1,
		Secure:   m.App.Session.Cookie.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &ck)
	// log the user out
	m.App.Session.Remove(r.Context(), "userID")
	m.App.Session.Destroy(r.Context())
	// always good to renew token
	_ = m.App.Session.RenewToken(r.Context())
}
