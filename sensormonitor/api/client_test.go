package sensormonitor

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNormalizeURL(t *testing.T) {
	for in, want := range map[string]string{
		"mlcmonit.mlcgo.eu":                          "https://mlcmonit.mlcgo.eu",
		" https://mlcmonit.mlcgo.eu/ ":               "https://mlcmonit.mlcgo.eu",
		"https://mlcmonit.mlcgo.eu/mlcmonit/api/v1/": "https://mlcmonit.mlcgo.eu",
		"https://mac-studio.local:8443":              "https://mac-studio.local:8443",
		"http://localhost:8080":                      "http://localhost:8080",
		"http://127.0.0.1:9000/x":                    "http://127.0.0.1:9000",
	} {
		if got, err := NormalizeURL(in); err != nil || got != want {
			t.Errorf("NormalizeURL(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"", "http://mlcmonit.mlcgo.eu", "ftp://x", "https://"} {
		if got, err := NormalizeURL(in); err == nil {
			t.Errorf("NormalizeURL(%q) = %q, want an error", in, got)
		}
	}
}

// server answers like the Sensor Monitor behind the gate.
func server(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := New(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	c.Token = "tok"
	c.MinPoll = 10 * time.Millisecond
	return c
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func TestByRefAndCompare(t *testing.T) {
	c := server(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" || r.Header.Get("Accept") != "application/json" {
			t.Errorf("headers: %v", r.Header)
		}
		approved := map[string]any{"id": "a1", "doc_no": "D-1", "ref_no": "SOP-QM-001", "version": "1.0", "status": "approved", "valid_until": "2027-10-01", "title": nil}
		switch r.URL.Path {
		case "/mlcmonit/api/v1/documents/by-ref/SOP-QM-001":
			writeJSON(w, 200, map[string]any{"ref_no": "SOP-QM-001", "approved": approved, "draft": nil})
		case "/mlcmonit/api/v1/documents/by-ref/SOP-QM-001/compare":
			b, _ := io.ReadAll(r.Body)
			if r.Method != http.MethodPost || !strings.HasPrefix(r.Header.Get("Content-Type"), "text/markdown") || string(b) != "# Text\n" {
				t.Errorf("compare: %s %q %q", r.Method, r.Header.Get("Content-Type"), b)
			}
			writeJSON(w, 200, map[string]any{"ref_no": "SOP-QM-001", "approved": approved, "draft": nil, "matches_approved": true, "matches_draft": false})
		default:
			http.NotFound(w, r)
		}
	})
	s, err := c.ByRef(context.Background(), " SOP-QM-001 ")
	if err != nil || s.Approved == nil || s.Approved.Version != "1.0" || s.Approved.ValidUntil != "2027-10-01" || s.Draft != nil {
		t.Fatalf("ByRef = %+v, %v", s, err)
	}
	cmp, err := c.Compare(context.Background(), "SOP-QM-001", "# Text\n")
	if err != nil || !cmp.MatchesApproved || cmp.MatchesDraft || cmp.Approved.ID != "a1" {
		t.Fatalf("Compare = %+v, %v", cmp, err)
	}
}

func TestSource(t *testing.T) {
	c := server(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mlcmonit/api/v1/documents/a1/source" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		_, _ = io.WriteString(w, "---\ndoc_id: SOP-QM-001\n---\n# Text\n")
	})
	got, err := c.Source(context.Background(), "a1")
	if err != nil || got != "---\ndoc_id: SOP-QM-001\n---\n# Text\n" {
		t.Fatalf("Source = %q, %v", got, err)
	}
}

func TestSaveDraft(t *testing.T) {
	c := server(t, func(w http.ResponseWriter, r *http.Request) {
		reason, _ := url.QueryUnescape(r.Header.Get("X-Change-Reason"))
		if r.URL.Path != "/mlcmonit/api/v1/documents/drafts" || r.URL.Query().Get("filename") != "SOP QM.md" ||
			reason != "Prüfung ergänzt" || r.Header.Get("X-Sign-Pin") != "1234" {
			t.Errorf("request: %s %q %q", r.URL, reason, r.Header.Get("X-Sign-Pin"))
		}
		writeJSON(w, 201, map[string]any{"id": "d2", "doc_no": "D-2", "status": "draft", "version": "1.1",
			"warnings": []map[string]any{{"line": 3, "col": 1, "severity": "warning", "code": "header-in-body", "message": "…"}}})
	})
	d, err := c.SaveDraft(context.Background(), "---\ndoc_id: SOP-QM-001\n---\n", DraftOptions{Filename: "SOP QM.md", Reason: "Prüfung ergänzt", PIN: " 1234 "})
	if err != nil || d.ID != "d2" || d.Status != "draft" || len(d.Warnings) != 1 || d.Warnings[0].Code != "header-in-body" {
		t.Fatalf("SaveDraft = %+v, %v", d, err)
	}
}

func TestErrors(t *testing.T) {
	c := server(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/mlcmonit/api/v1/documents/drafts":
			writeJSON(w, 409, map[string]any{"error": map[string]any{"code": "ERR_DOCUMENT_DRAFT_PENDING", "message": "draft pending", "params": map[string]any{"doc_no": "D-7"}}})
		default: // the gate without a valid token
			writeJSON(w, 401, map[string]any{"code": "unauthorized", "error": "invalid or expired token"})
		}
	})
	_, err := c.SaveDraft(context.Background(), "x", DraftOptions{})
	var e *Error
	if !errors.As(err, &e) || e.Status != 409 || e.Code != "ERR_DOCUMENT_DRAFT_PENDING" || e.Param("doc_no") != "D-7" || IsUnauthorized(err) {
		t.Fatalf("draft pending: %#v", err)
	}
	_, err = c.ByRef(context.Background(), "SOP-QM-001")
	if !IsUnauthorized(err) || ErrorCode(err) != "unauthorized" || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("unauthorized: %#v", err)
	}
}

func TestAnalyze(t *testing.T) {
	c := server(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"detected": true, "diagnostics": []map[string]any{{"line": 1, "col": 1, "severity": "error", "code": "x", "message": "m"}},
			"document": map[string]any{"refNo": "SOP-QM-001"}})
	})
	a, err := c.Analyze(context.Background(), "x")
	if err != nil || !a.Detected || len(a.Diagnostics) != 1 || a.Document.RefNo != "SOP-QM-001" {
		t.Fatalf("Analyze = %+v, %v", a, err)
	}
}

func TestDeviceLogin(t *testing.T) {
	var polls atomic.Int32
	c := server(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		switch r.URL.Path {
		case "/mlcauth/device/code":
			if body["client_name"] != "MarkSafe auf PC" {
				t.Errorf("client_name = %q", body["client_name"])
			}
			writeJSON(w, 200, map[string]any{"device_code": "dev", "user_code": "ABCD-EFGH", "verification_uri": "https://x/device",
				"verification_uri_complete": "https://x/device?code=ABCD-EFGH", "expires_in": 600, "interval": 0})
		case "/mlcauth/device/token":
			if body["device_code"] != "dev" {
				t.Errorf("device_code = %q", body["device_code"])
			}
			switch polls.Add(1) {
			case 1:
				writeJSON(w, 400, map[string]any{"code": "authorization_pending", "error": "waiting for approval"})
			case 2:
				writeJSON(w, 400, map[string]any{"code": "slow_down", "error": "poll less often"})
			default:
				writeJSON(w, 200, map[string]any{"access_token": "new", "token_type": "Bearer", "expires_in": 3600, "email": "a@b.c"})
			}
		}
	})
	dc, err := c.StartLogin(context.Background(), "MarkSafe auf PC")
	if err != nil || dc.UserCode != "ABCD-EFGH" {
		t.Fatalf("StartLogin = %+v, %v", dc, err)
	}
	l, err := c.WaitLogin(context.Background(), dc)
	if err != nil || l.Token != "new" || l.Email != "a@b.c" || time.Until(l.Expires) < 59*time.Minute {
		t.Fatalf("WaitLogin = %+v, %v", l, err)
	}
}

func TestDeviceLoginDenied(t *testing.T) {
	c := server(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 400, map[string]any{"code": "access_denied", "error": "the login was denied"})
	})
	_, err := c.WaitLogin(context.Background(), &DeviceCode{DeviceCode: "dev", ExpiresIn: 60})
	if !errors.Is(err, ErrLoginDenied) {
		t.Fatalf("err = %v", err)
	}
}
