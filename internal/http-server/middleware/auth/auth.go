package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/arxonic/journal/internal/domain/models"
	"github.com/arxonic/journal/internal/storage/sqlite"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrOldCookie = errors.New("cookie is old")
)

type ContextKey string

const ContextAuthMiddlewareKey ContextKey = "authMiddleware"

type AuthMiddleware struct {
	Secret  string
	Storage *sqlite.Storage
}

func New(secret string, storage *sqlite.Storage) *AuthMiddleware {
	return &AuthMiddleware{
		Secret:  secret,
		Storage: storage,
	}
}

func (m *AuthMiddleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get jwt string from header
		jwtString := getJWTFromHeader(r)
		if jwtString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// get jwt from jwt string
		token, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.Secret), nil
		})
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		email := ""
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			email = claims["email"].(string)
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}

		var key models.Key

		encryptedRole := getRoleFromCookie(r)

		if encryptedRole == "" {
			key, err = setRoleToCookie(w, email, m)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		} else {
			key, err = checkRoleFromCookie(email, encryptedRole, m.Secret)
			if err != nil {
				key, err = setRoleToCookie(w, email, m)
				if err != nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}
		}

		ctx := context.WithValue(r.Context(), ContextAuthMiddlewareKey, &key)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func checkRoleFromCookie(email, encryptedRole, secret string) (models.Key, error) {
	var key models.Key
	encDecoded, err := base64.StdEncoding.DecodeString(encryptedRole)
	if err != nil {
		return key, err
	}

	dec, err := decrypt([]byte(encDecoded), []byte(secret))
	if err != nil {
		return key, err
	}

	err = json.Unmarshal(dec, &key)
	if err != nil {
		return key, err
	}

	if key.Email != email {
		return key, ErrOldCookie
	}

	return key, nil
}

func setRoleToCookie(w http.ResponseWriter, email string, m *AuthMiddleware) (models.Key, error) {
	key, err := m.Storage.UserRole(email)
	if err != nil {
		return models.Key{}, err
	}

	value, err := json.Marshal(key)
	if err != nil {
		return models.Key{}, err
	}

	enc, err := encrypt(value, []byte(m.Secret))
	if err != nil {
		return models.Key{}, err
	}

	encEncoded := base64.StdEncoding.EncodeToString(enc)

	cookie := http.Cookie{
		Name:     "role",
		Value:    encEncoded,
		Path:     "/",
		MaxAge:   9999,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)

	return key, nil
}

func getJWTFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	return ""
}

func getRoleFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("role")
	if err == nil {
		return cookie.Value
	}

	return ""
}

func encrypt(value, secret []byte) ([]byte, error) {
	c, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())

	return gcm.Seal(nonce, nonce, value, nil), nil
}

func decrypt(value, secret []byte) ([]byte, error) {
	c, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(value) < nonceSize {
		return nil, err
	}

	nonce, value := value[:nonceSize], value[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, value, nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}
