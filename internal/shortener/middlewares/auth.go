// Package middlewares includes all middlewares of application.
package middlewares

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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
			userID = uuid.New()
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
	return DecodeUserIDFromHashedString(cookie.Value)
}

// DecodeUserIDFromHashedString gets user id from hashed string.
func DecodeUserIDFromHashedString(hashedString string) (uuid.UUID, error) {
	arrayFromHashedString := strings.Split(hashedString, "|")
	if len(arrayFromHashedString) == 2 {
		decodedID, errDecode := hex.DecodeString(arrayFromHashedString[0])
		userIDString, signString := decodedID, arrayFromHashedString[1]
		userID, errParse := uuid.ParseBytes(userIDString)
		newSign := MakeSignature(userID.String())

		if hmac.Equal([]byte(newSign), []byte(signString)) && errParse == nil && errDecode == nil {
			return userID, nil
		}
	}

	return uuid.Nil, errors.New("failed to parse hashed string")
}
