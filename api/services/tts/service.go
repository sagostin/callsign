package tts

import (
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"callsign/config"
)

// Service orchestrates TTS synthesis with caching.
// On first request for a given text+engine+voice, it synthesizes the audio
// and stores the WAV on disk. Subsequent requests return the cached file.
type Service struct {
	Cache  *TTSCache
	Config *config.Config
	DB     *gorm.DB
}

// NewService creates and initialises the TTS service.
func NewService(cfg *config.Config, db *gorm.DB) (*Service, error) {
	cache, err := NewCache(cfg.TTSCachePath, db)
	if err != nil {
		return nil, fmt.Errorf("tts service: %w", err)
	}

	s := &Service{
		Cache:  cache,
		Config: cfg,
		DB:     db,
	}

	return s, nil
}

// Init pre-warms the cache with all system phrase translations.
// Call this after the DB is fully migrated / seeded.
func (s *Service) Init() {
	log.Info("tts service: warming system phrase cache")
	s.Cache.WarmSystemPhrases(s.SynthesizeToFile)
	log.WithFields(log.Fields(s.Cache.Stats())).Info("tts service: cache ready")
}

// SynthesizeToFile generates (or retrieves from cache) a WAV file for the
// given text, engine, and voice.  Returns the absolute file path.
func (s *Service) SynthesizeToFile(text, engine, voice string) (string, error) {
	key := CacheKey(text, engine, voice)

	// Cache hit?
	if path, ok := s.Cache.Get(key); ok {
		return path, nil
	}

	// Cache miss — synthesize
	outPath := s.Cache.FilePath(key)

	if err := synthesizeWAV(text, engine, voice, outPath); err != nil {
		return "", fmt.Errorf("tts synthesize: %w", err)
	}

	s.Cache.Put(key, outPath)
	log.WithFields(log.Fields{
		"engine": engine,
		"voice":  voice,
		"file":   outPath,
	}).Debug("tts: synthesized and cached")

	return outPath, nil
}

// PlaybackCommand returns either a cached file path (to use with the
// FreeSWITCH "playback" app) or falls back to the inline speak/say
// command string if synthesis is unavailable.
func (s *Service) PlaybackCommand(text, engine, voice string) string {
	path, err := s.SynthesizeToFile(text, engine, voice)
	if err != nil {
		log.Warnf("tts: cache miss fallback to speak: %v", err)
		// Fallback to inline TTS
		return ""
	}
	return path
}

// PhraseFile returns the cached audio file for a system phrase, or "".
func (s *Service) PhraseFile(phraseKey, language string) string {
	p, ok := s.Cache.PhraseAudioPath(phraseKey, language)
	if !ok {
		return ""
	}
	return p
}

// synthesizeWAV shells out to the TTS engine binary to produce a WAV file.
// Currently supports flite; extend for other engines as needed.
func synthesizeWAV(text, engine, voice, outPath string) error {
	switch engine {
	case "flite":
		// flite -voice <voice> -t "<text>" -o <outPath>
		args := []string{"-t", text, "-o", outPath}
		if voice != "" && voice != "default" {
			args = append([]string{"-voice", voice}, args...)
		}
		cmd := exec.Command("flite", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("flite: %s: %w", string(out), err)
		}
		return nil

	default:
		// For other engines (ElevenLabs, OpenAI, Google, etc.)
		// an API-based synthesize call would go here.
		return fmt.Errorf("unsupported TTS engine for offline synthesis: %s", engine)
	}
}
