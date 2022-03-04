package kabarda

import (
	"github.com/justinas/nosurf"
	"net/http"
	"strconv"
)

////////////////////////////////
// Kabarda framework middlewares
////////////////////////////////

// SessionLoad loading and saving session on every request
func (k *Kabarda) SessionLoad(next http.Handler) http.Handler {
	// uses scs session manager to load the session
	return k.Session.LoadAndSave(next)
}

func (k *Kabarda) NoSurf(next http.Handler) http.Handler {
	// Constructs a new CSRFHandler that calls the specified handler if the CSRF check succeeds.
	csrfHandler := nosurf.New(next)
	// Sets the base cookie to use when building a CSRF token cookie This way you can
	//specify the Domain, Path, HttpOnly, Secure, etc.
	secure, _ := strconv.ParseBool(k.config.cookie.secure)

	// to Exclude some domains(for api) we can use
	csrfHandler.ExemptGlob("/api/*")

	csrfHandler.SetBaseCookie(http.Cookie{
		Name:     "",
		Path:     "/",
		Domain:   k.config.cookie.domain,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	return csrfHandler
}
