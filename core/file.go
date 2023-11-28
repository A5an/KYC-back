package core

import (
	"bytes"
	"io"
	"net/http"

	supabase "github.com/supabase-community/storage-go"

	"github.com/Sinbad-HQ/kyc/config"
)

func UploadFile(fileName string, url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	contentType := http.DetectContentType(buf)

	config := config.GetSupabaseConfig()
	client := supabase.NewClient(config.BaseURL+"/storage/v1", config.ApiKey, nil)

	_, err = client.UploadFile(config.Bucket, fileName, bytes.NewReader(buf), supabase.FileOptions{
		ContentType: &contentType,
	})
	if err != nil {
		return "", err
	}

	return client.GetPublicUrl(config.Bucket, fileName).SignedURL, nil
}
