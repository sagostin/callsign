package ivr

import (
	"callsign/models"
	"callsign/services/esl"
	"callsign/services/messaging"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fiorix/go-eventsocket/eventsocket"
	log "github.com/sirupsen/logrus"
)

const (
	ServiceName    = "ivr"
	ServiceAddress = "127.0.0.7:9001"
)

// Service implements a custom IVR/Auto-Attendant engine using ESL
// Instead of FreeSWITCH mod_ivr, this walks the visual flow graph stored
// in IVRMenu.FlowData, executing each node via ESL commands.
type Service struct {
	*esl.BaseService
	messagingManager *messaging.Manager
}

// New creates a new IVR service
func New() *Service {
	return &Service{
		BaseService: esl.NewBaseService(ServiceName, ServiceAddress),
	}
}

// Init initializes the IVR service
func (s *Service) Init(manager *esl.Manager) error {
	if err := s.BaseService.Init(manager); err != nil {
		return err
	}
	log.Info("Custom IVR engine initialized")
	return nil
}

// SetMessagingManager sets the messaging manager for SMS sending
func (s *Service) SetMessagingManager(mgr *messaging.Manager) {
	s.messagingManager = mgr
}

// Handle processes incoming IVR call connections
func (s *Service) Handle(conn *eventsocket.Connection) {
	defer conn.Close()

	manager := s.Manager()
	if manager == nil {
		log.Error("IVR: manager not initialized")
		return
	}

	// Connect and get channel info
	ev, err := conn.Send("connect")
	if err != nil {
		log.Errorf("IVR: connect failed: %v", err)
		return
	}

	uuid := ev.Get("Unique-ID")
	callerID := ev.Get("Caller-Caller-ID-Number")
	callerName := ev.Get("Caller-Caller-ID-Name")
	dest := ev.Get("Caller-Destination-Number")
	domain := ev.Get("variable_domain_name")
	tenantIDStr := ev.Get("variable_tenant_id")

	logger := log.WithFields(log.Fields{
		"uuid":     uuid,
		"caller":   callerID,
		"dest":     dest,
		"domain":   domain,
		"tenantID": tenantIDStr,
	})
	logger.Info("IVR: handling call")

	conn.Send("linger")
	conn.Send("myevents")

	// Answer the call
	conn.Execute("answer", "", true)

	// Look up the IVR menu by extension
	var menu models.IVRMenu
	if err := manager.DB.Where("extension = ? AND enabled = ?", dest, true).
		Preload("Options").First(&menu).Error; err != nil {
		logger.Warnf("IVR menu not found for extension %s, trying by tenant + ext", dest)
		// Try with tenant filter if variable is set
		if err2 := manager.DB.Where("extension = ? AND enabled = ?", dest, true).
			First(&menu).Error; err2 != nil {
			logger.Errorf("IVR: no menu found for %s", dest)
			conn.Execute("playback", "ivr/ivr-invalid_entry.wav", true)
			conn.Execute("hangup", "", false)
			return
		}
	}

	logger = logger.WithField("ivr_menu", menu.Name)
	logger.Info("IVR: found menu, executing flow")

	// Build execution context
	ctx := &flowContext{
		conn:       conn,
		manager:    manager,
		menu:       &menu,
		uuid:       uuid,
		callerID:   callerID,
		callerName: callerName,
		dest:       dest,
		domain:     domain,
		logger:     logger,
		variables:  make(map[string]string),
	}

	// Set standard channel variables
	ctx.variables["caller_id"] = callerID
	ctx.variables["caller_name"] = callerName
	ctx.variables["destination"] = dest
	ctx.variables["domain"] = domain
	ctx.variables["ivr_name"] = menu.Name

	// Execute the flow graph if we have flow data
	if len(menu.FlowData.Nodes) > 0 {
		s.executeFlowGraph(ctx)
	} else {
		// Fallback: execute traditional IVR from menu options
		s.executeLegacyIVR(ctx)
	}
}

// flowContext holds the execution state for a flow graph
type flowContext struct {
	conn       *eventsocket.Connection
	manager    *esl.Manager
	menu       *models.IVRMenu
	uuid       string
	callerID   string
	callerName string
	dest       string
	domain     string
	logger     *log.Entry
	variables  map[string]string
}

// executeFlowGraph walks the visual flow graph, executing each node
func (s *Service) executeFlowGraph(ctx *flowContext) {
	nodes := ctx.menu.FlowData.Nodes
	connections := ctx.menu.FlowData.Connections

	// Build node map and adjacency list
	nodeMap := make(map[string]*models.IVRFlowNode)
	for i := range nodes {
		nodeMap[nodes[i].ID] = &nodes[i]
	}

	// Build connection map: sourceID+output → targetID
	connMap := make(map[string]string) // "nodeId:output" → targetNodeId
	for _, c := range connections {
		key := c.SourceID + ":" + c.SourceOutput
		connMap[key] = c.TargetID
	}

	// Find the start node — look for ivr_start type, or first node
	var currentNode *models.IVRFlowNode
	for id, node := range nodeMap {
		if node.Type == "ivr_start" {
			currentNode = nodeMap[id]
			break
		}
	}
	if currentNode == nil && len(nodes) > 0 {
		currentNode = &nodes[0]
	}
	if currentNode == nil {
		ctx.logger.Warn("IVR: no start node found in flow")
		ctx.conn.Execute("hangup", "", false)
		return
	}

	// Walk the graph (max 100 steps to prevent infinite loops)
	for step := 0; step < 100 && currentNode != nil; step++ {
		ctx.logger.WithFields(log.Fields{
			"step":      step,
			"node_id":   currentNode.ID,
			"node_type": currentNode.Type,
		}).Debug("IVR: executing node")

		output := s.executeNode(ctx, currentNode)

		if output == "__hangup__" || output == "" {
			break
		}

		// Find next node via connection
		key := currentNode.ID + ":" + output
		nextID, ok := connMap[key]
		if !ok {
			// Try generic "next" connection
			key = currentNode.ID + ":next"
			nextID, ok = connMap[key]
			if !ok {
				// Try connection with empty output
				key = currentNode.ID + ":"
				nextID, ok = connMap[key]
			}
		}

		if ok {
			currentNode = nodeMap[nextID]
		} else {
			ctx.logger.WithField("output", output).Debug("IVR: no connection for output, ending flow")
			currentNode = nil
		}
	}

	// If we fell through without hanging up, hang up
	ctx.conn.Execute("hangup", "", false)
}

// executeNode executes a single flow node and returns the output port name
func (s *Service) executeNode(ctx *flowContext, node *models.IVRFlowNode) string {
	config := node.Config

	switch node.Type {
	case "ivr_start":
		return "next"

	case "gather":
		return s.nodeGather(ctx, config)

	case "play_audio":
		return s.nodePlayAudio(ctx, config)

	case "play_tts":
		return s.nodePlayTTS(ctx, config)

	case "say_digits":
		return s.nodeSayDigits(ctx, config)

	case "web_request":
		return s.nodeWebRequest(ctx, config)

	case "send_sms":
		return s.nodeSendSMS(ctx, config)

	case "condition":
		return s.nodeCondition(ctx, config)

	case "set_variable":
		return s.nodeSetVariable(ctx, config)

	case "extension":
		return s.nodeTransferExtension(ctx, config)

	case "queue":
		return s.nodeTransferQueue(ctx, config)

	case "ring_group":
		return s.nodeTransferRingGroup(ctx, config)

	case "ivr_menu":
		return s.nodeTransferIVR(ctx, config)

	case "external":
		return s.nodeTransferExternal(ctx, config)

	case "voicemail":
		return s.nodeTransferVoicemail(ctx, config)

	case "speech":
		return s.nodeSpeech(ctx, config)

	case "database":
		return s.nodeDatabase(ctx, node)

	case "hangup":
		ctx.conn.Execute("hangup", "", false)
		return "__hangup__"

	default:
		ctx.logger.Warnf("IVR: unknown node type: %s", node.Type)
		return "next"
	}
}

// =====================
// Node Implementations
// =====================

// nodeGather collects DTMF digits from the caller
func (s *Service) nodeGather(ctx *flowContext, config map[string]interface{}) string {
	minDigits := getConfigInt(config, "minDigits", 1)
	maxDigits := getConfigInt(config, "maxDigits", 1)
	timeout := getConfigInt(config, "timeout", 10)
	terminator := getConfigStr(config, "terminator", "#")
	invalidSound := getConfigStr(config, "invalidSound", "")
	maxRetries := getConfigInt(config, "maxRetries", 3)
	validPattern := getConfigStr(config, "validPattern", "")

	// Determine prompt
	promptType := getConfigStr(config, "promptType", "tts")
	var promptFile string
	if promptType == "audio" {
		promptFile = getConfigStr(config, "audioFile", "silence_stream://250")
	} else {
		ttsText := getConfigStr(config, "ttsText", "Please make your selection")
		// Try cached TTS playback first
		if ctx.manager.TTS != nil {
			if cached := ctx.manager.TTS.PlaybackCommand(ttsText, "flite", "kal"); cached != "" {
				promptFile = cached
			} else {
				promptFile = fmt.Sprintf("say:%s", ttsText)
			}
		} else {
			promptFile = fmt.Sprintf("say:%s", ttsText)
		}
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Use play_and_get_digits for prompt + capture
		cmd := fmt.Sprintf("%d %d 1 %d %s %s %s digits \\d+ %d",
			minDigits, maxDigits, timeout*1000, terminator, promptFile, invalidSound, timeout*1000)
		ctx.conn.Execute("play_and_get_digits", cmd, true)

		// Get the captured digits from channel variable
		ev, err := ctx.conn.Send("api uuid_getvar " + ctx.uuid + " digits")
		if err != nil {
			ctx.logger.Errorf("IVR gather: failed to get digits: %v", err)
			return "timeout"
		}

		digits := strings.TrimSpace(ev.Body)
		if digits == "" || digits == "_undef_" {
			if attempt < maxRetries-1 {
				continue
			}
			return "timeout"
		}

		// Validate against pattern if specified
		if validPattern != "" {
			// Simple digit matching (for regex, would need regexp package)
			if !isValidInput(digits, validPattern) {
				if attempt < maxRetries-1 {
					if invalidSound != "" {
						ctx.conn.Execute("playback", invalidSound, true)
					}
					continue
				}
				return "invalid"
			}
		}

		// Store the captured digits
		ctx.variables["caller_input"] = digits
		ctx.variables["gathered_digits"] = digits
		ctx.logger.WithField("digits", digits).Info("IVR: gathered digits")
		return "match"
	}

	return "timeout"
}

// nodePlayAudio plays a sound file
func (s *Service) nodePlayAudio(ctx *flowContext, config map[string]interface{}) string {
	audioFile := getConfigStr(config, "audioFile", "")
	if audioFile == "" {
		return "next"
	}

	loop := getConfigBool(config, "loop", false)
	if loop {
		ctx.conn.Execute("endless_playback", audioFile, true)
	} else {
		ctx.conn.Execute("playback", audioFile, true)
	}
	return "next"
}

// nodePlayTTS speaks text using TTS engine
func (s *Service) nodePlayTTS(ctx *flowContext, config map[string]interface{}) string {
	text := s.resolveVars(ctx, getConfigStr(config, "text", ""))
	if text == "" {
		return "next"
	}

	engine := getConfigStr(config, "engine", "flite")
	voice := getConfigStr(config, "voice", "default")

	// Use cached file if available, else fall back to inline speak
	if ctx.manager.TTS != nil {
		if cached := ctx.manager.TTS.PlaybackCommand(text, engine, voice); cached != "" {
			ctx.conn.Execute("playback", cached, true)
			return "next"
		}
	}

	// Fallback: FreeSWITCH TTS inline speak <engine>|<voice>|<text>
	cmd := fmt.Sprintf("%s|%s|%s", engine, voice, text)
	ctx.conn.Execute("speak", cmd, true)
	return "next"
}

// nodeSayDigits reads digits/numbers aloud
func (s *Service) nodeSayDigits(ctx *flowContext, config map[string]interface{}) string {
	value := s.resolveVars(ctx, getConfigStr(config, "value", ""))
	if value == "" {
		return "next"
	}

	format := getConfigStr(config, "format", "digits")
	switch format {
	case "number":
		ctx.conn.Execute("say", fmt.Sprintf("en number pronounced %s", value), true)
	case "currency":
		ctx.conn.Execute("say", fmt.Sprintf("en currency pronounced %s", value), true)
	default:
		ctx.conn.Execute("say", fmt.Sprintf("en number iterated %s", value), true)
	}
	return "next"
}

// nodeWebRequest makes an HTTP request to an external API
func (s *Service) nodeWebRequest(ctx *flowContext, config map[string]interface{}) string {
	method := getConfigStr(config, "method", "GET")
	url := s.resolveVars(ctx, getConfigStr(config, "url", ""))
	if url == "" {
		return "error"
	}

	timeout := getConfigInt(config, "timeout", 5)
	responseVar := getConfigStr(config, "responseVar", "api_response")
	bodyStr := s.resolveVars(ctx, getConfigStr(config, "body", ""))

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}

	var body io.Reader
	if bodyStr != "" && method != "GET" {
		body = strings.NewReader(bodyStr)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		ctx.logger.Errorf("IVR web_request: failed to create request: %v", err)
		return "error"
	}

	// Parse headers
	headersStr := getConfigStr(config, "headers", "")
	if headersStr != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(headersStr), &headers); err == nil {
			for k, v := range headers {
				req.Header.Set(k, v)
			}
		}
	}

	if method != "GET" && bodyStr != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		ctx.logger.Errorf("IVR web_request: request failed: %v", err)
		ctx.variables[responseVar] = ""
		return "error"
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	ctx.variables[responseVar] = string(respBody)
	ctx.logger.WithFields(log.Fields{
		"status":   resp.StatusCode,
		"response": string(respBody[:min(len(respBody), 200)]),
	}).Info("IVR: web request completed")

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "success"
	}
	return "error"
}

// nodeSendSMS sends an SMS message (via configured provider)
func (s *Service) nodeSendSMS(ctx *flowContext, config map[string]interface{}) string {
	to := s.resolveVars(ctx, getConfigStr(config, "to", ""))
	from := s.resolveVars(ctx, getConfigStr(config, "from", ""))
	msgBody := s.resolveVars(ctx, getConfigStr(config, "body", ""))

	// Validate required fields
	if to == "" || msgBody == "" {
		ctx.logger.Warn("IVR send_sms: missing required fields (to or body)")
		return "error"
	}

	// Check if messaging manager is available
	if s.messagingManager == nil {
		ctx.logger.Error("IVR send_sms: messaging manager not available")
		return "error"
	}

	// Get tenant ID from the IVR menu
	tenantID := ctx.menu.TenantID
	if tenantID == 0 {
		ctx.logger.Error("IVR send_sms: no tenant ID available")
		return "error"
	}

	ctx.logger.WithFields(log.Fields{
		"to":   to,
		"from": from,
		"body": msgBody[:min(len(msgBody), 50)],
	}).Info("IVR: sending SMS")

	// Send the SMS via the messaging manager
	err := s.messagingManager.SendMessage(tenantID, from, to, msgBody, nil, 0)
	if err != nil {
		ctx.logger.WithError(err).Error("IVR send_sms: failed to send SMS")
		return "error"
	}

	ctx.logger.Info("IVR: SMS sent successfully")
	return "sent"
}

// nodeCondition evaluates a condition and branches
func (s *Service) nodeCondition(ctx *flowContext, config map[string]interface{}) string {
	variable := s.resolveVars(ctx, getConfigStr(config, "variable", ""))
	operator := getConfigStr(config, "operator", "==")
	value := s.resolveVars(ctx, getConfigStr(config, "value", ""))

	result := false
	switch operator {
	case "==":
		result = variable == value
	case "!=":
		result = variable != value
	case "contains":
		result = strings.Contains(variable, value)
	case ">":
		result = variable > value
	case "<":
		result = variable < value
	}

	ctx.logger.WithFields(log.Fields{
		"variable": variable, "op": operator, "value": value, "result": result,
	}).Debug("IVR: condition evaluated")

	if result {
		return "true"
	}
	return "false"
}

// nodeSetVariable stores a value in the flow context
func (s *Service) nodeSetVariable(ctx *flowContext, config map[string]interface{}) string {
	name := getConfigStr(config, "name", "")
	value := s.resolveVars(ctx, getConfigStr(config, "value", ""))
	if name != "" {
		ctx.variables[name] = value
	}
	return "next"
}

// nodeTransferExtension transfers the call to an internal extension
func (s *Service) nodeTransferExtension(ctx *flowContext, config map[string]interface{}) string {
	ext := s.resolveVars(ctx, getConfigStr(config, "extension", ""))
	if ext == "" {
		return "__hangup__"
	}
	ctx.logger.WithField("extension", ext).Info("IVR: transferring to extension")
	ctx.conn.Execute("transfer", fmt.Sprintf("%s XML %s", ext, ctx.domain), false)
	return "__hangup__"
}

// nodeTransferQueue transfers the call to a call queue
func (s *Service) nodeTransferQueue(ctx *flowContext, config map[string]interface{}) string {
	queueID := s.resolveVars(ctx, getConfigStr(config, "queueId", ""))
	if queueID == "" {
		return "__hangup__"
	}
	ctx.logger.WithField("queue", queueID).Info("IVR: transferring to queue")
	ctx.conn.Execute("transfer", fmt.Sprintf("%s XML %s", queueID, ctx.domain), false)
	return "__hangup__"
}

// nodeTransferRingGroup transfers to a ring group
func (s *Service) nodeTransferRingGroup(ctx *flowContext, config map[string]interface{}) string {
	groupID := s.resolveVars(ctx, getConfigStr(config, "groupId", ""))
	if groupID == "" {
		return "__hangup__"
	}
	ctx.logger.WithField("ring_group", groupID).Info("IVR: transferring to ring group")
	ctx.conn.Execute("transfer", fmt.Sprintf("%s XML %s", groupID, ctx.domain), false)
	return "__hangup__"
}

// nodeTransferIVR transfers to another IVR menu
func (s *Service) nodeTransferIVR(ctx *flowContext, config map[string]interface{}) string {
	menuID := s.resolveVars(ctx, getConfigStr(config, "menuId", ""))
	if menuID == "" {
		return "__hangup__"
	}
	ctx.logger.WithField("ivr_menu", menuID).Info("IVR: transferring to IVR menu")
	ctx.conn.Execute("transfer", fmt.Sprintf("%s XML %s", menuID, ctx.domain), false)
	return "__hangup__"
}

// nodeTransferExternal bridges to an external phone number
func (s *Service) nodeTransferExternal(ctx *flowContext, config map[string]interface{}) string {
	number := s.resolveVars(ctx, getConfigStr(config, "number", ""))
	if number == "" {
		return "__hangup__"
	}
	ctx.logger.WithField("external", number).Info("IVR: bridging to external number")
	// Bridge via default gateway
	ctx.conn.Execute("bridge", fmt.Sprintf("sofia/gateway/default/%s", number), true)
	return "__hangup__"
}

// nodeTransferVoicemail sends caller to voicemail
func (s *Service) nodeTransferVoicemail(ctx *flowContext, config map[string]interface{}) string {
	mailboxID := s.resolveVars(ctx, getConfigStr(config, "mailboxId", ""))
	if mailboxID == "" {
		mailboxID = ctx.dest // Default to called extension
	}
	ctx.logger.WithField("voicemail", mailboxID).Info("IVR: transferring to voicemail")
	ctx.conn.Execute("transfer", fmt.Sprintf("*99%s XML %s", mailboxID, ctx.domain), false)
	return "__hangup__"
}

// nodeDatabase executes a database query (REST, MySQL, or default)
func (s *Service) nodeDatabase(ctx *flowContext, node *models.IVRFlowNode) string {
	connection := getConfigStr(node.Config, "connection", "default")
	operation := getConfigStr(node.Config, "operation", "query")
	query := s.resolveVars(ctx, getConfigStr(node.Config, "query", ""))
	resultVar := getConfigStr(node.Config, "result_variable", "db_result")

	if query == "" {
		ctx.logger.Warn("IVR database: no query specified")
		return "error"
	}

	var result string
	var err error

	switch connection {
	case "rest":
		result, err = s.executeRestQuery(ctx, query, operation)
	case "mysql":
		result, err = s.executeMySQLQuery(ctx, query, operation)
	default:
		result, err = s.executeDefaultQuery(ctx, query, operation)
	}

	if err != nil {
		ctx.logger.WithFields(log.Fields{
			"connection": connection,
			"operation":  operation,
			"error":      err,
		}).Warn("IVR database: query failed")
		return "error"
	}

	// Store result in flow variable
	ctx.variables[resultVar] = result

	// Determine next node based on result
	if result != "" {
		return "success"
	}
	return "empty"
}

// executeRestQuery performs a REST API query
func (s *Service) executeRestQuery(ctx *flowContext, query, operation string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	var req *http.Request
	var err error

	switch operation {
	case "query":
		req, err = http.NewRequest("GET", query, nil)
	case "insert", "update", "delete":
		req, err = http.NewRequest("POST", query, strings.NewReader("{}"))
	default:
		req, err = http.NewRequest("GET", query, nil)
	}

	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// executeMySQLQuery executes a MySQL query (placeholder)
func (s *Service) executeMySQLQuery(ctx *flowContext, query, operation string) (string, error) {
	ctx.logger.Warn("IVR database: MySQL queries not yet implemented")
	return "", fmt.Errorf("mysql connection not implemented")
}

// executeDefaultQuery executes a default database query (placeholder)
func (s *Service) executeDefaultQuery(ctx *flowContext, query, operation string) (string, error) {
	ctx.logger.Warn("IVR database: default database queries not yet implemented")
	return "", fmt.Errorf("default database connection not implemented")
}

// =====================
// Speech Recognition
// =====================

// SpeechResult holds the result of speech recognition
type SpeechResult struct {
	Text       string
	Confidence float64
}

// nodeSpeech handles speech recognition using FreeSWITCH detect_speech
func (s *Service) nodeSpeech(ctx *flowContext, config map[string]interface{}) string {
	provider := getConfigStr(config, "provider", "google")
	language := getConfigStr(config, "language", "en-US")
	hints := getConfigStr(config, "hints", "")
	timeout := getConfigInt(config, "timeout", 10)
	_ = getConfigInt(config, "maxSpeechNodes", 1) // maxNodes - reserved for multi-utterance support
	variable := getConfigStr(config, "variable", "speech_text")

	ctx.logger.WithFields(log.Fields{
		"provider": provider,
		"language": language,
		"timeout":  timeout,
		"variable": variable,
	}).Info("IVR: speech recognition started")

	// Play prompt if specified
	if nodePrompt := getConfigStr(config, "prompt", ""); nodePrompt != "" {
		promptType := getConfigStr(config, "promptType", "tts")
		if promptType == "audio" {
			ctx.conn.Execute("playback", getConfigStr(config, "audioFile", nodePrompt), true)
		} else {
			ttsText := getConfigStr(config, "ttsText", nodePrompt)
			if ctx.manager.TTS != nil {
				if cached := ctx.manager.TTS.PlaybackCommand(ttsText, "flite", "kal"); cached != "" {
					ctx.conn.Execute("playback", cached, true)
				} else {
					ctx.conn.Execute("speak", fmt.Sprintf("flite|kal|%s", ttsText), true)
				}
			} else {
				ctx.conn.Execute("speak", fmt.Sprintf("flite|kal|%s", ttsText), true)
			}
		}
	}

	// Build the detect_speech command
	// Format: detect_speech <grammar_name> <language> [timeout [params]]
	detectCmd := fmt.Sprintf("detect_speech speech %s %d", language, timeout)
	if hints != "" {
		detectCmd += " " + hints
	}

	// Send detect_speech command
	resp, err := ctx.conn.Send(detectCmd)
	if err != nil {
		ctx.logger.Errorf("IVR speech: failed to start speech recognition: %v", err)
		return "error"
	}

	// Check if detect_speech was accepted
	respBody := strings.TrimSpace(resp.Body)
	if respBody != "+OK" && respBody != "" && !strings.Contains(respBody, "OK") {
		ctx.logger.Warnf("IVR speech: detect_speech returned: %s", respBody)
	}

	// Wait for speech event or timeout
	result := s.waitForSpeechEvent(ctx, timeout)

	// Stop speech recognition
	ctx.conn.Send("detect_speech off")

	if result != nil {
		// Store recognized text
		ctx.variables[variable] = result.Text
		ctx.variables[variable+"_confidence"] = fmt.Sprintf("%.2f", result.Confidence)
		ctx.logger.WithFields(log.Fields{
			"text":       result.Text,
			"confidence": result.Confidence,
		}).Info("IVR: speech recognized")
		return "recognized"
	}

	ctx.logger.Info("IVR: speech recognition timed out")
	return "timeout"
}

// waitForSpeechEvent waits for speech to be detected or timeout to occur
func (s *Service) waitForSpeechEvent(ctx *flowContext, timeout int) *SpeechResult {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)

	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil
		}

		// Set a shorter sleep to check timeout more precisely
		sleepDur := remaining
		if sleepDur > 500*time.Millisecond {
			sleepDur = 500 * time.Millisecond
		}
		time.Sleep(sleepDur)

		// Try to read an event (non-blocking check)
		ev, err := ctx.conn.ReadEvent()
		if err != nil {
			// If we're past the deadline, return nil (timeout)
			if time.Now().After(deadline) {
				return nil
			}
			continue // Try again
		}

		// Check for speech detected event
		speechType := ev.Get("Speech-Type")
		if speechType == "detected" || speechType == "speech-detected" {
			text := ev.Get("Speech-Text")
			if text != "" {
				confidence := 0.85 // Default confidence
				if confStr := ev.Get("Confidence"); confStr != "" {
					if conf, err := strconv.ParseFloat(confStr, 64); err == nil {
						confidence = conf
					}
				}
				return &SpeechResult{
					Text:       text,
					Confidence: confidence,
				}
			}
		}

		// Check for ASR result alternative header
		if resultText := ev.Get("variable_speech_text"); resultText != "" {
			return &SpeechResult{
				Text:       resultText,
				Confidence: 0.85,
			}
		}
	}
}

// =====================
// Legacy IVR (fallback for menus without flow data)
// =====================

// executeLegacyIVR runs a traditional IVR using the IVRMenuOption rows
func (s *Service) executeLegacyIVR(ctx *flowContext) {
	menu := ctx.menu
	logger := ctx.logger

	for attempt := 0; attempt < menu.MaxFailures+menu.MaxTimeouts; attempt++ {
		// Play greeting
		greeting := menu.GreetLong
		if attempt > 0 {
			greeting = menu.GreetShort
			if greeting == "" {
				greeting = menu.GreetLong
			}
		}

		if greeting != "" {
			ctx.conn.Execute("playback", greeting, true)
		}

		// Collect digits
		timeout := menu.Timeout
		if timeout == 0 {
			timeout = 10
		}
		maxDigits := menu.DigitLen
		if maxDigits == 0 {
			maxDigits = 1
		}

		cmd := fmt.Sprintf("1 %d 1 %d # %s %s digits \\d+ %d",
			maxDigits, timeout*1000, "silence_stream://250",
			menu.InvalidSound, timeout*1000)
		ctx.conn.Execute("play_and_get_digits", cmd, true)

		ev, err := ctx.conn.Send("api uuid_getvar " + ctx.uuid + " digits")
		if err != nil {
			logger.Errorf("IVR legacy: failed to get digits: %v", err)
			break
		}

		digits := strings.TrimSpace(ev.Body)
		if digits == "" || digits == "_undef_" {
			// Timeout
			logger.Debug("IVR legacy: timeout, retrying")
			continue
		}

		// Find matching option
		matched := false
		for _, opt := range menu.Options {
			if !opt.Enabled {
				continue
			}
			if opt.Digits == digits {
				matched = true
				logger.WithFields(log.Fields{"digits": digits, "action": opt.Action, "param": opt.ActionParam}).
					Info("IVR legacy: matched option")

				switch opt.Action {
				case models.IVRActionTransfer:
					ctx.conn.Execute("transfer", fmt.Sprintf("%s XML %s", opt.ActionParam, ctx.domain), false)
				case models.IVRActionIVR:
					ctx.conn.Execute("transfer", fmt.Sprintf("%s XML %s", opt.ActionParam, ctx.domain), false)
				case models.IVRActionVoicemail:
					ctx.conn.Execute("transfer", fmt.Sprintf("*99%s XML %s", opt.ActionParam, ctx.domain), false)
				case models.IVRActionQueue:
					ctx.conn.Execute("transfer", fmt.Sprintf("%s XML %s", opt.ActionParam, ctx.domain), false)
				case models.IVRActionRingGroup:
					ctx.conn.Execute("transfer", fmt.Sprintf("%s XML %s", opt.ActionParam, ctx.domain), false)
				case models.IVRActionPlayback:
					ctx.conn.Execute("playback", opt.ActionParam, true)
					continue // Stay in IVR after playback
				case models.IVRActionHangup:
					ctx.conn.Execute("hangup", "", false)
				case models.IVRActionRepeat:
					continue // Re-enter loop
				default:
					ctx.conn.Execute("transfer", fmt.Sprintf("%s XML %s", opt.ActionParam, ctx.domain), false)
				}
				return
			}
		}

		if !matched {
			if menu.InvalidSound != "" {
				ctx.conn.Execute("playback", menu.InvalidSound, true)
			}
		}
	}

	// Exhausted retries
	if menu.ExitSound != "" {
		ctx.conn.Execute("playback", menu.ExitSound, true)
	}
	ctx.conn.Execute("hangup", "", false)
}

// =====================
// Helpers
// =====================

// resolveVars replaces ${variable_name} with context variable values
func (s *Service) resolveVars(ctx *flowContext, input string) string {
	result := input
	for k, v := range ctx.variables {
		result = strings.ReplaceAll(result, "${"+k+"}", v)
	}
	return result
}

// isValidInput checks if digits match a simple pattern
func isValidInput(digits, pattern string) bool {
	// Simple check: if pattern is like "^[1-5]$", just extract the charset
	// For full regex support, use regexp package
	if strings.HasPrefix(pattern, "^[") && strings.HasSuffix(pattern, "]$") {
		chars := pattern[2 : len(pattern)-2]
		for _, d := range digits {
			if !strings.ContainsRune(chars, d) {
				return false
			}
		}
		return len(digits) > 0
	}
	// Fallback: any non-empty input is valid
	return digits != ""
}

func getConfigStr(config map[string]interface{}, key, defaultVal string) string {
	if v, ok := config[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

func getConfigInt(config map[string]interface{}, key string, defaultVal int) int {
	if v, ok := config[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		case int64:
			return int(n)
		}
	}
	return defaultVal
}

func getConfigBool(config map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := config[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultVal
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
