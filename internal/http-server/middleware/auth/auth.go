package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/arxonic/journal/internal/domain/models"
	"github.com/arxonic/journal/internal/storage/sqlite"
	"github.com/golang-jwt/jwt/v5"
)

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

		// куков нет - выдать
		// куки есть
		// куки есть, но плохие - выдать
		encryptedRole := getRoleFromCookie(r)
		if encryptedRole == "" {
			err = setRoleToCookie(w, email, m)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		} else {
			err = checkRoleFromCookie(encryptedRole, m.Secret)
			if err != nil {
				err = setRoleToCookie(w, email, m)
				if err != nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func checkRoleFromCookie(encryptedRole, secret string) error {
	encDecoded, err := base64.StdEncoding.DecodeString(encryptedRole)
	if err != nil {
		return err
	}

	dec, err := decrypt([]byte(encDecoded), []byte(secret))
	if err != nil {
		return err
	}

	var key models.Key

	err = json.Unmarshal(dec, &key)
	if err != nil {
		return err
	}

	return nil
}

func setRoleToCookie(w http.ResponseWriter, email string, m *AuthMiddleware) error {
	key, err := m.Storage.UserRole(email)
	if err != nil {
		return err
	}

	value, err := json.Marshal(key)
	if err != nil {
		return err
	}

	enc, err := encrypt(value, []byte(m.Secret))
	if err != nil {
		return err
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

	return nil
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
