package featurecodes

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// handleTransfer handles transfer to destination
func handleTransfer(ctx *ExecutionContext) {
	fc := ctx.FeatureCode

	if fc.TransferDest != "" {
		context := fc.TransferContext
		if context == "" {
			context = ctx.Domain
		}
		ctx.Conn.Execute("transfer", fmt.Sprintf("%s XML %s", fc.TransferDest, context), true)
	}
}

// handleWebhook calls external webhook
func handleWebhook(ctx *ExecutionContext) {
	fc := ctx.FeatureCode

	if fc.WebhookURL == "" {
		return
	}

	// Build URL with parameters
	url := strings.ReplaceAll(fc.WebhookURL, "${caller_id}", ctx.CallerID)
	url = strings.ReplaceAll(url, "${caller_name}", ctx.CallerName)
	url = strings.ReplaceAll(url, "${domain}", ctx.Domain)
	url = strings.ReplaceAll(url, "${code}", fc.Code)
	url = strings.ReplaceAll(url, "${uuid}", ctx.UUID)
	url = strings.ReplaceAll(url, "${tenant_id}", fmt.Sprintf("%d", ctx.TenantID))

	// Add any captures
	for k, v := range ctx.Captures {
		url = strings.ReplaceAll(url, "${"+k+"}", v)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	method := fc.WebhookMethod
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Errorf("Webhook request failed: %v", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Webhook call failed: %v", err)
		ctx.Conn.Execute("playback", "ivr/ivr-error.wav", true)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		ctx.Conn.Execute("playback", "tone_stream://%(100,0,600);%(100,0,800)", true)
	} else {
		ctx.Conn.Execute("playback", "ivr/ivr-error.wav", true)
	}
}

// handleLua executes a Lua script
func handleLua(ctx *ExecutionContext) {
	fc := ctx.FeatureCode

	if fc.LuaScript != "" {
		// Set variables for Lua script
		ctx.Conn.Execute("set", fmt.Sprintf("fc_caller_id=%s", ctx.CallerID), true)
		ctx.Conn.Execute("set", fmt.Sprintf("fc_domain=%s", ctx.Domain), true)
		ctx.Conn.Execute("set", fmt.Sprintf("fc_code=%s", fc.Code), true)

		// Add captures
		for k, v := range ctx.Captures {
			ctx.Conn.Execute("set", fmt.Sprintf("fc_capture_%s=%s", k, v), true)
		}

		ctx.Conn.Execute("lua", fc.LuaScript, true)
	}
}

// handleCustom executes custom FreeSWITCH commands
func handleCustom(ctx *ExecutionContext) {
	fc := ctx.FeatureCode

	if fc.ActionData == "" {
		return
	}

	// ActionData format: "app1 data1|app2 data2|..."
	commands := strings.Split(fc.ActionData, "|")
	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		// Replace variables
		cmd = strings.ReplaceAll(cmd, "${caller_id}", ctx.CallerID)
		cmd = strings.ReplaceAll(cmd, "${domain}", ctx.Domain)
		cmd = strings.ReplaceAll(cmd, "${code}", fc.Code)

		// Add captures
		for k, v := range ctx.Captures {
			cmd = strings.ReplaceAll(cmd, "${"+k+"}", v)
		}

		// Parse "app data"
		parts := strings.SplitN(cmd, " ", 2)
		app := parts[0]
		data := ""
		if len(parts) > 1 {
			data = parts[1]
		}

		ctx.Conn.Execute(app, data, true)
	}
}
