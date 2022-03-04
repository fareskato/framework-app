package handlers

import (
	"context"
	"net/http"

	"github.com/fareskato/kabarda"
)

/*
	Helpers function to be used in handlers: for example to simplify h.App.Render.Page   ..etc
*/

// render render jet/golang template
func (h *Handlers) render(w http.ResponseWriter, r *http.Request, view string, variables, data interface{}) error {
	return h.App.Render.Page(w, r, view, variables, data)
}

// sessionPut puts data in the session
func (h *Handlers) sessionPut(ctx context.Context, k string, val interface{}) {
	h.App.Session.Put(ctx, k, val)
}

// sessionExists checks if key exists in the session
func (h *Handlers) sessionExists(ctx context.Context, k string) bool {
	return h.App.Session.Exists(ctx, k)
}

// sessionGet get value from the session
func (h *Handlers) sessionGet(ctx context.Context, k string) interface{} {
	return h.App.Session.Get(ctx, k)
}

// sessionRemove remove the session
func (h *Handlers) sessionRemove(ctx context.Context, k string) {
	h.App.Session.Remove(ctx, k)
}

// sessionRenew renew token uses with login/logout
func (h *Handlers) sessionRenew(ctx context.Context) error {
	return h.App.Session.RenewToken(ctx)
}

// sessionDestroy remove the session and destroy everything in it
func (h *Handlers) sessionDestroy(ctx context.Context) error {
	return h.App.Session.Destroy(ctx)
}

func (h *Handlers) randomString(n int) string {
	return h.App.RandomString(n)
}

func (h *Handlers) encrypt(txt string) (string, error) {
	enc := kabarda.Encryption{Key: []byte(h.App.EncryptionKey)}
	encrypted, err := enc.Encrypt(txt)
	if err != nil {
		return "", err
	}
	return encrypted, nil
}

func (h *Handlers) decrypt(txt string) (string, error) {
	enc := kabarda.Encryption{Key: []byte(h.App.EncryptionKey)}
	decrypted, err := enc.Decrypt(txt)
	if err != nil {
		return "", err
	}
	return decrypted, nil
}
