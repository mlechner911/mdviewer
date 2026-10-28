package main

// Connection to an MLC Sensor Monitor (internal/sensormonitor): off until
// an address is set; then sign-in by device login, the token only in
// memory. See plan P-MARKS-20261021-01.

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"marksafe/internal/sensormonitor"
)

type sensorState struct {
	mu     sync.Mutex
	login  *sensormonitor.Login
	cancel context.CancelFunc // a running device login
}

// SensorMonitorStatus is what the frontend shows of the connection.
type SensorMonitorStatus struct {
	// Address is "" when the connection is off.
	Address  string `json:"address"`
	SignedIn bool   `json:"signedIn"`
	Email    string `json:"email"`
	// Expires (RFC 3339) of the sign-in; "" when not signed in.
	Expires string `json:"expires"`
}

// SensorLoginCode is the code the user confirms in the browser.
type SensorLoginCode struct {
	UserCode string `json:"userCode"`
	URL      string `json:"url"`
	// ExpiresIn seconds.
	ExpiresIn int `json:"expiresIn"`
}

// SensorLoginResult is the "sensormonitor-login" event at the end of a
// device login. Error is "denied", "expired", "cancelled" or a message.
type SensorLoginResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

// sensorLogin returns the sign-in while it is valid.
func (a *App) sensorLogin() *sensormonitor.Login {
	a.sensor.mu.Lock()
	defer a.sensor.mu.Unlock()
	if l := a.sensor.login; l != nil && (l.Expires.IsZero() || time.Now().Before(l.Expires)) {
		return l
	}
	a.sensor.login = nil
	return nil
}

func (a *App) sensorClient() (*sensormonitor.Client, error) {
	c, err := sensormonitor.New(a.config.GetSensorMonitorURL())
	if err != nil {
		return nil, err
	}
	c.UserAgent = "MarkSafe/" + a.GetVersion()
	if l := a.sensorLogin(); l != nil {
		c.Token = l.Token
	}
	return c, nil
}

// GetSensorMonitorStatus returns the connection state.
func (a *App) GetSensorMonitorStatus() SensorMonitorStatus {
	s := SensorMonitorStatus{Address: a.config.GetSensorMonitorURL()}
	if s.Address == "" {
		return s
	}
	if l := a.sensorLogin(); l != nil {
		s.SignedIn, s.Email = true, l.Email
		if !l.Expires.IsZero() {
			s.Expires = l.Expires.Format(time.RFC3339)
		}
	}
	return s
}

// SetSensorMonitorURL stores the address (normalized; "" turns the
// connection off) and returns it. A new address ends the sign-in. The
// frontend rebuilds the menu afterwards.
func (a *App) SetSensorMonitorURL(address string) (string, error) {
	u := ""
	if address != "" {
		var err error
		if u, err = sensormonitor.NormalizeURL(address); err != nil {
			return "", err
		}
	}
	if u != a.config.GetSensorMonitorURL() {
		a.endSensorLogin()
	}
	return u, a.config.SetSensorMonitorURL(u)
}

// SensorMonitorStartLogin requests a code, opens the confirmation page in
// the browser and waits in the background; the result comes as the event
// "sensormonitor-login".
func (a *App) SensorMonitorStartLogin() (*SensorLoginCode, error) {
	c, err := a.sensorClient()
	if err != nil {
		return nil, err
	}
	host, _ := os.Hostname()
	name := "MarkSafe"
	if host != "" {
		name += " auf " + host
	}
	dc, err := c.StartLogin(context.Background(), name)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dc.ExpiresIn+30)*time.Second)
	a.sensor.mu.Lock()
	if a.sensor.cancel != nil {
		a.sensor.cancel()
	}
	a.sensor.cancel = cancel
	a.sensor.mu.Unlock()
	go func() {
		defer cancel()
		l, err := c.WaitLogin(ctx, dc)
		res := SensorLoginResult{OK: err == nil}
		switch {
		case err == nil:
			a.sensor.mu.Lock()
			a.sensor.login = l
			a.sensor.mu.Unlock()
		case errors.Is(err, sensormonitor.ErrLoginDenied):
			res.Error = "denied"
		case errors.Is(err, sensormonitor.ErrLoginExpired), errors.Is(err, context.DeadlineExceeded):
			res.Error = "expired"
		case errors.Is(err, context.Canceled):
			res.Error = "cancelled"
		default:
			res.Error = err.Error()
		}
		emit("sensormonitor-login", res)
	}()
	u := dc.VerificationURIComplete
	if u == "" {
		u = dc.VerificationURI
	}
	if app := application.Get(); app != nil {
		_ = app.Browser.OpenURL(u)
	}
	return &SensorLoginCode{UserCode: dc.UserCode, URL: u, ExpiresIn: dc.ExpiresIn}, nil
}

// SensorMonitorCancelLogin stops a running device login.
func (a *App) SensorMonitorCancelLogin() {
	a.sensor.mu.Lock()
	defer a.sensor.mu.Unlock()
	if a.sensor.cancel != nil {
		a.sensor.cancel()
		a.sensor.cancel = nil
	}
}

// SensorMonitorLogout forgets the token. The event "sensormonitor-login"
// with ok false lets the frontend update menu and status bar.
func (a *App) SensorMonitorLogout() {
	a.endSensorLogin()
	emit("sensormonitor-login", SensorLoginResult{Error: "signed-out"})
}

func (a *App) endSensorLogin() {
	a.sensor.mu.Lock()
	defer a.sensor.mu.Unlock()
	if a.sensor.cancel != nil {
		a.sensor.cancel()
		a.sensor.cancel = nil
	}
	a.sensor.login = nil
}

// SensorError is a refusal for the frontend to translate by Code
// (ERR_DOCUMENT_DRAFT_PENDING …; "not-signed-in" before any request).
// SignIn: the sign-in is missing or ran out; sign in and try again.
type SensorError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Params  map[string]string `json:"params"`
	SignIn  bool              `json:"signIn"`
}

func (a *App) sensorError(err error) *SensorError {
	if err == nil {
		return nil
	}
	e := &SensorError{Message: err.Error(), Params: map[string]string{}}
	var se *sensormonitor.Error
	if errors.As(err, &se) {
		e.Code, e.Message = se.Code, se.Message
		for k := range se.Params {
			e.Params[k] = se.Param(k)
		}
	}
	if sensormonitor.IsUnauthorized(err) {
		// The token expired or was revoked: forget it, the menu follows.
		e.SignIn = true
		a.endSensorLogin()
		emit("sensormonitor-login", SensorLoginResult{Error: "session-ended"})
	}
	return e
}

// signedInClient is the client with a valid sign-in, or the refusal.
func (a *App) signedInClient() (*sensormonitor.Client, *SensorError) {
	if a.sensorLogin() == nil {
		return nil, &SensorError{Code: "not-signed-in", SignIn: true, Params: map[string]string{}}
	}
	c, err := a.sensorClient()
	if err != nil {
		return nil, &SensorError{Message: err.Error(), Params: map[string]string{}}
	}
	return c, nil
}

// SensorStandResult answers SensorMonitorLookup.
type SensorStandResult struct {
	Stand *sensormonitor.Stand `json:"stand"`
	Error *SensorError         `json:"error"`
}

// SensorMonitorLookup returns the approved version and the waiting draft
// of document id ref.
func (a *App) SensorMonitorLookup(ref string) SensorStandResult {
	c, e := a.signedInClient()
	if e != nil {
		return SensorStandResult{Error: e}
	}
	s, err := c.ByRef(context.Background(), ref)
	return SensorStandResult{Stand: s, Error: a.sensorError(err)}
}

// SensorSourceResult answers SensorMonitorSource.
type SensorSourceResult struct {
	Text  string       `json:"text"`
	Error *SensorError `json:"error"`
}

// SensorMonitorSource returns the text of version id to edit (front matter
// from the document's data, no signature).
func (a *App) SensorMonitorSource(id string) SensorSourceResult {
	c, e := a.signedInClient()
	if e != nil {
		return SensorSourceResult{Error: e}
	}
	text, err := c.Source(context.Background(), id)
	return SensorSourceResult{Text: text, Error: a.sensorError(err)}
}

// SensorCompareResult answers SensorMonitorCompare.
type SensorCompareResult struct {
	Comparison *sensormonitor.Comparison `json:"comparison"`
	Error      *SensorError              `json:"error"`
}

// SensorMonitorCompare compares md with the stored versions of its doc_id.
// Nothing is saved.
func (a *App) SensorMonitorCompare(ref, md string) SensorCompareResult {
	c, e := a.signedInClient()
	if e != nil {
		return SensorCompareResult{Error: e}
	}
	r, err := c.Compare(context.Background(), ref, md)
	return SensorCompareResult{Comparison: r, Error: a.sensorError(err)}
}

// SensorDraftResult answers SensorMonitorSaveDraft.
type SensorDraftResult struct {
	Draft *sensormonitor.SavedDraft `json:"draft"`
	Error *SensorError              `json:"error"`
}

// SensorMonitorSaveDraft stores md as a draft (never approved): a new
// version of the approved document with its doc_id, else a new document.
// reason and pin only when the organisation asks for them.
func (a *App) SensorMonitorSaveDraft(md, filename, reason, pin string) SensorDraftResult {
	c, e := a.signedInClient()
	if e != nil {
		return SensorDraftResult{Error: e}
	}
	d, err := c.SaveDraft(context.Background(), md, sensormonitor.DraftOptions{Filename: filename, Reason: reason, PIN: pin})
	return SensorDraftResult{Draft: d, Error: a.sensorError(err)}
}
