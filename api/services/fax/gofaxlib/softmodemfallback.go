// retrystrategy.go — Progressive fax retry/fallback strategy engine
//
// Replaces the simple mod_db softmodem fallback from gofaxserver with a
// comprehensive per-pair retry strategy that progressively degrades:
//
// Level 0: T.38 + ECM + V.17 (14400 bps) — best case
// Level 1: T.38 + ECM + V.17 disabled (9600 bps) — baudrate fallback
// Level 2: G.711 passthrough + ECM + V.17 (T.38 disabled) — protocol fallback
// Level 3: G.711 passthrough + ECM + V.17 disabled — baudrate + protocol
// Level 4: G.711 passthrough + ECM disabled — last resort
//
// For inbound with bridge/transcoding enabled:
// The bridge leg toward the far-end device uses G.711, while the
// upstream trunk leg uses T.38 if available — FreeSWITCH SpanDSP
// handles the transcoding between the two.

package gofaxlib

import (
	"fmt"
	"sync"
	"time"
)

// RetryLevel defines the progressive fallback levels for fax transmission
type RetryLevel int

const (
	// RetryLevelT38Full — T.38 + ECM + V.17 (best quality, 14400 bps)
	RetryLevelT38Full RetryLevel = 0
	// RetryLevelT38NoV17 — T.38 + ECM, V.17 disabled (falls to 9600 bps)
	RetryLevelT38NoV17 RetryLevel = 1
	// RetryLevelG711Full — G.711 passthrough + ECM + V.17 (no T.38 negotiation)
	RetryLevelG711Full RetryLevel = 2
	// RetryLevelG711NoV17 — G.711 passthrough + ECM, V.17 disabled
	RetryLevelG711NoV17 RetryLevel = 3
	// RetryLevelG711NoECM — G.711 passthrough, no ECM, no V.17 (last resort)
	RetryLevelG711NoECM RetryLevel = 4
)

// MaxRetryLevel is the highest fallback level
const MaxRetryLevel = RetryLevelG711NoECM

// FaxChannelParams holds the FreeSWITCH channel variables for a fax attempt
type FaxChannelParams struct {
	EnableT38  bool   // fax_enable_t38
	RequestT38 bool   // fax_enable_t38_request
	DisableV17 bool   // fax_disable_v17 (forces 9600 bps max)
	DisableECM bool   // fax_ecm_disable
	BridgeMode bool   // Use bridge/transcoding instead of rxfax/txfax
	Codec      string // G.711 ulaw/alaw for non-T.38 legs
	RetryLevel RetryLevel
}

// ParamsForLevel returns the FreeSWITCH channel parameters for a given retry level
func ParamsForLevel(level RetryLevel, bridgeEnabled bool) FaxChannelParams {
	p := FaxChannelParams{
		RetryLevel: level,
		Codec:      "PCMU", // G.711 μ-law default
	}

	switch level {
	case RetryLevelT38Full:
		p.EnableT38 = true
		p.RequestT38 = true
		p.DisableV17 = false
		p.DisableECM = false
	case RetryLevelT38NoV17:
		p.EnableT38 = true
		p.RequestT38 = true
		p.DisableV17 = true
		p.DisableECM = false
	case RetryLevelG711Full:
		p.EnableT38 = false
		p.RequestT38 = false
		p.DisableV17 = false
		p.DisableECM = false
	case RetryLevelG711NoV17:
		p.EnableT38 = false
		p.RequestT38 = false
		p.DisableV17 = true
		p.DisableECM = false
	case RetryLevelG711NoECM:
		p.EnableT38 = false
		p.RequestT38 = false
		p.DisableV17 = true
		p.DisableECM = true
	}

	// Bridge/transcoding: upstream side uses T.38 if available,
	// device side forced to G.711 — FS SpanDSP transcodes between them
	if bridgeEnabled {
		p.BridgeMode = true
	}

	return p
}

// ToFreeSwitchVars converts params to FreeSWITCH channel variable string
func (p FaxChannelParams) ToFreeSwitchVars() string {
	vars := fmt.Sprintf(
		"fax_enable_t38=%v,fax_enable_t38_request=%v",
		p.EnableT38, p.RequestT38,
	)

	if p.DisableV17 {
		vars += ",fax_disable_v17=true"
	}
	if p.DisableECM {
		vars += ",fax_ecm_disable=true"
	}

	// Force G.711 codec when T.38 is disabled (passthrough mode)
	if !p.EnableT38 {
		vars += fmt.Sprintf(",absolute_codec_string=%s", p.Codec)
	}

	return vars
}

// String returns a human-readable description of the retry level
func (p FaxChannelParams) String() string {
	protocol := "T.38"
	if !p.EnableT38 {
		protocol = "G.711"
	}
	baudrate := "V.17 (14400)"
	if p.DisableV17 {
		baudrate = "V.29 (9600)"
	}
	ecm := "ECM"
	if p.DisableECM {
		ecm = "no-ECM"
	}
	mode := "direct"
	if p.BridgeMode {
		mode = "bridge/transcode"
	}
	return fmt.Sprintf("Level %d: %s + %s + %s [%s]", p.RetryLevel, protocol, baudrate, ecm, mode)
}

// PairRetryTTL is how long pair state is remembered before resetting
const PairRetryTTL = 30 * time.Minute

// PairRetryState tracks the retry history for a specific src→dst pair
type PairRetryState struct {
	LastLevel   RetryLevel // What level we used last time
	LastSuccess bool       // Did it succeed?
	LastSeen    time.Time  // When was the last attempt
	Failures    int        // Consecutive failure count at current level
}

// RetryStrategy manages per-pair retry state for the fax module
type RetryStrategy struct {
	mu    sync.Mutex
	pairs map[string]*PairRetryState
}

// NewRetryStrategy creates a new RetryStrategy
func NewRetryStrategy() *RetryStrategy {
	return &RetryStrategy{
		pairs: make(map[string]*PairRetryState),
	}
}

// GetNextLevel determines the retry level for a src→dst pair based on history
//
// Logic (flip-flop inspired):
//   - If no history or state expired: start at Level 0 (T.38 full)
//   - If last attempt succeeded: use the same level
//   - If last attempt failed: escalate to next level
//   - If max level reached and still failing: cycle back to level 0
//     (the remote side might have changed or been rebooted)
func (rs *RetryStrategy) GetNextLevel(srcNum, dstNum string) RetryLevel {
	key := srcNum + "_" + dstNum

	rs.mu.Lock()
	defer rs.mu.Unlock()

	st, ok := rs.pairs[key]
	if !ok || time.Since(st.LastSeen) > PairRetryTTL {
		// No recent history — start fresh at T.38
		return RetryLevelT38Full
	}

	if st.LastSuccess {
		// Last attempt worked — use the same level
		return st.LastLevel
	}

	// Last attempt failed — escalate
	next := st.LastLevel + 1
	if next > MaxRetryLevel {
		// Cycle back — maybe the remote side has changed
		next = RetryLevelT38Full
	}
	return next
}

// RecordAttempt records the result of a fax attempt for a src→dst pair
func (rs *RetryStrategy) RecordAttempt(srcNum, dstNum string, level RetryLevel, success bool) {
	key := srcNum + "_" + dstNum

	rs.mu.Lock()
	defer rs.mu.Unlock()

	st, ok := rs.pairs[key]
	if !ok {
		st = &PairRetryState{}
		rs.pairs[key] = st
	}

	st.LastLevel = level
	st.LastSuccess = success
	st.LastSeen = time.Now()

	if success {
		st.Failures = 0
	} else {
		st.Failures++
	}
}

// GetPairState returns the current retry state for a pair (for API/stats)
func (rs *RetryStrategy) GetPairState(srcNum, dstNum string) *PairRetryState {
	key := srcNum + "_" + dstNum

	rs.mu.Lock()
	defer rs.mu.Unlock()

	st, ok := rs.pairs[key]
	if !ok {
		return nil
	}
	// Return a copy
	copy := *st
	return &copy
}

// Cleanup removes expired pair states
func (rs *RetryStrategy) Cleanup() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	now := time.Now()
	for key, st := range rs.pairs {
		if now.Sub(st.LastSeen) > PairRetryTTL {
			delete(rs.pairs, key)
		}
	}
}
