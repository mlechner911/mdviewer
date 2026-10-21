// Package sensormonitor talks to the document API of the MLC Sensor Monitor
// (mlcmonit): open an SOP or form, save the working copy as a draft, and
// compare it with the approved version. Sign-in is the device login of
// mlcauth on the same host (RFC 8628); the token lives only in memory.
//
// MarkSafe cannot approve, set a status or "valid from", or delete: that
// stays in the Sensor Monitor (four-eyes principle, PIN).
//
// The API is described in mlcsensormonitor/.mlcai/MARKSAFE.md, section
// "Schnittstelle für MarkSafe".
package sensormonitor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mlechner911/mdviewer/mlcext"
)

// apiPath is the document API below the host.
const apiPath = "/mlcmonit/api/v1"

// maxAnswer bounds what is read from the server.
const maxAnswer = 16 << 20

// Client of one Sensor Monitor. Token is the access token from the device
// login (or a personal API token); every API call needs it.
type Client struct {
	// BaseURL is scheme and host, e.g. "https://mlcmonit.mlcgo.eu".
	BaseURL   string
	Token     string
	HTTP      *http.Client
	UserAgent string
	// MinPoll is the shortest wait between two polls of the device login
	// (tests make it short).
	MinPoll time.Duration
}

// New returns a client for the Sensor Monitor at address. NormalizeURL
// says what is accepted.
func New(address string) (*Client, error) {
	base, err := NormalizeURL(address)
	if err != nil {
		return nil, err
	}
	return &Client{BaseURL: base, HTTP: &http.Client{Timeout: 30 * time.Second}}, nil
}

// NormalizeURL turns what a user enters ("mlcmonit.mlcgo.eu",
// "https://mlcmonit.mlcgo.eu/mlcmonit/api/v1/") into scheme and host.
// The token goes with every request, so the address must use https;
// plain http only for this computer (localhost), for tests.
func NormalizeURL(address string) (string, error) {
	s := strings.TrimSpace(address)
	if s == "" {
		return "", errors.New("no address of the Sensor Monitor")
	}
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("not an address: %q", address)
	}
	switch u.Scheme {
	case "https":
	case "http":
		if !isLoopback(u.Hostname()) {
			return "", fmt.Errorf("the address must use https: %q", address)
		}
	default:
		return "", fmt.Errorf("not an address: %q", address)
	}
	return u.Scheme + "://" + u.Host, nil
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// Error is a refusal of the server. Code is the fixed code to translate
// (ERR_DOCUMENT_DRAFT_PENDING …, or mlcauth's "unauthorized",
// "authorization_pending" …); Message is the server's fallback text.
type Error struct {
	Status  int
	Code    string
	Message string
	Params  map[string]any
}

func (e *Error) Error() string {
	switch {
	case e.Code != "" && e.Message != "":
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	case e.Code != "":
		return e.Code
	case e.Message != "":
		return fmt.Sprintf("HTTP %d: %s", e.Status, e.Message)
	}
	return fmt.Sprintf("HTTP %d", e.Status)
}

// Param returns params[name] as text ("" when missing).
func (e *Error) Param(name string) string {
	if v, ok := e.Params[name]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

// IsUnauthorized: the token is missing, expired or revoked; sign in again.
func IsUnauthorized(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Status == http.StatusUnauthorized
}

// ErrorCode is the server's code of err, "" for other errors.
func ErrorCode(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

// request sends body (JSON for a value, the text itself for a string)
// and returns the answer.
func (c *Client) request(ctx context.Context, method, path string, body any, header http.Header) (*http.Response, []byte, error) {
	var rd io.Reader
	contentType := ""
	switch b := body.(type) {
	case nil:
	case string:
		rd, contentType = strings.NewReader(b), "text/markdown; charset=utf-8"
	default:
		j, err := json.Marshal(b)
		if err != nil {
			return nil, nil, err
		}
		rd, contentType = bytes.NewReader(j), "application/json"
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, rd)
	if err != nil {
		return nil, nil, err
	}
	for k, vs := range header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	// application/json: without it the gate answers with a redirect to the
	// login page instead of 401.
	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	hc := c.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxAnswer))
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp, data, parseError(resp.StatusCode, data)
	}
	return resp, data, nil
}

// do is request with a JSON answer decoded into out.
func (c *Client) do(ctx context.Context, method, path string, body, out any, header http.Header) error {
	_, data, err := c.request(ctx, method, path, body, header)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("%s %s: unexpected answer: %w", method, path, err)
	}
	return nil
}

// parseError reads both forms: the Sensor Monitor's
// {"error":{"code","message","params"}} and mlcauth's {"code","error"}.
func parseError(status int, data []byte) error {
	e := &Error{Status: status}
	var raw struct {
		Code  string          `json:"code"`
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(data, &raw) == nil {
		e.Code = raw.Code
		var body struct {
			Code    string         `json:"code"`
			Message string         `json:"message"`
			Params  map[string]any `json:"params"`
		}
		var msg string
		switch {
		case json.Unmarshal(raw.Error, &body) == nil:
			e.Code, e.Message, e.Params = body.Code, body.Message, body.Params
		case json.Unmarshal(raw.Error, &msg) == nil:
			e.Message = msg
		}
	}
	return e
}

// Document is a version of a document in the Sensor Monitor (the fields
// MarkSafe uses). Empty fields come as "".
type Document struct {
	ID              string `json:"id"`
	DocNo           string `json:"doc_no"`
	RefNo           string `json:"ref_no"`
	Version         string `json:"version"`
	Title           string `json:"title"`
	Filename        string `json:"filename"`
	Status          string `json:"status"` // draft | approved | superseded
	ApprovedAt      string `json:"approved_at"`
	ApprovedByLabel string `json:"approved_by_label"`
	UploadedByLabel string `json:"uploaded_by_label"`
	// ValidUntil of the approved version: the review is due.
	ValidUntil string `json:"valid_until"`
	CreatedAt  string `json:"created_at"`
}

// Stand of a document id: the approved version and a waiting draft, nil
// when there is none (an unknown id has neither).
type Stand struct {
	RefNo    string    `json:"ref_no"`
	Approved *Document `json:"approved"`
	Draft    *Document `json:"draft"`
}

// Comparison of a working copy with the stored versions. The server
// compares the text as it would store it (mlcext.StoredText), without
// front matter and signature; line endings do not count.
type Comparison struct {
	Stand
	MatchesApproved bool `json:"matches_approved"`
	MatchesDraft    bool `json:"matches_draft"`
}

// Finding of the check on saving (Message is German).
type Finding struct {
	Line     int    `json:"line"`
	Col      int    `json:"col"`
	Severity string `json:"severity"` // error | warning | info
	Code     string `json:"code"`
	Message  string `json:"message"`
}

// SavedDraft is the new draft and the warnings of the check.
type SavedDraft struct {
	Document
	Warnings []Finding `json:"warnings"`
}

// ByRef returns the stand of document id ref.
func (c *Client) ByRef(ctx context.Context, ref string) (*Stand, error) {
	var s Stand
	if err := c.do(ctx, http.MethodGet, apiPath+"/documents/by-ref/"+url.PathEscape(strings.TrimSpace(ref)), nil, &s, nil); err != nil {
		return nil, err
	}
	return &s, nil
}

// Compare says whether md matches the approved version or the draft of ref.
// Nothing is saved.
func (c *Client) Compare(ctx context.Context, ref, md string) (*Comparison, error) {
	var r Comparison
	if err := c.do(ctx, http.MethodPost, apiPath+"/documents/by-ref/"+url.PathEscape(strings.TrimSpace(ref))+"/compare", md, &r, nil); err != nil {
		return nil, err
	}
	return &r, nil
}

// Source returns the text of version id to edit: front matter from the
// document's data (doc_id, version, title, description, event_kinds), no
// signature block.
func (c *Client) Source(ctx context.Context, id string) (string, error) {
	_, data, err := c.request(ctx, http.MethodGet, apiPath+"/documents/"+url.PathEscape(id)+"/source", nil, nil)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DraftOptions go with SaveDraft. Reason and PIN are needed only when the
// organisation asks for them; the server says so with ERR_REASON_REQUIRED
// or ERR_SIGN_PIN_REQUIRED.
type DraftOptions struct {
	Filename string // default "<doc_id>.md"
	Reason   string
	PIN      string
}

// SaveDraft stores md (doc_id in the front matter) as a draft: a new
// version of the approved document with that id, otherwise a new
// document. Never approved.
func (c *Client) SaveDraft(ctx context.Context, md string, opt DraftOptions) (*SavedDraft, error) {
	path := apiPath + "/documents/drafts"
	if opt.Filename != "" {
		path += "?filename=" + url.QueryEscape(opt.Filename)
	}
	h := http.Header{}
	if r := strings.TrimSpace(opt.Reason); r != "" {
		// Headers are ASCII; the server unescapes it.
		h.Set("X-Change-Reason", url.QueryEscape(r))
	}
	if p := strings.TrimSpace(opt.PIN); p != "" {
		h.Set("X-Sign-Pin", p)
	}
	var r SavedDraft
	if err := c.do(ctx, http.MethodPost, path, md, &r, h); err != nil {
		return nil, err
	}
	return &r, nil
}

// Analyze checks md as the server does on saving, without saving.
func (c *Client) Analyze(ctx context.Context, md string) (*mlcext.Analysis, error) {
	var a mlcext.Analysis
	if err := c.do(ctx, http.MethodPost, apiPath+"/mlcext/analyze", md, &a, nil); err != nil {
		return nil, err
	}
	return &a, nil
}
