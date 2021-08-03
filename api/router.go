package api

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
)

type ErrorHandlingWrapper func(w http.ResponseWriter, r *http.Request) error

func (h ErrorHandlingWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		handleError(err, w, r)
	}
}

func (a *API) AuthHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("apikey"); if tokenString == "" {
			sendJSON(w, http.StatusUnauthorized, "")
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte(a.config.JwtSecret), nil
		}); if err != nil {
			sendJSON(w, http.StatusUnauthorized, err)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			role, ok := claims["role"]; if ok && role == "supabase_admin" {
				// successful authentication
				next.ServeHTTP(w, r)
				return
			} else {
				sendJSON(w, http.StatusForbidden, "this token does not have a valid claim over the correct role")
				return
			}
		} else {
			sendJSON(w, http.StatusForbidden, err)
			return
		}
	}

	return http.HandlerFunc(fn)
}
