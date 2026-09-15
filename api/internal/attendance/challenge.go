package attendance

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/hris-face/api/internal/faceclient"
)

const (
	// Below this a frame is a confident spoof and no challenge can rescue it.
	// Between this and minLivenessScore the passive model is unsure, which is
	// what the challenge exists to resolve (docs/PRD.md F5).
	challengeFloor = 0.15

	minChallengeFrames = 3
	// Frames are analysed concurrently, so this caps face-service calls per request.
	maxChallengeFrames = 8
	// Mean absolute difference between two 8x8 face thumbnails, 0-255 scale.
	// Below this the frames are the same picture: a still held to the camera,
	// or a paused video.
	minFrameVariation = 4.0
	// The challenge accepts a lower per-frame liveness than the direct path,
	// because the movement evidence carries part of the weight.
	minChallengeMeanLiveness = 0.30
)

var (
	ErrChallengeRequired = errors.New("perlu verifikasi gerakan")
	ErrTooFewFrames      = errors.New("minimal 3 frame diperlukan")
	ErrStillImage        = errors.New("frame tidak menunjukkan gerakan")
	ErrChallengeFailed   = errors.New("verifikasi gerakan gagal")
)

// ChallengeResult carries what the challenge measured, so an auditor can see
// why a borderline attendance was accepted rather than just that it was.
type ChallengeResult struct {
	Frames         int     `json:"frames"`
	Variation      float64 `json:"variation"`
	MeanLiveness   float64 `json:"mean_liveness"`
	EyeOpenRange   float64 `json:"eye_open_range"`
	BestSimilarity float32 `json:"best_similarity"`
}

// RecordWithChallenge is the fallback path: several frames instead of one.
//
// What this verifies, and what it does not:
//   - VERIFIED: every frame is the logged-in employee, the frames are visually
//     different from each other (a still photo cannot pass), and passive
//     anti-spoof averages above a floor across all of them.
//   - NOT VERIFIED: that the person actually blinked. Eye openness is measured
//     and recorded, but the 106-point landmark model regresses a plausible eye
//     shape rather than tracking eyelids, and we have no closed-eye footage to
//     validate it against. It is a recorded signal, never a gate.
func (s *Service) RecordWithChallenge(ctx context.Context, kind, userID string, frames [][]byte, deviceKey, userAgent, clientIP string, lat, lng *float64) (*Result, *ChallengeResult, string, error) {
	if len(frames) < minChallengeFrames {
		return nil, nil, "", ErrTooFewFrames
	}

	sessionEmployeeID, err := s.employeeForUser(ctx, userID)
	if err != nil {
		return nil, nil, "", err
	}

	type frameData struct {
		embedding []float32
		liveness  float64
		eyeOpen   float64
		signature []int
	}

	if len(frames) > maxChallengeFrames {
		frames = frames[:maxChallengeFrames]
	}
	results := make([]*faceclient.AnalyzeResult, len(frames))
	errs := make([]error, len(frames))
	var wg sync.WaitGroup
	for i, f := range frames {
		wg.Go(func() { results[i], errs[i] = s.face.Analyze(f) })
	}
	wg.Wait()

	analysed := make([]frameData, 0, len(frames))
	for i := range frames {
		a, err := results[i], errs[i]
		if err != nil {
			return nil, nil, "", err
		}
		if a.FaceCount != 1 {
			return nil, nil, "", ErrNoFaceMatch
		}
		if a.LivenessScore < challengeFloor {
			// A frame the passive model is confident about cannot be argued away.
			return nil, nil, "", ErrLivenessFailed
		}
		analysed = append(analysed, frameData{
			embedding: a.Embedding,
			liveness:  float64(a.LivenessScore),
			eyeOpen:   float64(a.EyeOpenness),
			signature: a.Signature,
		})
	}

	variation := meanPairwiseVariation(analysed[0].signature, analysed[len(analysed)-1].signature)
	for i := 1; i < len(analysed); i++ {
		if v := meanPairwiseVariation(analysed[i-1].signature, analysed[i].signature); v > variation {
			variation = v
		}
	}
	if variation < minFrameVariation {
		return nil, nil, "", ErrStillImage
	}

	var sumLiveness float64
	minEye, maxEye := math.MaxFloat64, 0.0
	for _, a := range analysed {
		sumLiveness += a.liveness
		minEye = math.Min(minEye, a.eyeOpen)
		maxEye = math.Max(maxEye, a.eyeOpen)
	}
	meanLiveness := sumLiveness / float64(len(analysed))
	if meanLiveness < minChallengeMeanLiveness {
		return nil, nil, "", ErrChallengeFailed
	}

	// Every frame has to be the same person, and that person has to be whoever
	// is logged in. One stranger's frame in the middle invalidates the sequence.
	var best *matchResult
	for _, a := range analysed {
		m, err := s.findBestMatch(ctx, a.embedding)
		if err != nil {
			return nil, nil, "", err
		}
		if m == nil || m.best < matchThreshold {
			return nil, nil, "", ErrNoFaceMatch
		}
		if m.employeeID != sessionEmployeeID {
			return nil, nil, "", &ErrWrongPerson{Name: m.fullName}
		}
		if best == nil || m.best > best.best {
			best = m
		}
	}

	if err := s.checkGeofence(ctx, best.employeeID, lat, lng, best.allowRemote); err != nil {
		return nil, nil, "", err
	}

	challenge := &ChallengeResult{
		Frames:         len(analysed),
		Variation:      variation,
		MeanLiveness:   meanLiveness,
		EyeOpenRange:   maxEye - minEye,
		BestSimilarity: best.best,
	}

	// The best frame is what gets stored as evidence. issuedKey must reach the
	// caller even if err != nil, for the same reason as record(): resolveDevice
	// inside commit() already wrote the device row before any business-rule
	// check could fail.
	result, issuedKey, err := s.commit(ctx, kind, best, float32(meanLiveness), frames[0], deviceKey, userAgent, clientIP, lat, lng, time.Now())
	if err != nil {
		return nil, nil, issuedKey, err
	}
	return result, challenge, issuedKey, nil
}

// meanPairwiseVariation is the mean absolute difference between two 8x8
// grayscale face thumbnails.
func meanPairwiseVariation(a, b []int) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	sum := 0.0
	for i := range a {
		sum += math.Abs(float64(a[i] - b[i]))
	}
	return sum / float64(len(a))
}
