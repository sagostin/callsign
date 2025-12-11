package freeswitch

import (
	"callsign/models"
	"encoding/xml"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
)

// XMLCDRData represents the FreeSWITCH XML CDR structure
type XMLCDRData struct {
	XMLName       xml.Name         `xml:"cdr"`
	CoreUUID      string           `xml:"core-uuid,attr"`
	Switchname    string           `xml:"switchname,attr"`
	ChannelData   XMLChannelData   `xml:"channel_data"`
	Variables     XMLVariables     `xml:"variables"`
	Callflow      []XMLCallflow    `xml:"callflow"`
	CallerProfile XMLCallerProfile `xml:"caller_profile,omitempty"`
	Times         XMLTimes         `xml:"times,omitempty"`
}

type XMLChannelData struct {
	State       string `xml:"state"`
	Direction   string `xml:"direction"`
	StateNumber string `xml:"state_number"`
}

type XMLVariables struct {
	UUID              string `xml:"uuid"`
	Direction         string `xml:"direction"`
	CallerIDName      string `xml:"caller_id_name"`
	CallerIDNumber    string `xml:"caller_id_number"`
	DestinationNumber string `xml:"destination_number"`
	DialedNumber      string `xml:"dialed_number"`
	DomainName        string `xml:"domain_name"`
	TenantID          string `xml:"tenant_id"`
	Extension         string `xml:"user_name"`
	HangupCause       string `xml:"hangup_cause"`
	BridgeChannel     string `xml:"bridge_channel"`
	LastBridgeTo      string `xml:"last_bridge_to"`
	SignalBond        string `xml:"signal_bond"`
	Duration          string `xml:"duration"`
	Billsec           string `xml:"billsec"`
	ProgressSec       string `xml:"progresssec"`
	ProgressMediaSec  string `xml:"progress_mediasec"`
	RecordPath        string `xml:"record_path"`
	Context           string `xml:"context"`
	SIPProfileName    string `xml:"sofia_profile_name"`
	GatewayName       string `xml:"sip_gateway_name"`
	CallcenterQueue   string `xml:"cc_queue"`
	SIPResponseCode   string `xml:"sip_term_status"`
}

type XMLCallflow struct {
	Dialplan      string           `xml:"dialplan,attr"`
	CallerProfile XMLCallerProfile `xml:"caller_profile"`
	Times         XMLTimes         `xml:"times"`
}

type XMLCallerProfile struct {
	CallerIDName      string `xml:"caller_id_name"`
	CallerIDNumber    string `xml:"caller_id_number"`
	DestinationNumber string `xml:"destination_number"`
	UUID              string `xml:"uuid"`
	Context           string `xml:"context"`
}

type XMLTimes struct {
	CreatedTime  string `xml:"created_time"`
	AnsweredTime string `xml:"answered_time"`
	BridgeTime   string `xml:"bridged_time"`
	HangupTime   string `xml:"hangup_time"`
}

// HandleXMLCDR processes incoming CDR posts from FreeSWITCH mod_xml_cdr
// POST /freeswitch/cdr
func (h *FSHandler) HandleXMLCDR(ctx iris.Context) {
	// Get the CDR data - can be form param or raw body
	var cdrXML string

	// Check if it's form-encoded (cdr variable)
	cdrXML = ctx.FormValue("cdr")
	if cdrXML == "" {
		// Try raw body
		body, err := ctx.GetBody()
		if err != nil || len(body) == 0 {
			log.Warn("CDR: empty request body")
			ctx.StatusCode(http.StatusBadRequest)
			return
		}
		cdrXML = string(body)
	}

	// URL decode if needed (mod_xml_cdr can URL-encode the XML)
	if decoded, err := url.QueryUnescape(cdrXML); err == nil {
		cdrXML = decoded
	}

	// Parse XML
	var xmlCDR XMLCDRData
	if err := xml.Unmarshal([]byte(cdrXML), &xmlCDR); err != nil {
		log.Warnf("CDR: failed to parse XML: %v", err)
		ctx.StatusCode(http.StatusBadRequest)
		return
	}

	// Convert to CallRecord
	record := h.xmlToCallRecord(&xmlCDR)

	// Save to PostgreSQL
	if err := h.DB.Create(record).Error; err != nil {
		log.Errorf("CDR: database save failed: %v", err)
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{
		"uuid":     record.UUID,
		"caller":   record.CallerIDNumber,
		"dest":     record.DestinationNumber,
		"duration": record.Duration,
		"hangup":   record.HangupCause,
	}).Info("CDR: received and stored")

	// Return 200 OK to indicate successful receipt
	ctx.StatusCode(http.StatusOK)
}

// xmlToCallRecord converts XML CDR to CallRecord model
func (h *FSHandler) xmlToCallRecord(xmlCDR *XMLCDRData) *models.CallRecord {
	vars := xmlCDR.Variables

	// Parse UUID
	callUUID, err := uuid.Parse(urlDecode(vars.UUID))
	if err != nil {
		callUUID = uuid.New()
	}

	// Parse tenant ID
	tenantID := parseUint(vars.TenantID)
	if tenantID == 0 {
		tenantID = 1 // Default tenant
	}

	// Parse times
	startTime := parseMicroseconds(xmlCDR.Times.CreatedTime)
	if startTime.IsZero() && len(xmlCDR.Callflow) > 0 {
		startTime = parseMicroseconds(xmlCDR.Callflow[0].Times.CreatedTime)
	}
	if startTime.IsZero() {
		startTime = time.Now()
	}

	var answerTime, bridgeTime, endTime *time.Time
	if t := parseMicroseconds(xmlCDR.Times.AnsweredTime); !t.IsZero() {
		answerTime = &t
	}
	if t := parseMicroseconds(xmlCDR.Times.BridgeTime); !t.IsZero() {
		bridgeTime = &t
	}
	if t := parseMicroseconds(xmlCDR.Times.HangupTime); !t.IsZero() {
		endTime = &t
	}

	// Parse durations
	duration := parseInt(vars.Duration)
	billableSec := parseInt(vars.Billsec)
	progressSec := parseInt(vars.ProgressSec)

	// Determine direction
	direction := models.CallDirectionInbound
	if urlDecode(vars.Direction) == "outbound" || urlDecode(xmlCDR.ChannelData.Direction) == "outbound" {
		direction = models.CallDirectionOutbound
	}

	// Check for recording
	recorded := vars.RecordPath != ""

	// Check for queue
	queue := vars.CallcenterQueue != ""

	// Parse SIP code
	sipCode := parseInt(vars.SIPResponseCode)

	return &models.CallRecord{
		UUID:              callUUID,
		TenantID:          tenantID,
		CallerIDName:      urlDecode(vars.CallerIDName),
		CallerIDNumber:    urlDecode(vars.CallerIDNumber),
		DestinationNumber: urlDecode(vars.DestinationNumber),
		DialedNumber:      urlDecode(vars.DialedNumber),
		AlegUUID:          urlDecode(vars.UUID),
		BlegUUID:          urlDecode(vars.SignalBond),
		BridgedTo:         urlDecode(vars.LastBridgeTo),
		StartTime:         startTime,
		AnswerTime:        answerTime,
		BridgeTime:        bridgeTime,
		EndTime:           endTime,
		Duration:          duration,
		BillableSec:       billableSec,
		ProgressSec:       progressSec,
		Direction:         direction,
		Context:           urlDecode(vars.Context),
		HangupCause:       urlDecode(vars.HangupCause),
		SIPCode:           sipCode,
		GatewayName:       urlDecode(vars.GatewayName),
		SIPProfileName:    urlDecode(vars.SIPProfileName),
		Recorded:          recorded,
		RecordingPath:     urlDecode(vars.RecordPath),
		Queue:             queue,
		QueueName:         urlDecode(vars.CallcenterQueue),
		Extension:         urlDecode(vars.Extension),
	}
}

// Helper functions
func urlDecode(s string) string {
	if decoded, err := url.QueryUnescape(s); err == nil {
		return decoded
	}
	return s
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(urlDecode(s))
	return i
}

func parseUint(s string) uint {
	i, _ := strconv.ParseUint(urlDecode(s), 10, 32)
	return uint(i)
}

// parseMicroseconds parses FreeSWITCH microsecond timestamps
func parseMicroseconds(s string) time.Time {
	s = urlDecode(s)
	if s == "" || s == "0" {
		return time.Time{}
	}
	microseconds, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}
	}
	seconds := microseconds / 1000000
	nanos := (microseconds % 1000000) * 1000
	return time.Unix(seconds, nanos)
}
