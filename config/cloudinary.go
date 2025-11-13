package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type CloudinaryResponse struct {
	PublicID     string `json:"public_id"`
	SecureURL    string `json:"secure_url"`
	ResourceType string `json:"resource_type"`
}

// UploadToCloudinary mengunggah file ke Cloudinary
func UploadToCloudinary(file io.Reader, fileName, folder, resourceType string) (string, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	uploadPreset := os.Getenv("CLOUDINARY_UPLOAD_PRESET")

	// Tentukan endpoint upload
	url := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/%s/upload", cloudName, resourceType)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", err
	}
	io.Copy(part, file)

	// Tambahkan folder target di Cloudinary
	if folder != "" {
		writer.WriteField("folder", folder)
	}

	// Tambahkan upload preset (penting!)
	writer.WriteField("upload_preset", uploadPreset)

	writer.Close()

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal upload: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("upload gagal: %s", string(bodyBytes))
	}

	var res CloudinaryResponse
	json.Unmarshal(bodyBytes, &res)

	return res.SecureURL, nil
}
