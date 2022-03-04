package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/fareskato/kabarda/mailer"
	"github.com/fareskato/kabarda/signer"
	"myapp/data"
	"net/http"
	"os"
	"time"
)

const RememberToken = "remember_token"

func (h *Handlers) UserRegister(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)
	validator := h.App.Validator(nil)
	vars.Set("validator", validator)
	vars.Set("user", data.User{})
	err := h.render(w, r, "register", vars, nil)
	if err != nil {
		h.App.ErrorLog.Println(err)
	}
}

func (h *Handlers) UserRegisterPost(w http.ResponseWriter, r *http.Request) {
	// first we parse the form
	err := r.ParseForm()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	// init user
	var user data.User
	// parse data and password
	firstName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	// validation
	validator := h.App.Validator(nil)

	validator.Required(r, "first_name", "last_name", "email", "password")
	validator.IsEmail("email", email)

	validator.Check(len(r.Form.Get("first_name")) > 3, "first_name", "Must be at least 3 characters")
	validator.Check(len(r.Form.Get("last_name")) > 3, "last_name", "Must be at least 3 characters")
	validator.Check(len(r.Form.Get("password")) > 4, "password", "Must be at least 4 characters")

	if !validator.Valid() {
		vars := make(jet.VarMap)
		vars.Set("validator", validator)
		user.FirstName = firstName
		user.LastName = lastName
		user.Email = email
		user.Password = password
		vars.Set("user", user)
		if err := h.App.Render.Page(w, r, "register", vars, nil); err != nil {
			h.App.ErrorLog.Println(err)
			return
		}
		return
	}
	// validation passed
	// if this is the first user make is supper user
	users, err := h.Models.Users.GetAll()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email
	user.Password = password
	// make it supper user
	if len(users) == 0 {
		user.SuperUser = 1
		user.Active = 1
	}
	_, err = h.Models.Users.Insert(user)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	// redirect
	http.Redirect(w, r, fmt.Sprintf("%s", os.Getenv("USER_LOGIN")), http.StatusSeeOther)

}

// UserLogin display user login form
func (h *Handlers) UserLogin(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)
	validator := h.App.Validator(nil)
	vars.Set("validator", validator)
	vars.Set("user", data.User{})
	err := h.render(w, r, "login", vars, nil)
	if err != nil {
		h.App.ErrorLog.Println(err)
	}
}

// UserLoginPost handle user login
func (h *Handlers) UserLoginPost(w http.ResponseWriter, r *http.Request) {
	// first we parse the form
	err := r.ParseForm()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	// parse data and password
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	// validation
	validator := h.App.Validator(nil)

	validator.Required(r, "email", "password")
	validator.IsEmail("email", email)
	validator.Check(len(r.Form.Get("password")) > 4, "password", "Must be at least 4 characters")

	if !validator.Valid() {
		vars := make(jet.VarMap)
		vars.Set("validator", validator)
		var user data.User
		user.Email = email
		user.Password = email
		vars.Set("user", user)
		if err := h.App.Render.Page(w, r, "login", vars, nil); err != nil {
			h.App.ErrorLog.Println(err)
			return
		}
		return
	}
	// fetch user by email
	user, err := h.Models.Users.GetByEmail(email)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	// now we have the user, so we can check password matches
	missMatches, err := user.PasswordMatches(password)
	if err != nil {
		w.Write([]byte("Validating password error"))
		return
	}
	if !missMatches {
		w.Write([]byte("Password matches error"))
		return
	}
	// remember me checked?
	if r.Form.Get("remember") == "remember" {
		// generate token(random text) and save in the cookie, so the user can use to log in
		rdText := h.randomString(12)
		hasher := sha256.New()
		_, err := hasher.Write([]byte(rdText))
		if err != nil {
			h.App.ErrorStatus(w, http.StatusBadRequest)
			return
		}
		// all good, so save the token in DB
		sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
		rm := data.RememberToken{}
		err = rm.Insert(user.ID, sha)
		if err != nil {
			h.App.ErrorStatus(w, http.StatusBadRequest)
			return
		}
		// then set and save cookie
		exp := time.Now().Add(365 * 24 * 60 * 60 * time.Second)
		cookie := http.Cookie{
			Name:     fmt.Sprintf("_%s_remember", h.App.AppName),
			Value:    fmt.Sprintf("%d|%s", user.ID, sha),
			Path:     "/",
			Domain:   h.App.Session.Cookie.Domain,
			Expires:  exp,
			MaxAge:   315_350_000,
			Secure:   h.App.Session.Cookie.Secure,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, &cookie)
		// save hash in session
		h.App.Session.Put(r.Context(), RememberToken, sha)
	}

	// put user id in the session
	h.App.Session.Put(r.Context(), "userID", user.ID)
	// redirect
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (h *Handlers) UserLogout(w http.ResponseWriter, r *http.Request) {
	// delete remember token if exists
	if h.App.Session.Exists(r.Context(), RememberToken) {
		rt := data.RememberToken{}
		_ = rt.Delete(h.App.Session.GetString(r.Context(), RememberToken))
	}
	// delete remember cookie
	otherCookie := http.Cookie{
		Name:     fmt.Sprintf("_%s_remember", h.App.AppName),
		Value:    "",
		Path:     "/",
		Domain:   h.App.Session.Cookie.Domain,
		Expires:  time.Now().Add(-100 * time.Hour),
		MaxAge:   -1,
		Secure:   h.App.Session.Cookie.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &otherCookie)
	// Always renew token(Important)
	h.App.Session.RenewToken(r.Context())
	// remove userId from session
	h.App.Session.Remove(r.Context(), "userID")
	// remove remember token
	h.App.Session.Remove(r.Context(), RememberToken)
	// destroy the session
	h.App.Session.Destroy(r.Context())
	//  renew token again
	h.App.Session.RenewToken(r.Context())
	// redirect
	http.Redirect(w, r, fmt.Sprintf("%s", os.Getenv("USER_LOGIN")), http.StatusSeeOther)
}

// ForgotPassword displays forgot password page
func (h *Handlers) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)
	validator := h.App.Validator(nil)
	vars.Set("validator", validator)
	vars.Set("user", data.User{})
	err := h.render(w, r, "forgot", vars, nil)
	if err != nil {
		h.App.ErrorLog.Println(err)
		h.App.Error500(w, r)
	}
}

// ForgotPasswordPost handle the forgot password functionality
func (h *Handlers) ForgotPasswordPost(w http.ResponseWriter, r *http.Request) {
	// parse form(always)
	err := r.ParseForm()
	if err != nil {
		h.App.ErrorStatus(w, http.StatusBadRequest)
		return
	}
	email := r.Form.Get("email")
	// init user
	var user data.User

	// validation
	validator := h.App.Validator(nil)
	validator.Required(r, "email")
	validator.IsEmail("email", email)
	if !validator.Valid() {
		vars := make(jet.VarMap)
		vars.Set("validator", validator)
		user.Email = email
		vars.Set("user", user)
		if err := h.App.Render.Page(w, r, "forgot", vars, nil); err != nil {
			h.App.ErrorLog.Println(err)
			return
		}
		return
	}

	// check if there is a user has the email
	var u *data.User
	u, err = u.GetByEmail(email)
	if err != nil {
		h.App.ErrorStatus(w, http.StatusBadRequest)
		return
	}
	// create a link to password reset from
	link := fmt.Sprintf("%s/users/reset-password?email=%s", h.App.Server.URL, email)
	// sign the link: send email with link specific to this user via token in the link
	sign := signer.Signer{
		Secret: []byte(h.App.EncryptionKey),
	}
	signedLink := sign.GenerateTokenFromString(link)

	// email the message
	var data struct {
		Link string
	}
	data.Link = signedLink
	// email address
	senderEmail := os.Getenv("SMTP_FROM")
	fmt.Println(senderEmail)
	msg := mailer.Message{
		From:        senderEmail,
		To:          u.Email,
		Subject:     "Reset password",
		Template:    "password-reset",
		Attachments: nil,
		Data:        data,
	}
	h.App.Mail.Jobs <- msg
	res := <-h.App.Mail.Result
	if res.Error != nil {
		h.App.ErrorStatus(w, http.StatusBadRequest)
		return
	}
	// redirect user
	// redirect with flash message
	h.App.Session.Put(r.Context(), "flash", "password reset email sent")
	http.Redirect(w, r, fmt.Sprintf("%s", os.Getenv("USER_LOGIN")), http.StatusSeeOther)
}

// ResetPasswordForm display reset password form
func (h *Handlers) ResetPasswordForm(w http.ResponseWriter, r *http.Request) {
	// get form values
	email := r.URL.Query().Get("email")
	theURL := r.RequestURI
	testURL := fmt.Sprintf("%s%s", h.App.Server.URL, theURL)

	// validate the url
	signer := signer.Signer{
		Secret: []byte(h.App.EncryptionKey),
	}

	valid := signer.VerifyToken(testURL)
	if !valid {
		h.App.ErrorLog.Print("Invalid url")
		h.App.ErrorUnauthorized(w, r)
		return
	}

	/// make sure it's not expired
	expired := signer.Expired(testURL, 60)
	if expired {
		h.App.ErrorLog.Print("Link expired")
		h.App.ErrorUnauthorized(w, r)
		return
	}

	// display form
	encryptedEmail, _ := h.encrypt(email)
	validator := h.App.Validator(nil)
	vars := make(jet.VarMap)
	vars.Set("email", encryptedEmail)
	vars.Set("validator", validator)

	err := h.render(w, r, "reset-password", vars, nil)
	if err != nil {
		return
	}
}

// ResetPasswordFormPost handle reset password functionality
func (h *Handlers) ResetPasswordFormPost(w http.ResponseWriter, r *http.Request) {
	// parse the form
	err := r.ParseForm()
	if err != nil {
		h.App.Error500(w, r)
		return
	}

	// get and decrypt the email
	email, err := h.decrypt(r.Form.Get("email"))
	if err != nil {
		h.App.Error500(w, r)
		return
	}
	// validation
	password := r.Form.Get("password")
	verifyPassword := r.Form.Get("verify-password")
	validator := h.App.Validator(nil)
	validator.Required(r, "password", "verify-password")
	validator.Check(len(password) > 4, "password", "Must be at least 4 characters")
	validator.Check(len(verifyPassword) > 4, "verify-password", "Must be at least 4 characters")
	if !validator.Valid() {
		vars := make(jet.VarMap)
		vars.Set("validator", validator)
		vars.Set("password", password)
		vars.Set("email", email)
		if err := h.App.Render.Page(w, r, "reset-password", vars, nil); err != nil {
			h.App.ErrorLog.Println(err)
			return
		}
		return
	}
	// get the user
	var u data.User
	user, err := u.GetByEmail(email)
	if err != nil {
		h.App.Error500(w, r)
		return
	}
	// reset the password
	err = user.ResetPassword(user.ID, r.Form.Get("password"))
	if err != nil {
		h.App.Error500(w, r)
		return
	}

	// redirect with flash message
	h.App.Session.Put(r.Context(), "flash", "password reset, you can log in.")
	http.Redirect(w, r, fmt.Sprintf("%s", os.Getenv("USER_LOGIN")), http.StatusSeeOther)
}
