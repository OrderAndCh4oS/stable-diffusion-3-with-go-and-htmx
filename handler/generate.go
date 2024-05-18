package handler

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"net/http"
	"os"
	"strconv"
	"token-based-payment-service-api/db"
	"token-based-payment-service-api/pkg/kit/validate"
	"token-based-payment-service-api/pkg/sd"
	"token-based-payment-service-api/types"
	"token-based-payment-service-api/view/generate"
)

func HandleGenerateIndex(w http.ResponseWriter, r *http.Request) error {
	user := GetAuthenticatedUser(r)
	images, err := db.GetImagesByUserId(user.ID)
	if err != nil {
		return err
	}
	data := generate.ViewData{Images: images}
	return Render(w, r, generate.Index(data))
}

func HandleGenerateCreate(w http.ResponseWriter, r *http.Request) error {
	user := GetAuthenticatedUser(r)
	amount, err := strconv.Atoi(r.FormValue("amount"))

	params := generate.GenerateFormParams{
		Prompt: r.FormValue("prompt"),
	}

	errors := generate.GenerateFormErrors{}

	if err != nil {
		errors.Amount = "Not a valid amount"
		return Render(w, r, generate.GenerateForm(params, errors, generate.GenerateFormResult{}))
	}

	params.Amount = amount

	ok := validate.New(&params, validate.Fields{
		"Prompt": validate.Rules(validate.MinLength(3), validate.MaxLength(3000)),
		"Amount": validate.Rules(validate.Min(1), validate.Max(5)),
	}).Validate(&errors)

	if !ok {
		return Render(w, r, generate.GenerateForm(params, errors, generate.GenerateFormResult{}))
	}

	result := generate.GenerateFormResult{
		Images: []types.Image{},
	}
	if err = db.Bun.RunInTx(r.Context(), &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		for i := 0; i < params.Amount; i++ {
			image := types.Image{
				Prompt:  params.Prompt,
				UserID:  user.ID,
				Status:  types.ImageStatusPending,
				BatchID: uuid.New(),
			}

			// Todo: move this image generation to a separate service
			//       Create an event queue, send events to the queue to generate and save the image to s3 then update the db

			imageData, err := sd.TextToImageRequest(params.Prompt)
			if err != nil {
				fmt.Println("Error making HTTP request:", err)
				return err
			}
			filename := fmt.Sprintf("public/%s.jpeg", uuid.New())
			err = os.WriteFile(filename, imageData, 0644)
			if err != nil {
				fmt.Println("Error writing image to file:", err)
				return err
			}

			image.ImageLocation = filename
			image.Status = types.ImageStatusCompleted

			// End todo

			if err := db.CreateImage(&image); err != nil {
				return err
			}

			result.Images = append(result.Images, image)
		}
		return nil
	}); err != nil {
		errors.ServerError = "Failed to save images"
		return Render(w, r, generate.GenerateForm(params, errors, generate.GenerateFormResult{}))
	}

	return Render(w, r, generate.GenerateForm(params, errors, result))
}

func HandleImageStatus(w http.ResponseWriter, r *http.Request) error {
	user := GetAuthenticatedUser(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return err
	}
	image, err := db.GetImageById(user.ID, id)
	if err != nil {
		return err
	}
	return Render(w, r, generate.GalleryImage(image))
}
