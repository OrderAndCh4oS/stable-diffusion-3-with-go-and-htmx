package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/nedpals/supabase-go"
	"io"
	"net/http"
	"os"
	"token-based-payment-service-api/db"
	"token-based-payment-service-api/pkg/kit/validate"
	"token-based-payment-service-api/pkg/sb"
	"token-based-payment-service-api/types"
	"token-based-payment-service-api/view/auth"
)

const (
	sessionUserKey        = "user"
	sessionAccessTokenKey = "accessToken"
)

func HandleForgotPasswordIndex(w http.ResponseWriter, r *http.Request) error {
	return Render(w, r, auth.ForgotPassword())
}

func resetPasswordForEmailRequest(email string) error {
	reqBody, err := json.Marshal(map[string]string{
		"email":      email,
		"redirectTo": fmt.Sprintf("%s%s", os.Getenv("APP_URL"), "/auth/reset-password"),
	})
	if err != nil {
		return err
	}
	reqURL := fmt.Sprintf("%s/auth/v1/recover", os.Getenv("SUPABASE_URL"))
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("apiKey", os.Getenv("SUPABASE_SECRET"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// Todo: remove this
	//fmt.Println(string(body))

	return nil
}

func HandleForgotPasswordCreate(w http.ResponseWriter, r *http.Request) error {
	params := auth.ForgotPasswordParams{
		Email: r.FormValue("email"),
	}
	errors := auth.ForgotPasswordErrors{}
	ok := validate.New(&params, validate.Fields{
		"Email": validate.Rules(validate.Email),
	}).Validate(&errors)
	if !ok {
		return Render(w, r, auth.ForgotPasswordForm(params, errors))
	}
	if err := resetPasswordForEmailRequest(params.Email); err != nil {
		return Render(w, r, auth.ForgotPasswordForm(params, auth.ForgotPasswordErrors{
			ServerError: "Failed to save account data, due to a server issue. Please try later",
		}))
	}

	return Render(w, r, auth.ForgotPasswordSuccess())
}

func HandleAccountSetupIndex(w http.ResponseWriter, r *http.Request) error {
	return Render(w, r, auth.AccountSetup())
}

func HandleAccountSetupCreate(w http.ResponseWriter, r *http.Request) error {
	params := auth.AccountSetupParams{
		Username: r.FormValue("username"),
	}
	errors := auth.AccountSetupErrors{}
	ok := validate.New(&params, validate.Fields{
		"Username": validate.Rules(validate.MinLength(3), validate.MaxLength(50)),
	}).Validate(&errors)
	if !ok {
		return Render(w, r, auth.AccountSetupForm(params, errors))
	}
	user := GetAuthenticatedUser(r)
	accountData := types.Account{
		UserID:   user.ID,
		Username: params.Username,
	}
	if err := db.CreateAccount(&accountData); err != nil {
		fmt.Println(err)
		return Render(w, r, auth.AccountSetupForm(params, auth.AccountSetupErrors{
			ServerError: "Failed to save account data, due to a server issue. Please try later",
		}))
	}
	HxRedirect(w, r, "/dashboard")
	return nil
}

func HandleSignInIndex(w http.ResponseWriter, r *http.Request) error {
	return Render(w, r, auth.SignIn())
}

func HandleSignInWithGoogle(w http.ResponseWriter, r *http.Request) error {
	resp, err := sb.Client.Auth.SignInWithProvider(supabase.ProviderSignInOptions{
		Provider:   "google",
		RedirectTo: "http://localhost:3000/auth/callback",
	})
	if err != nil {
		return err
	}
	http.Redirect(w, r, resp.URL, http.StatusSeeOther)
	return nil
}

func HandleSignInCreate(w http.ResponseWriter, r *http.Request) error {
	credentials := supabase.UserCredentials{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	resp, err := sb.Client.Auth.SignIn(r.Context(), credentials)

	if err != nil {
		return Render(w, r, auth.SignInForm(credentials, auth.SignInErrors{
			InvalidCredentials: "The credentials you provided were invalid",
		}))
	}
	if err = setAuthSession(w, r, resp.AccessToken); err != nil {
		return err
	}
	to := r.URL.Query().Get("to")
	if len(to) == 0 {
		HxRedirect(w, r, "/dashboard")
		return nil
	}
	HxRedirect(w, r, to)
	return nil
}

func HandleSignUpIndex(w http.ResponseWriter, r *http.Request) error {
	return Render(w, r, auth.SignUp())
}

func HandleSignUpCreate(w http.ResponseWriter, r *http.Request) error {
	params := auth.SignUpParams{
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirmPassword"),
	}

	signUpErrors := auth.SignUpErrors{}

	if ok := validate.New(&params, validate.Fields{
		"Email":           validate.Rules(validate.Email),
		"Password":        validate.Rules(validate.Password),
		"ConfirmPassword": validate.Rules(validate.Equal(params.Password), validate.Message("Passwords do not match")),
	}).Validate(&signUpErrors); !ok {
		return Render(w, r, auth.SignUpForm(params, signUpErrors))
	}

	credentials := supabase.UserCredentials{
		Email:    params.Email,
		Password: params.Password,
	}

	user, err := sb.Client.Auth.SignUp(r.Context(), credentials)

	if err != nil {
		return Render(w, r, auth.SignUpForm(params, auth.SignUpErrors{
			SignUpError: "Error creating account",
		}))
	}

	return Render(w, r, auth.SignUpSuccess(user.Email))
}

func HandleResetPasswordIndex(w http.ResponseWriter, r *http.Request) error {
	return Render(w, r, auth.ResetPasswordIndex())
}

func HandleResetPasswordCreate(w http.ResponseWriter, r *http.Request) error {
	params := auth.ResetPasswordParams{
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirmPassword"),
	}
	errors := auth.ResetPasswordErrors{}
	ok := validate.New(&params, validate.Fields{
		"Password":        validate.Rules(validate.Password),
		"ConfirmPassword": validate.Rules(validate.Equal(params.Password), validate.Message("Passwords do not match")),
	}).Validate(&errors)
	if !ok {
		return Render(w, r, auth.ResetPasswordForm(errors))
	}
	user := GetAuthenticatedUser(r)
	_, err := sb.Client.Auth.UpdateUser(r.Context(), user.AccessToken, map[string]interface{}{"password": params.Password})
	if err != nil {
		fmt.Println(err)
		return Render(w, r, auth.ResetPasswordForm(auth.ResetPasswordErrors{
			ServerError: "Failed to update account data, due to a server issue. Please try later",
		}))
	}
	HxRedirect(w, r, "/dashboard")
	return nil
}

// HandleAuthCallback Todo: handle errors here: eg ?error=access_denied&error_code=403&error_description=Email+link+is+invalid+or+has+expired
func HandleAuthCallback(w http.ResponseWriter, r *http.Request) error {
	accessToken := r.URL.Query().Get("access_token")
	callbackType := r.URL.Query().Get("type")
	// Todo: This auth callback needs to be replaced with authorisation code flow
	if len(accessToken) == 0 {
		return Render(w, r, auth.AuthCallbackScript())
	}
	if err := setAuthSession(w, r, accessToken); err != nil {
		return err
	}
	if callbackType == "recovery" {
		http.Redirect(w, r, "/reset-password", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
	return nil
}

func HandleSignOutCreate(w http.ResponseWriter, r *http.Request) error {
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	session, _ := store.Get(r, sessionUserKey)
	session.Values[sessionAccessTokenKey] = ""
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		return err
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func setAuthSession(w http.ResponseWriter, r *http.Request, accessToken string) error {
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	session, _ := store.Get(r, sessionUserKey)
	session.Values[sessionAccessTokenKey] = accessToken
	return session.Save(r, w)
}
