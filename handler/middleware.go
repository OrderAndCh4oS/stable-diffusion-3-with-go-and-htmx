package handler

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"net/http"
	"os"
	"strings"
	"token-based-payment-service-api/db"
	"token-based-payment-service-api/pkg/sb"
	"token-based-payment-service-api/types"
)

func WithAuth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "public") {
			next.ServeHTTP(w, r)
			return
		}
		user := GetAuthenticatedUser(r)
		if !user.LoggedIn {
			path := r.URL.Path
			HxRedirect(w, r, "/sign-in?to="+path)
		}
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func WithAccountSetup(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := GetAuthenticatedUser(r)
		account, err := db.GetAccountByUserId(user.ID)
		if errors.Is(err, sql.ErrNoRows) {
			http.Redirect(w, r, "/account/setup", http.StatusSeeOther)
			return
		}
		if err != nil {
			next.ServeHTTP(w, r)
		}
		user.Account = account
		ctx := context.WithValue(r.Context(), types.UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func WithUser(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "public") {
			next.ServeHTTP(w, r)
			return
		}
		store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
		session, err := store.Get(r, sessionUserKey)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		accessToken := session.Values[sessionAccessTokenKey]
		if err != nil || accessToken == nil {
			next.ServeHTTP(w, r)
			return
		}
		resp, err := sb.Client.Auth.User(r.Context(), accessToken.(string))
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		user := types.AuthenticatedUser{
			ID:          uuid.MustParse(resp.ID),
			Email:       resp.Email,
			LoggedIn:    true,
			AccessToken: accessToken.(string),
		}

		ctx := context.WithValue(r.Context(), types.UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}
