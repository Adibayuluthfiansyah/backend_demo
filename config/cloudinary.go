package config

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var CLD *cloudinary.Cloudinary

func InitCloudinary() error {
	if CLD != nil {
		return nil
	}

	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return fmt.Errorf("cloudinary env tidak lengkap")
	}

	var err error
	CLD, err = cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	return err
}

// Upload file ke folder arsip
// Upload file ke Cloudinary
func UploadToCloudinary(file interface{}, resourceType string) (string, string, error) {
	if err := InitCloudinary(); err != nil {
		return "", "", err
	}

	// Tentukan folder berdasarkan tipe file
	var folder string
	switch resourceType {
	case "image":
		folder = "gambar" // semua gambar masuk folder "gambar"
	case "raw":
		folder = "arsip" // semua PDF masuk folder "arsip"
	default:
		folder = "arsip" // default fallback ke "arsip"
	}

	params := uploader.UploadParams{
		Folder:       folder,
		ResourceType: resourceType, // "image" atau "raw"
		Type:         "upload",
	}

	result, err := CLD.Upload.Upload(context.Background(), file, params)
	if err != nil {
		return "", "", err
	}

	fmt.Printf("Upload berhasil! File tersedia di: %s\n", result.SecureURL)
	fmt.Printf("ResourceType: %s, PublicID: %s\n", result.ResourceType, result.PublicID)

	return result.SecureURL, result.PublicID, nil
}

// Delete file dari Cloudinary dengan fallback ke image jika raw tidak ditemukan
func DeleteFromCloudinary(publicID, resourceType string) error {
	if err := InitCloudinary(); err != nil {
		return err
	}

	invalidate := true
	params := uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: resourceType,
		Type:         "upload",
		Invalidate:   &invalidate,
	}

	res, err := CLD.Upload.Destroy(context.Background(), params)
	if err != nil {
		return err
	}

	// Fallback: jika file raw tidak ditemukan, coba hapus sebagai image
	if res.Result == "not found" && resourceType == "raw" {
		fmt.Printf("File dengan publicID '%s' tidak ditemukan sebagai raw, mencoba image...\n", publicID)
		params.ResourceType = "image"
		res2, err2 := CLD.Upload.Destroy(context.Background(), params)
		if err2 != nil {
			return err2
		}
		switch res2.Result {
		case "ok":
			fmt.Printf("File berhasil dihapus sebagai image: %s\n", publicID)
		case "not found":
			fmt.Printf("File tetap tidak ditemukan: %s\n", publicID)
		default:
			fmt.Printf("Hasil tak terduga dari Cloudinary: %s\n", res2.Result)
		}
		return nil
	}

	// Normal handling
	switch res.Result {
	case "ok":
		fmt.Printf("File berhasil dihapus: %s\n", publicID)
	case "not found":
		fmt.Printf("File dengan publicID '%s' tidak ditemukan di Cloudinary.\n", publicID)
	default:
		fmt.Printf("Hasil tak terduga dari Cloudinary: %s\n", res.Result)
	}

	return nil
}
