package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(mediaType string) string {
	base := make([]byte, 32)
	_, err := rand.Read(base)
	if err != nil {
		panic("failed to generate random bytes")
	}
	id := base64.RawURLEncoding.EncodeToString(base)
	ext := mediaTypeToExt(mediaType)
	return fmt.Sprintf("%s%s", id, ext)
}

func getVideoAspectRatio(filepath string) (string, error) {
	// Get the aspect ratio of the video using ffprobe
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath)
	ffprobeResult := bytes.Buffer{}
	cmd.Stdout = &ffprobeResult
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	var ffprobeData struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}

	err = json.Unmarshal(ffprobeResult.Bytes(), &ffprobeData)
	if err != nil {
		return "", err
	}

	// Do some math to determine if the aspect ratio is 16:9, 9:16, or other
	aspect := "other"
	stream := ffprobeData.Streams[0]
	fmt.Printf("Width: %d, Height: %d\n", stream.Width, stream.Height)
	w9 := stream.Width / 9
	h9 := stream.Height / 9
	w16 := stream.Width / 16
	h16 := stream.Height / 16

	if w9 == h16 {
		aspect = "9:16"
	}
	if w16 == h9 {
		aspect = "16:9"
	}
	fmt.Println("Aspect ratio: ", aspect)
	return aspect, nil
}

func (cfg apiConfig) getObjectURL(key string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func mediaTypeToExt(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}
