// Package faceclient calls the internal Python face service (127.0.0.1 only).
package faceclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

type AnalyzeResult struct {
	Embedding     []float32 `json:"embedding"`
	QualityScore  float32   `json:"quality_score"`
	LivenessScore float32   `json:"liveness_score"`
	FaceCount     int       `json:"face_count"`
	EyeOpenness   float32   `json:"eye_openness"`
	// Coarse 8x8 grayscale of the face crop, for spotting a replayed still.
	Signature []int `json:"signature"`
}

// Analyze sends a JPEG frame to the face service and returns embedding + liveness score.
// The server never trusts a verdict computed in the browser; this call is the verdict.
func (c *Client) Analyze(imageJPEG []byte) (*AnalyzeResult, error) {
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/face/analyze", bytes.NewReader(imageJPEG))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "image/jpeg")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("face service returned %d", resp.StatusCode)
	}

	var result AnalyzeResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
