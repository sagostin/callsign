package tts

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"callsign/models"
)

// TTSCache provides a disk-backed audio cache for TTS output.
// Files are stored under basePath with SHA-256 keyed filenames so that
// identical (text, engine, voice) tuples are synthesized only once.
type TTSCache struct {
	basePath string
	db       *gorm.DB

	// In-memory index: cacheKey -> absolute file path
	mu    sync.RWMutex
	index map[string]string

	// System phrase fast-lookup: "phraseKey:lang" -> file path
	phraseMu    sync.RWMutex
	phraseIndex map[string]string
}

// NewCache creates a TTSCache and ensures the base directory exists.
func NewCache(basePath string, db *gorm.DB) (*TTSCache, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("tts cache: create dir %s: %w", basePath, err)
	}

	c := &TTSCache{
		basePath:    basePath,
		db:          db,
		index:       make(map[string]string),
		phraseIndex: make(map[string]string),
	}

	// Rebuild in-memory index from existing files on disk
	c.rebuildIndex()

	return c, nil
}

// CacheKey produces a deterministic key for a (text, engine, voice) tuple.
func CacheKey(text, engine, voice string) string {
	h := sha256.Sum256([]byte(text + "|" + engine + "|" + voice))
	return fmt.Sprintf("%x", h)
}

// Get returns the cached file path for the given key, or ("", false) on miss.
func (c *TTSCache) Get(key string) (string, bool) {
	c.mu.RLock()
	path, ok := c.index[key]
	c.mu.RUnlock()
	if !ok {
		return "", false
	}
	// Verify the file still exists on disk
	if _, err := os.Stat(path); err != nil {
		c.mu.Lock()
		delete(c.index, key)
		c.mu.Unlock()
		return "", false
	}
	return path, true
}

// Put registers a synthesized file in the cache under the given key.
func (c *TTSCache) Put(key, filePath string) {
	c.mu.Lock()
	c.index[key] = filePath
	c.mu.Unlock()
}

// FilePath returns the expected on-disk path for a cache key.
func (c *TTSCache) FilePath(key string) string {
	return filepath.Join(c.basePath, key+".wav")
}

// PhraseAudioPath returns the cached audio path for a system phrase + language.
func (c *TTSCache) PhraseAudioPath(phraseKey, language string) (string, bool) {
	c.phraseMu.RLock()
	defer c.phraseMu.RUnlock()
	p, ok := c.phraseIndex[phraseKey+":"+language]
	return p, ok
}

// SetPhraseAudioPath sets the cached audio path for a system phrase + language.
func (c *TTSCache) SetPhraseAudioPath(phraseKey, language, path string) {
	c.phraseMu.Lock()
	c.phraseIndex[phraseKey+":"+language] = path
	c.phraseMu.Unlock()
}

// WarmSystemPhrases pre-generates audio for all SystemPhrase translations
// that do not already have a cached audio file.
func (c *TTSCache) WarmSystemPhrases(synthesize func(text, engine, voice string) (string, error)) {
	if c.db == nil {
		return
	}

	var phrases []models.SystemPhrase
	if err := c.db.Preload("Translations").Find(&phrases).Error; err != nil {
		log.Warnf("tts cache: failed to load system phrases: %v", err)
		return
	}

	for _, p := range phrases {
		for i := range p.Translations {
			t := &p.Translations[i]

			// Already have audio?
			if t.AudioPath != "" {
				if _, err := os.Stat(t.AudioPath); err == nil {
					c.SetPhraseAudioPath(p.PhraseKey, t.Language, t.AudioPath)
					continue
				}
			}

			// Synthesize and cache
			path, err := synthesize(t.Text, "flite", "kal")
			if err != nil {
				log.Warnf("tts cache: warm phrase %s/%s: %v", p.PhraseKey, t.Language, err)
				continue
			}

			// Update DB
			c.db.Model(t).Update("audio_path", path)
			c.SetPhraseAudioPath(p.PhraseKey, t.Language, path)
			log.Infof("tts cache: warmed phrase %s/%s -> %s", p.PhraseKey, t.Language, path)
		}
	}
}

// rebuildIndex scans the cache directory and populates the in-memory index
// from existing .wav files (whose names are SHA-256 keys).
func (c *TTSCache) rebuildIndex() {
	entries, err := os.ReadDir(c.basePath)
	if err != nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := filepath.Ext(name)
		if ext == ".wav" {
			key := name[:len(name)-len(ext)]
			c.index[key] = filepath.Join(c.basePath, name)
		}
	}
	log.Infof("tts cache: rebuilt index with %d entries from %s", len(c.index), c.basePath)
}

// Stats returns cache statistics.
func (c *TTSCache) Stats() map[string]interface{} {
	c.mu.RLock()
	cached := len(c.index)
	c.mu.RUnlock()

	c.phraseMu.RLock()
	phrases := len(c.phraseIndex)
	c.phraseMu.RUnlock()

	return map[string]interface{}{
		"cached_files":   cached,
		"warmed_phrases": phrases,
		"cache_dir":      c.basePath,
	}
}
