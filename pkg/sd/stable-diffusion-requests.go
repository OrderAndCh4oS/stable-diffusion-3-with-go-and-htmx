package sd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type RequestBody struct {
	Prompt         string
	AspectRatio    string
	Mode           string
	NegativePrompt string
	Model          string
	Seed           int
	OutputFormat   string
	StylePreset    string
}

type ResponseBody struct {
	Image string `json:"image"`
}

func TextToImageRequest(prompt string) ([]byte, error) {
	apiKey := os.Getenv("STABILITY_AI_API_KEY")
	apiEndpoint := "https://api.stability.ai/v2beta/stable-image/generate/sd3"

	requestBody := RequestBody{
		Prompt:       prompt,
		AspectRatio:  "1:1",
		Mode:         "text-to-image",
		Model:        "sd3", // Note use "sd3" or "sd3-turbo"
		OutputFormat: "jpeg",
		Seed:         0, // Note: use 0 for random
	}

	var requestBodyBuffer bytes.Buffer
	writer := multipart.NewWriter(&requestBodyBuffer)

	if err := writer.WriteField("prompt", requestBody.Prompt); err != nil {
		return nil, err
	}
	if err := writer.WriteField("aspect_ratio", requestBody.AspectRatio); err != nil {
		return nil, err
	}
	if err := writer.WriteField("mode", requestBody.Mode); err != nil {
		return nil, err
	}
	if err := writer.WriteField("negative_prompt", requestBody.NegativePrompt); err != nil {
		return nil, err
	}
	if err := writer.WriteField("model", requestBody.Model); err != nil {
		return nil, err
	}
	if err := writer.WriteField("seed", fmt.Sprintf("%d", requestBody.Seed)); err != nil {
		return nil, err
	}
	if err := writer.WriteField("output_format", requestBody.OutputFormat); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiEndpoint, &requestBodyBuffer)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("non-success status code: %d", resp.StatusCode)
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response ResponseBody
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, err
	}

	imageData, err := base64.StdEncoding.DecodeString(response.Image)
	if err != nil {
		return nil, err
	}

	return imageData, nil
}
