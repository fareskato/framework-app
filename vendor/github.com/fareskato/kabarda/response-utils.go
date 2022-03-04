package kabarda

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

func (kbr *Kabarda) ReadJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// max json data size: 1 mb
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// WriteJSON json response functionality
func (kbr *Kabarda) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	// marshal to json
	jsData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// check if there is any additional headers
	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	// set json headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// write to response
	_, err = w.Write(jsData)
	if err != nil {
		return err
	}
	return nil
}

// DownloadFile download files
func (kbr *Kabarda) DownloadFile(w http.ResponseWriter, r *http.Request, pathToFile, fileName string) error {
	fileFullPath := path.Join(pathToFile, fileName)
	// do cleaning
	fileToServe := filepath.Clean(fileFullPath)
	// set headers
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; file=\"%s\"", fileName))
	// serve the file
	http.ServeFile(w, r, fileToServe)
	return nil
}

func (kbr *Kabarda) Error404(w http.ResponseWriter, r *http.Request) {
	kbr.ErrorStatus(w, http.StatusNotFound)
}

func (kbr *Kabarda) Error500(w http.ResponseWriter, r *http.Request) {
	kbr.ErrorStatus(w, http.StatusInternalServerError)
}

func (kbr *Kabarda) ErrorUnauthorized(w http.ResponseWriter, r *http.Request) {
	kbr.ErrorStatus(w, http.StatusUnauthorized)
}

func (kbr *Kabarda) ErrorForbidden(w http.ResponseWriter, r *http.Request) {
	kbr.ErrorStatus(w, http.StatusForbidden)
}

func (kbr *Kabarda) ErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
