// Package middlewares includes all middlewares of application.
package middlewares

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/google/uuid"
)

// CookieName string name for cookie.
const CookieName = "user-id"

// UserKey represents UserKey.
type UserKey string

// UserIDKey constant for UserIdKey.
const UserIDKey UserKey = "id"

// AuthCookie middleware coops with user id in cookie.
func AuthCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID uuid.UUID
		var err error
		userID, err = GetUserIDFromCookie(r)

		if err != nil || userID == uuid.Nil {
			userID, err = uuid.NewUUID()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}

		cookie := &http.Cookie{
			Name:     CookieName,
			Value:    GenerateCookieStringForUserID(userID),
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			Path:     "/",
		}
		r.AddCookie(cookie)
		http.SetCookie(w, cookie)

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// MakeSignature makes signature.
func MakeSignature(itemToEncode string) string {
	h := hmac.New(sha256.New, []byte(config.Settings.SecretAuthKey))
	h.Write([]byte(itemToEncode))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// GenerateCookieStringForUserID generates cookie string for user id.
func GenerateCookieStringForUserID(userID uuid.UUID) string {
	return hex.EncodeToString([]byte(userID.String())) + "|" + MakeSignature(userID.String())
}

// GetUserIDFromCookie gets user id from cookie.
func GetUserIDFromCookie(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie(CookieName)
	if cookie == nil {
		return uuid.Nil, err
	}
	cookieArray := strings.Split(cookie.Value, "|")
	if len(cookieArray) == 2 {
		decodedID, errDecode := hex.DecodeString(cookieArray[0])
		userIDString, signString := decodedID, cookieArray[1]
		userID, errParse := uuid.ParseBytes(userIDString)
		newSign := MakeSignature(userID.String())

		if hmac.Equal([]byte(newSign), []byte(signString)) && errParse == nil && errDecode == nil {
			return userID, nil
		}
	}

	return uuid.Nil, err
}
