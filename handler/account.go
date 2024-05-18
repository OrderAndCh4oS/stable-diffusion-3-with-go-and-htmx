package handler

import (
	"fmt"
	"net/http"
	"token-based-payment-service-api/db"
	"token-based-payment-service-api/pkg/kit/validate"
	"token-based-payment-service-api/pkg/sb"
	"token-based-payment-service-api/view/account"
)

func HandleAccountIndex(w http.ResponseWriter, r *http.Request) error {
	user := GetAuthenticatedUser(r)
	return Render(w, r, account.Index(user))
}

func HandleAccountUpdate(w http.ResponseWriter, r *http.Request) error {
	params := account.AccountUpdateParams{
		Username: r.FormValue("username"),
	}
	errors := account.AccountUpdateErrors{}
	ok := validate.New(&params, validate.Fields{
		"Username": validate.Rules(validate.MinLength(3), validate.MaxLength(50)),
	}).Validate(&errors)
	if !ok {
		return Render(w, r, account.AccountUpdateForm(params, errors))
	}
	user := GetAuthenticatedUser(r)
	user.Account.Username = params.Username
	if err := db.UpdateAccount(&user.Account); err != nil {
		fmt.Println(err)
		return Render(w, r, account.AccountUpdateForm(params, account.AccountUpdateErrors{
			ServerError: "Failed to update account data, due to a server issue. Please try later",
		}))
	}
	params.Success = "Account updated successfully"
	return Render(w, r, account.AccountUpdateForm(params, errors))
}

func HandleAccountChangePassword(w http.ResponseWriter, r *http.Request) error {
	params := account.ChangePasswordParams{
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirmPassword"),
	}
	errors := account.ChangePasswordErrors{}
	ok := validate.New(&params, validate.Fields{
		"Password":        validate.Rules(validate.Password),
		"ConfirmPassword": validate.Rules(validate.Equal(params.Password), validate.Message("Passwords do not match")),
	}).Validate(&errors)
	if !ok {
		return Render(w, r, account.ChangePasswordForm("", errors))
	}
	user := GetAuthenticatedUser(r)
	_, err := sb.Client.Auth.UpdateUser(r.Context(), user.AccessToken, map[string]interface{}{"password": params.Password})
	if err != nil {
		fmt.Println(err)
		return Render(w, r, account.ChangePasswordForm("", account.ChangePasswordErrors{
			ServerError: "Failed to update account data, due to a server issue. Please try later",
		}))
	}
	return Render(w, r, account.ChangePasswordForm("Password updated successfully", account.ChangePasswordErrors{}))
}
