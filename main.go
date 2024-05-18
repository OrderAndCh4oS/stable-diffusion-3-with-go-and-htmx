package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"log"
	"log/slog"
	"net/http"
	"os"
	"token-based-payment-service-api/db"
	. "token-based-payment-service-api/handler"
	"token-based-payment-service-api/pkg/sb"
)

func main() {
	if err := initAll(); err != nil {
		log.Fatal(err)
	}

	router := chi.NewMux()
	router.Use(WithUser)
	router.Handle("/*", http.StripPrefix("/public/", http.FileServerFS(os.DirFS("public"))))
	router.Get("/", Make(HandleHomeIndex))
	router.Get("/sign-up", Make(HandleSignUpIndex))
	router.Post("/sign-up", Make(HandleSignUpCreate))
	router.Get("/sign-in", Make(HandleSignInIndex))
	router.Get("/sign-in/provider/google", Make(HandleSignInWithGoogle))
	router.Post("/sign-in", Make(HandleSignInCreate))
	router.Post("/sign-out", Make(HandleSignOutCreate))
	router.Get("/forgot-password", Make(HandleForgotPasswordIndex))
	router.Post("/forgot-password", Make(HandleForgotPasswordCreate))
	router.Get("/auth/callback", Make(HandleAuthCallback))

	router.Group(func(authedRouter chi.Router) {
		authedRouter.Use(WithAuth)
		authedRouter.Get("/account/setup", Make(HandleAccountSetupIndex))
		authedRouter.Post("/account/setup", Make(HandleAccountSetupCreate))
	})

	router.Group(func(authedRouter chi.Router) {
		authedRouter.Use(WithAuth, WithAccountSetup)
		authedRouter.Get("/dashboard", Make(HandleDashboardIndex))
		authedRouter.Get("/account", Make(HandleAccountIndex))
		authedRouter.Put("/account/update", Make(HandleAccountUpdate))
		authedRouter.Put("/account/change-password", Make(HandleAccountChangePassword))
		authedRouter.Get("/reset-password", Make(HandleResetPasswordIndex))
		authedRouter.Put("/reset-password", Make(HandleResetPasswordCreate))
		authedRouter.Get("/generate", Make(HandleGenerateIndex))
		authedRouter.Post("/generate", Make(HandleGenerateCreate))
		authedRouter.Get("/image/status/{id}", Make(HandleImageStatus))
	})

	port := os.Getenv("HTTP_LISTEN_PORT")
	slog.Info("application running", "port", port)
	log.Fatal(http.ListenAndServe(port, router))
}

func initAll() error {
	if err := godotenv.Load(); err != nil {
		return err
	}
	if err := db.Init(); err != nil {
		return err
	}
	return sb.InitSb()
}
