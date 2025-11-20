package config

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

type CloudinaryResponse struct {
	PublicID     string `json:"public_id"`
	SecureURL    string `json:"secure_url"`
	ResourceType string `json:"resource_type"`
	Format       string `json:"format"`
}

type CloudinaryErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// UploadToCloudinary
func UploadToCloudinary(file io.Reader, fileName, folder, resourceType string) (CloudinaryResponse, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return CloudinaryResponse{}, fmt.Errorf("Cloudinary credentials tidak lengkap. Pastikan CLOUDINARY_CLOUD_NAME, API_KEY, API_SECRET dan CLOUDINARY_URL sudah terisi")
	}

	fmt.Printf("☁️ Cloudinary Config loaded: cloud=%s\n", cloudName)

	// Endpoint upload Cloudinary
	url := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/%s/upload", cloudName, resourceType)

	// Timestamp
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Signature
	var signatureString string
	if folder != "" {
		signatureString = fmt.Sprintf("folder=%s&timestamp=%s%s", folder, timestamp, apiSecret)
	} else {
		signatureString = fmt.Sprintf("timestamp=%s%s", timestamp, apiSecret)
	}

	h := sha1.New()
	h.Write([]byte(signatureString))
	signature := hex.EncodeToString(h.Sum(nil))

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return CloudinaryResponse{}, fmt.Errorf("failed to create form file: %v", err) // Perubahan
	}
	if _, err := io.Copy(part, file); err != nil {
		return CloudinaryResponse{}, fmt.Errorf("failed to copy file: %v", err) // Perubahan
	}

	writer.WriteField("api_key", apiKey)
	writer.WriteField("timestamp", timestamp)
	writer.WriteField("signature", signature)

	if folder != "" {
		writer.WriteField("folder", folder)
	}

	writer.Close()

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return CloudinaryResponse{}, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}

	fmt.Println("📤 Sending file to Cloudinary...")

	resp, err := client.Do(req)
	if err != nil {
		return CloudinaryResponse{}, fmt.Errorf("request to Cloudinary failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	fmt.Printf("📥 Cloudinary Status: %d\n", resp.StatusCode)
	fmt.Printf("📥 Cloudinary Body: %s\n", string(respBody))

	if resp.StatusCode != 200 {
		var errResp CloudinaryErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil {
			return CloudinaryResponse{}, fmt.Errorf("cloudinary error: %s", errResp.Error.Message)
		}
		return CloudinaryResponse{}, fmt.Errorf("upload failed: %s", string(respBody))
	}

	var result CloudinaryResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return CloudinaryResponse{}, fmt.Errorf("invalid response format: %v", err)
	}

	fmt.Printf("✅ Upload success: %s\n", result.SecureURL)
	return result, nil
}

type CloudinaryDeleteResponse struct {
	Result string `json:"result"`
}

// DeleteFromCloudinary - menghapus file
func DeleteFromCloudinary(publicID string, resourceType string) error {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return fmt.Errorf("Cloudinary credentials tidak lengkap")
	}

	// Endpoint destroy Cloudinary
	url := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/%s/destroy", cloudName, resourceType)

	// Timestamp
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Signature
	signatureString := fmt.Sprintf("public_id=%s&timestamp=%s%s", publicID, timestamp, apiSecret)
	h := sha1.New()
	h.Write([]byte(signatureString))
	signature := hex.EncodeToString(h.Sum(nil))

	// Form data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("public_id", publicID)
	writer.WriteField("api_key", apiKey)
	writer.WriteField("timestamp", timestamp)
	writer.WriteField("signature", signature)
	writer.Close()

	// Request
	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return fmt.Errorf("failed to create destroy request: %v", err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}

	fmt.Printf("🗑️ Deleting from Cloudinary: %s (Type: %s)\n", publicID, resourceType)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request to Cloudinary destroy failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("🗑️ Cloudinary Delete Status: %d\n", resp.StatusCode)
	fmt.Printf("🗑️ Cloudinary Delete Body: %s\n", string(respBody))

	if resp.StatusCode != 200 {
		var errResp CloudinaryErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil {
			return fmt.Errorf("cloudinary delete error: %s", errResp.Error.Message)
		}
		return fmt.Errorf("delete failed: %s", string(respBody))
	}

	var result CloudinaryDeleteResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("invalid delete response format: %v", err)
	}

	if result.Result == "ok" || result.Result == "not found" {
		fmt.Printf("✅ Delete success (or file not found): %s\n", publicID)
		return nil // Sukses
	}

	return fmt.Errorf("delete failed with result: %s", result.Result)
}
