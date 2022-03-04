package kabarda

import (
	"github.com/asaskevich/govalidator"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Validation struct {
	Data   url.Values
	Errors map[string]string
}

func (k *Kabarda) Validator(data url.Values) *Validation {
	return &Validation{
		Data:   data,
		Errors: make(map[string]string),
	}
}

// Valid checks the validity
func (v *Validation) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds errors to validator
func (v *Validation) AddError(k, msg string) {
	if _, exists := v.Errors[k]; !exists {
		v.Errors[k] = msg
	}
}

// Has uses for form request and check if the form has a particular field
func (v *Validation) Has(field string, r *http.Request) bool {
	val := r.Form.Get(field)
	if val == "" {
		return false
	}
	return true
}

// Required uses for form post
func (v *Validation) Required(r *http.Request, fields ...string) {
	for _, f := range fields {
		val := r.Form.Get(f)
		if strings.TrimSpace(val) == "" {
			v.AddError(f, "This field is required")
		}
	}
}

func (v *Validation) Check(ok bool, k, msg string) {
	if !ok {
		v.AddError(k, msg)
	}
}

// IsEmail check if valid email
func (v *Validation) IsEmail(f, val string) {
	if !govalidator.IsEmail(val) {
		v.AddError(f, "Invalid email address")
	}
}

func (v *Validation) IsInt(f, val string) {
	_, err := strconv.Atoi(val)
	if err != nil {
		v.AddError(f, "This field must be an integer")
	}
}

func (v *Validation) IsFloat(f, val string) {
	_, err := strconv.ParseFloat(val, 64)
	if err != nil {
		v.AddError(f, "This field must be a decimal number")
	}
}

func (v *Validation) IsISODate(f, val string) {
	_, err := time.Parse("2006-01-02", val)
	if err != nil {
		v.AddError(f, "This field must be date in form YYYY-MM-DD")
	}
}

func (v *Validation) NoSpaces(f, val string) {
	if govalidator.HasWhitespace(val) {
		v.AddError(f, "Spaces are not permitted")
	}
}
