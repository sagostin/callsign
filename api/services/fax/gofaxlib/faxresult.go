// faxresult.go — Fax result types and SpanDSP event parsing
// Originally from GOfax.IP project (https://github.com/gonicus/gofaxip)
// Adapted for CallSign fax module

package gofaxlib

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/fiorix/go-eventsocket/eventsocket"
	"github.com/google/uuid"
)

// Resolution is the image resolution of a fax
type Resolution struct {
	X uint
	Y uint
}

func (r Resolution) String() string {
	return fmt.Sprintf("%vx%v", r.X, r.Y)
}

func parseResolution(resstr string) (*Resolution, error) {
	parts := strings.SplitN(resstr, "x", 2)
	if len(parts) != 2 {
		return nil, errors.New("error parsing resolution string")
	}
	res := new(Resolution)
	if x, err := strconv.Atoi(parts[0]); err == nil {
		res.X = uint(x)
	} else {
		return nil, err
	}
	if y, err := strconv.Atoi(parts[1]); err == nil {
		res.Y = uint(y)
	} else {
		return nil, err
	}
	return res, nil
}

// PageResult is the result of a transmitted fax page as reported by SpanDSP
type PageResult struct {
	Ts               time.Time  `json:"ts"`
	Page             uint       `json:"page,omitempty"`
	BadRows          uint       `json:"bad_rows,omitempty"`
	LongestBadRowRun uint       `json:"longest_bad_row_run,omitempty"`
	EncodingName     string     `json:"encoding_name,omitempty"`
	ImagePixelSize   Resolution `json:"image_pixel_size"`
	FilePixelSize    Resolution `json:"file_pixel_size"`
	ImageResolution  Resolution `json:"image_resolution"`
	FileResolution   Resolution `json:"file_resolution"`
	ImageSize        uint       `json:"image_size,omitempty"`
}

func (p PageResult) String() string {
	return fmt.Sprintf("Image Size: %v, Compression: %v, Comp Size: %v bytes, Bad Rows: %v",
		p.ImagePixelSize, p.EncodingName, p.ImageSize, p.BadRows)
}

// FaxResult is the result of a completed or aborted fax transmission
type FaxResult struct {
	UUID uuid.UUID `json:"uuid,omitempty"`

	StartTs time.Time `json:"start_ts"`
	EndTs   time.Time `json:"end_ts"`

	HangupCause string `json:"hangupcause,omitempty"`

	TotalPages       uint         `json:"total_pages,omitempty"`
	TransferredPages uint         `json:"transferred_pages,omitempty"`
	Ecm              bool         `json:"ecm,omitempty"`
	EcmRequested     bool         `json:"ecm_requested,omitempty"`
	RemoteID         string       `json:"remote_id,omitempty"`
	LocalID          string       `json:"local_id,omitempty"`
	ResultCode       int          `json:"result_code,omitempty"`
	ResultText       string       `json:"result_text,omitempty"`
	Success          bool         `json:"success"`
	TransferRate     uint         `json:"transfer_rate,omitempty"`
	NegotiateCount   uint         `json:"negotiate_count,omitempty"`
	T38Status        string       `json:"t38_status,omitempty"`
	V17Disabled      bool         `json:"v17_disabled,omitempty"`
	Bridge           bool         `json:"bridge"`
	PageResults      []PageResult `json:"page_results,omitempty"`
}

// NewFaxResult creates a new FaxResult structure
func NewFaxResult(uuid uuid.UUID, bridge bool) *FaxResult {
	return &FaxResult{
		UUID:   uuid,
		Bridge: bridge,
	}
}

// AddEvent parses a FreeSWITCH EventSocket event and merges contained information into the FaxResult
func (f *FaxResult) AddEvent(ev *eventsocket.Event) {
	switch ev.Get("Event-Name") {
	case "CHANNEL_CALLSTATE":
		callstate := ev.Get("Channel-Call-State")
		log.WithField("uuid", f.UUID.String()).Infof("Call state change: %v", callstate)
		if callstate == "ACTIVE" {
			f.StartTs = time.Now()
		}
		if callstate == "HANGUP" {
			f.EndTs = time.Now()
			f.HangupCause = ev.Get("Hangup-Cause")
		}

	case "CUSTOM":
		subclass := ev.Get("Event-Subclass")
		action := ""

		// Log SpanDSP headers at debug level
		if strings.HasPrefix(subclass, "spandsp::") {
			faxHeaders := make(map[string]string)
			for key, value := range ev.Header {
				if strings.HasPrefix(key, "Fax-") ||
					key == "Event-Subclass" ||
					key == "Channel-Call-Uuid" ||
					strings.Contains(key, "fax_v17") ||
					strings.Contains(key, "fax_ecm") ||
					strings.Contains(key, "fax_t38") ||
					strings.Contains(key, "fax_result") {
					if strVal, ok := value.(string); ok {
						faxHeaders[key] = strVal
					} else {
						faxHeaders[key] = fmt.Sprintf("%v", value)
					}
				}
			}
			log.WithFields(log.Fields{
				"uuid":     f.UUID.String(),
				"subclass": subclass,
				"headers":  faxHeaders,
			}).Debug("SpanDSP event")
		}

		switch subclass {
		case "spandsp::rxfaxnegociateresult",
			"spandsp::txfaxnegociateresult":
			f.NegotiateCount++
			if ecm := ev.Get("Fax-Ecm-Used"); ecm == "on" {
				f.Ecm = true
			}
			if ecmReq := ev.Get("Variable_fax_ecm_requested"); ecmReq == "1" {
				f.EcmRequested = true
			}
			f.RemoteID = ev.Get("Fax-Remote-Station-Id")
			f.LocalID = ev.Get("Fax-Local-Station-Id")
			f.T38Status = ev.Get("Fax-T38-Status")
			if v17 := ev.Get("Variable_fax_v17_disabled"); v17 == "1" {
				f.V17Disabled = true
			}
			if rate, err := strconv.Atoi(ev.Get("Fax-Transfer-Rate")); err == nil {
				f.TransferRate = uint(rate)
			}
			log.WithFields(log.Fields{
				"uuid":          f.UUID.String(),
				"remote_id":     f.RemoteID,
				"local_id":      f.LocalID,
				"transfer_rate": f.TransferRate,
				"ecm":           f.Ecm,
				"t38_status":    f.T38Status,
			}).Info("Fax negotiation result")

		case "spandsp::rxfaxpageresult":
			action = "received"
			fallthrough
		case "spandsp::txfaxpageresult":
			if action == "" {
				action = "sent"
			}
			if pages, err := strconv.Atoi(ev.Get("Fax-Document-Transferred-Pages")); err == nil {
				f.TransferredPages = uint(pages)
			}

			pr := new(PageResult)
			pr.Page = f.TransferredPages

			if badrows, err := strconv.Atoi(ev.Get("Fax-Bad-Rows")); err == nil {
				pr.BadRows = uint(badrows)
			}
			pr.EncodingName = ev.Get("Fax-Encoding-Name")
			if imgsize, err := parseResolution(ev.Get("Fax-Image-Pixel-Size")); err == nil {
				pr.ImagePixelSize = *imgsize
			}
			if filesize, err := parseResolution(ev.Get("Fax-File-Image-Pixel-Size")); err == nil {
				pr.FilePixelSize = *filesize
			}
			if imgres, err := parseResolution(ev.Get("Fax-Image-Resolution")); err == nil {
				pr.ImageResolution = *imgres
			}
			if fileres, err := parseResolution(ev.Get("Fax-File-Image-Resolution")); err == nil {
				pr.FileResolution = *fileres
			}
			if size, err := strconv.Atoi(ev.Get("Fax-Image-Size")); err == nil {
				pr.ImageSize = uint(size)
			}
			if badrowrun, err := strconv.Atoi(ev.Get("Fax-Longest-Bad-Row-Run")); err == nil {
				pr.LongestBadRowRun = uint(badrowrun)
			}

			pr.Ts = time.Now()
			f.PageResults = append(f.PageResults, *pr)
			log.WithField("uuid", f.UUID.String()).Infof("Page %d %v: %v", f.TransferredPages, action, *pr)

		case "spandsp::rxfaxresult",
			"spandsp::txfaxresult":
			if totalpages, err := strconv.Atoi(ev.Get("Fax-Document-Total-Pages")); err == nil {
				f.TotalPages = uint(totalpages)
			}
			if transferredpages, err := strconv.Atoi(ev.Get("Fax-Document-Transferred-Pages")); err == nil {
				f.TransferredPages = uint(transferredpages)
			}
			if ecm := ev.Get("Fax-Ecm-Used"); ecm == "on" {
				f.Ecm = true
			}
			f.RemoteID = ev.Get("Fax-Remote-Station-Id")
			f.LocalID = ev.Get("Fax-Local-Station-Id")
			if rate, err := strconv.Atoi(ev.Get("Fax-Transfer-Rate")); err == nil {
				f.TransferRate = uint(rate)
			}
			if rc, err := strconv.Atoi(ev.Get("Fax-Result-Code")); err == nil {
				f.ResultCode = rc
			}
			f.ResultText = ev.Get("Fax-Result-Text")
			f.Success = ev.Get("Fax-Success") == "1"
			log.WithFields(log.Fields{
				"uuid":              f.UUID.String(),
				"success":           f.Success,
				"result_code":       f.ResultCode,
				"result_text":       f.ResultText,
				"total_pages":       f.TotalPages,
				"transferred_pages": f.TransferredPages,
				"transfer_rate":     f.TransferRate,
			}).Info("Fax result")
		}
	}
}

// EncodeParams encodes given baud rate and ecm status to the status byte
// used in HylaFAX's xferfaxlog.
func EncodeParams(baudrate uint, ecm bool) uint {
	var br uint
	switch {
	case baudrate > 12000:
		br = 5
	case baudrate > 9600:
		br = 4
	case baudrate > 7200:
		br = 3
	case baudrate > 4800:
		br = 2
	case baudrate > 2400:
		br = 1
	}

	var ec uint
	if ecm {
		ec = 1
	}

	return (br << 3) | (ec << 16)
}
