package messaging

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "image/gif" // Register GIF decoder

	log "github.com/sirupsen/logrus"
)

// TranscodeConfig holds transcoding settings
type TranscodeConfig struct {
	MaxSizeKB  int    // Target max file size in KB (default 600 for Tier 2 carriers)
	MaxWidth   int    // Max image/video width
	MaxHeight  int    // Max image/video height
	FFmpegPath string // Path to ffmpeg binary
	TmpDir     string // Temp directory for ffmpeg operations
}

// DefaultTranscodeConfig returns sensible defaults
func DefaultTranscodeConfig() TranscodeConfig {
	return TranscodeConfig{
		MaxSizeKB:  600,
		MaxWidth:   1920,
		MaxHeight:  1920,
		FFmpegPath: "ffmpeg",
		TmpDir:     "/tmp/callsign-transcode",
	}
}

// Transcoder handles media transcoding for MMS delivery
type Transcoder struct {
	config TranscodeConfig
}

// NewTranscoder creates a new transcoder
func NewTranscoder(cfg TranscodeConfig) *Transcoder {
	// Ensure temp directory exists
	os.MkdirAll(cfg.TmpDir, 0755)
	return &Transcoder{config: cfg}
}

// TranscodeResult holds the result of a transcoding operation
type TranscodeResult struct {
	Data        []byte
	ContentType string
	Width       int
	Height      int
	SizeBytes   int64
	Thumbnail   []byte // Optional thumbnail (for images/video)
}

// NeedsTranscoding checks if media needs transcoding based on size and type
func (t *Transcoder) NeedsTranscoding(data []byte, contentType string) bool {
	maxBytes := t.config.MaxSizeKB * 1024

	// Always transcode if over size limit
	if len(data) > maxBytes {
		return true
	}

	// Transcode non-standard formats
	ct := strings.ToLower(contentType)
	switch {
	case strings.HasPrefix(ct, "image/"):
		// JPEG and PNG under limit are fine
		if (ct == "image/jpeg" || ct == "image/png") && len(data) <= maxBytes {
			return false
		}
		return true
	case strings.HasPrefix(ct, "video/"):
		// Videos almost always need transcoding for carrier compat
		return true
	case strings.HasPrefix(ct, "audio/"):
		return len(data) > maxBytes
	}

	return false
}

// TranscodeImage transcodes an image to JPEG with adaptive quality
func (t *Transcoder) TranscodeImage(data []byte, contentType string) (*TranscodeResult, error) {
	maxBytes := t.config.MaxSizeKB * 1024

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Resize if dimensions exceed limits
	if width > t.config.MaxWidth || height > t.config.MaxHeight {
		img = resizeImage(img, t.config.MaxWidth, t.config.MaxHeight)
		bounds = img.Bounds()
		width = bounds.Dx()
		height = bounds.Dy()
	}

	// Adaptive quality loop — start at 85, decrease by 5 until under target
	for quality := 85; quality >= 50; quality -= 5 {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, fmt.Errorf("JPEG encode failed at quality %d: %w", quality, err)
		}

		if buf.Len() <= maxBytes {
			log.WithFields(log.Fields{
				"original_size":   len(data),
				"transcoded_size": buf.Len(),
				"quality":         quality,
				"dimensions":      fmt.Sprintf("%dx%d", width, height),
			}).Debug("Image transcoded successfully")

			// Generate thumbnail (320px max)
			thumb := generateThumbnail(img, 320)

			return &TranscodeResult{
				Data:        buf.Bytes(),
				ContentType: "image/jpeg",
				Width:       width,
				Height:      height,
				SizeBytes:   int64(buf.Len()),
				Thumbnail:   thumb,
			}, nil
		}
	}

	return nil, fmt.Errorf("image too large: could not compress to %dKB even at quality 50", t.config.MaxSizeKB)
}

// TranscodeVideo transcodes a video using ffmpeg
func (t *Transcoder) TranscodeVideo(data []byte, contentType string) (*TranscodeResult, error) {
	// Write input to temp file
	inputPath := filepath.Join(t.config.TmpDir, fmt.Sprintf("input_%d.tmp", randomSuffix()))
	outputPath := filepath.Join(t.config.TmpDir, fmt.Sprintf("output_%d.mp4", randomSuffix()))
	thumbPath := filepath.Join(t.config.TmpDir, fmt.Sprintf("thumb_%d.jpg", randomSuffix()))

	defer os.Remove(inputPath)
	defer os.Remove(outputPath)
	defer os.Remove(thumbPath)

	if err := os.WriteFile(inputPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp input: %w", err)
	}

	// Calculate target bitrate from size limit
	// Rough estimate: target_bytes * 8 / duration_seconds
	// For MMS, keep it simple: 500kbps max video, 128k audio
	maxWidth := 1280
	maxHeight := 720
	if t.config.MaxWidth < maxWidth {
		maxWidth = t.config.MaxWidth
	}
	if t.config.MaxHeight < maxHeight {
		maxHeight = t.config.MaxHeight
	}

	// FFmpeg command for carrier-compatible video
	args := []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-profile:v", "baseline",
		"-level", "3.0",
		"-c:a", "aac",
		"-ac", "2",
		"-b:a", "128k",
		"-movflags", "+faststart",
		"-maxrate", "500k",
		"-bufsize", "1M",
		"-vf", fmt.Sprintf("scale='min(%d,iw)':min'(%d,ih)':force_original_aspect_ratio=decrease", maxWidth, maxHeight),
		"-y",
		outputPath,
	}

	cmd := exec.Command(t.config.FFmpegPath, args...)
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"output": string(cmdOutput),
		}).Error("FFmpeg video transcoding failed")
		return nil, fmt.Errorf("ffmpeg failed: %w", err)
	}

	// Read output
	outputData, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read transcoded video: %w", err)
	}

	// Check size limit
	maxBytes := t.config.MaxSizeKB * 1024
	if len(outputData) > maxBytes {
		return nil, fmt.Errorf("video too large after transcoding: %d bytes (limit %d)", len(outputData), maxBytes)
	}

	// Generate thumbnail from first frame
	thumbArgs := []string{
		"-i", outputPath,
		"-vframes", "1",
		"-vf", "scale=320:-1",
		"-y",
		thumbPath,
	}
	exec.Command(t.config.FFmpegPath, thumbArgs...).Run()
	var thumbnail []byte
	if thumbData, err := os.ReadFile(thumbPath); err == nil {
		thumbnail = thumbData
	}

	log.WithFields(log.Fields{
		"original_size":   len(data),
		"transcoded_size": len(outputData),
	}).Debug("Video transcoded successfully")

	return &TranscodeResult{
		Data:        outputData,
		ContentType: "video/mp4",
		SizeBytes:   int64(len(outputData)),
		Thumbnail:   thumbnail,
	}, nil
}

// TranscodeAudio transcodes audio to AAC using ffmpeg
func (t *Transcoder) TranscodeAudio(data []byte, contentType string) (*TranscodeResult, error) {
	inputPath := filepath.Join(t.config.TmpDir, fmt.Sprintf("input_%d.tmp", randomSuffix()))
	outputPath := filepath.Join(t.config.TmpDir, fmt.Sprintf("output_%d.m4a", randomSuffix()))

	defer os.Remove(inputPath)
	defer os.Remove(outputPath)

	if err := os.WriteFile(inputPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp input: %w", err)
	}

	args := []string{
		"-i", inputPath,
		"-c:a", "aac",
		"-b:a", "128k",
		"-ac", "2",
		"-y",
		outputPath,
	}

	cmd := exec.Command(t.config.FFmpegPath, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.WithField("output", string(output)).Error("FFmpeg audio transcoding failed")
		return nil, fmt.Errorf("ffmpeg failed: %w", err)
	}

	outputData, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read transcoded audio: %w", err)
	}

	return &TranscodeResult{
		Data:        outputData,
		ContentType: "audio/mp4",
		SizeBytes:   int64(len(outputData)),
	}, nil
}

// Transcode automatically detects media type and transcodes appropriately
func (t *Transcoder) Transcode(data []byte, contentType string) (*TranscodeResult, error) {
	ct := strings.ToLower(contentType)

	switch {
	case strings.HasPrefix(ct, "image/"):
		return t.TranscodeImage(data, contentType)
	case strings.HasPrefix(ct, "video/"):
		return t.TranscodeVideo(data, contentType)
	case strings.HasPrefix(ct, "audio/"):
		return t.TranscodeAudio(data, contentType)
	default:
		return nil, fmt.Errorf("unsupported media type for transcoding: %s", contentType)
	}
}

// resizeImage scales an image to fit within maxWidth x maxHeight while maintaining aspect ratio
func resizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	// Calculate scale factor
	scaleW := float64(maxWidth) / float64(w)
	scaleH := float64(maxHeight) / float64(h)
	scale := scaleW
	if scaleH < scale {
		scale = scaleH
	}

	if scale >= 1.0 {
		return img // No resize needed
	}

	newW := int(float64(w) * scale)
	newH := int(float64(h) * scale)

	// Simple nearest-neighbor resize (for production, consider a better algorithm)
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			srcX := int(float64(x) / scale)
			srcY := int(float64(y) / scale)
			dst.Set(x, y, img.At(srcX+bounds.Min.X, srcY+bounds.Min.Y))
		}
	}

	return dst
}

// generateThumbnail creates a small thumbnail from an image
func generateThumbnail(img image.Image, maxDim int) []byte {
	thumb := resizeImage(img, maxDim, maxDim)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 70}); err != nil {
		return nil
	}
	return buf.Bytes()
}

// randomSuffix generates a simple numeric suffix for temp files
func randomSuffix() int64 {
	return int64(os.Getpid())*1000000 + int64(os.Getuid())
}

// init registers PNG decoder
func init() {
	// PNG decoder registration
	_ = png.Decode
}
