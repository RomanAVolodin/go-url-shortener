package middlewares

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

const CookieName = "user-id"

func AuthCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userId uuid.UUID
		var err error
		userId, err = GetUserIDFromCookie(r)

		if err != nil || userId == uuid.Nil {
			userId, err = uuid.NewUUID()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}

		cookie := &http.Cookie{
			Name:     CookieName,
			Value:    GenerateCookieStringForUserId(userId),
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			Path:     "/",
		}
		r.AddCookie(cookie)
		http.SetCookie(w, cookie)
		next.ServeHTTP(w, r)
	})
}

func MakeSignature(itemToEncode string) string {
	h := hmac.New(sha256.New, []byte(config.Settings.SecretAuthKey))
	h.Write([]byte(itemToEncode))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func GenerateCookieStringForUserId(userId uuid.UUID) string {
	return hex.EncodeToString([]byte(userId.String())) + "|" + MakeSignature(userId.String())
}

func GetUserIDFromCookie(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie(CookieName)
	if cookie == nil {
		return uuid.Nil, err
	}
	cookieArray := strings.Split(cookie.Value, "|")
	if len(cookieArray) == 2 {
		decodedId, errDecode := hex.DecodeString(cookieArray[0])
		userIdString, signString := decodedId, cookieArray[1]
		userId, errParse := uuid.ParseBytes(userIdString)
		newSign := MakeSignature(userId.String())

		if hmac.Equal([]byte(newSign), []byte(signString)) && errParse == nil && errDecode == nil {
			return userId, nil
		}
	}

	return uuid.Nil, err
}
