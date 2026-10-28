// Package sensormonitor is MarkSafe's optional connection to an MLC Sensor
// Monitor: open SOPs and forms from it, save them there as drafts and
// compare the working copy with the approved version.
//
// It is built in only with the build tag "sensormonitor" (task … SENSOR=1,
// see ../sensormonitor_on.go); public MarkSafe builds contain none of it.
// Even when built in it stays off until an address is set. The frontend
// part lives in frontend/src/sensormonitor/ and calls the methods of
// Service by name.
//
// Sign-in is the device login of mlcauth on the same host; the token lives
// only in memory. MarkSafe cannot approve, set a status or "valid from",
// or delete: that stays in the Sensor Monitor.
package sensormonitor

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"marksafe/sensormonitor/api"
)

// Service is the Wails service of the connection.
type Service struct {
	version string
	cfgPath string

	mu      sync.Mutex
	address string     // "" = off
	login   *api.Login // valid sign-in, nil when signed out
	cancel  context.CancelFunc
}

// settings is the module's own file next to MarkSafe's config.json.
type settings struct {
	Address string `json:"address,omitempty"`
}

// New returns the service; version goes into the User-Agent.
func New(version string) *Service {
	s := &Service{version: version}
	if dir, err := os.UserConfigDir(); err == nil {
		s.cfgPath = filepath.Join(dir, "marksafe", "sensormonitor.json")
		s.load(filepath.Join(dir, "marksafe", "config.json"))
	}
	return s
}

// load reads the settings; MarkSafe 1.8 kept the address in config.json
// ("sensor_monitor_url"), which is taken over once.
func (s *Service) load(legacy string) {
	var st settings
	if b, err := os.ReadFile(s.cfgPath); err == nil {
		_ = json.Unmarshal(b, &st)
	} else if b, err := os.ReadFile(legacy); err == nil {
		var old struct {
			URL string `json:"sensor_monitor_url"`
		}
		if json.Unmarshal(b, &old) == nil && old.URL != "" {
			st.Address = old.URL
			s.address = st.Address
			_ = s.save()
		}
	}
	s.address = st.Address
}

func (s *Service) save() error {
	if s.cfgPath == "" {
		return errors.New("no config folder")
	}
	b, err := json.MarshalIndent(settings{Address: s.address}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.cfgPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.cfgPath, b, 0o600)
}

func emit(name string, data ...any) {
	if app := application.Get(); app != nil {
		app.Event.Emit(name, data...)
	}
}

// Extension is how MarkSafe sees the module: the Wails service plus its
// menu entries and settings. Service itself carries only the methods the
// frontend calls (every exported method of a service is bound).
type Extension struct {
	svc *Service
}

// NewExtension returns the module; version goes into the User-Agent.
func NewExtension(version string) *Extension {
	return &Extension{svc: New(version)}
}

// Service is the Wails service to register.
func (e *Extension) Service() application.Service {
	return application.NewService(e.svc)
}

// FileMenu adds the submenu "Sensor Monitor" to MarkSafe's File menu, only
// with an address: for everyone else nothing of it is shown.
func (e *Extension) FileMenu(menu *application.Menu, label func(key, fallback string) string) {
	s := e.svc
	if s.currentAddress() == "" {
		return
	}
	sm := menu.AddSubmenu(label("menuSensorMonitor", "Sensor Monitor"))
	sm.Add(label("menuSensorOpen", "Open…")).OnClick(func(*application.Context) { emit("menu-sm-open") })
	sm.Add(label("menuSensorSaveDraft", "Save as Draft…")).OnClick(func(*application.Context) { emit("menu-sm-save-draft") })
	sm.AddSeparator()
	if s.signedIn() == nil {
		sm.Add(label("menuSensorLogin", "Sign In…")).OnClick(func(*application.Context) { emit("menu-sm-login") })
	} else {
		sm.Add(label("menuSensorLogout", "Sign Out")).OnClick(func(*application.Context) { s.Logout() })
	}
}

// HasSettings: the module has a section in File › Settings….
func (e *Extension) HasSettings() bool { return true }

func (s *Service) currentAddress() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.address
}

// signedIn returns the sign-in while it is valid.
func (s *Service) signedIn() *api.Login {
	s.mu.Lock()
	defer s.mu.Unlock()
	if l := s.login; l != nil && (l.Expires.IsZero() || time.Now().Before(l.Expires)) {
		return l
	}
	s.login = nil
	return nil
}

func (s *Service) client() (*api.Client, error) {
	c, err := api.New(s.currentAddress())
	if err != nil {
		return nil, err
	}
	c.UserAgent = "MarkSafe/" + s.version
	if l := s.signedIn(); l != nil {
		c.Token = l.Token
	}
	return c, nil
}

// Status is what the frontend shows of the connection.
type Status struct {
	// Address is "" when the connection is off.
	Address  string `json:"address"`
	SignedIn bool   `json:"signedIn"`
	Email    string `json:"email"`
	// Expires (RFC 3339) of the sign-in; "" when not signed in.
	Expires string `json:"expires"`
}

// GetStatus returns the connection state.
func (s *Service) GetStatus() Status {
	st := Status{Address: s.currentAddress()}
	if st.Address == "" {
		return st
	}
	if l := s.signedIn(); l != nil {
		st.SignedIn, st.Email = true, l.Email
		if !l.Expires.IsZero() {
			st.Expires = l.Expires.Format(time.RFC3339)
		}
	}
	return st
}

// SetAddress stores the address (normalized; "" turns the connection off)
// and returns it. A new address ends the sign-in. The frontend rebuilds
// the menu afterwards.
func (s *Service) SetAddress(address string) (string, error) {
	u := ""
	if address != "" {
		var err error
		if u, err = api.NormalizeURL(address); err != nil {
			return "", err
		}
	}
	if u != s.currentAddress() {
		s.endLogin()
	}
	s.mu.Lock()
	s.address = u
	err := s.save()
	s.mu.Unlock()
	return u, err
}

// LoginCode is the code the user confirms in the browser.
type LoginCode struct {
	UserCode string `json:"userCode"`
	URL      string `json:"url"`
	// ExpiresIn seconds.
	ExpiresIn int `json:"expiresIn"`
}

// LoginResult is the "sensormonitor-login" event at the end of a device
// login. Error is "denied", "expired", "cancelled", "signed-out",
// "session-ended" or a message.
type LoginResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

// StartLogin requests a code, opens the confirmation page in the browser
// and waits in the background; the result comes as the event
// "sensormonitor-login".
func (s *Service) StartLogin() (*LoginCode, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	name := "MarkSafe"
	if host, _ := os.Hostname(); host != "" {
		name += " auf " + host
	}
	dc, err := c.StartLogin(context.Background(), name)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dc.ExpiresIn+30)*time.Second)
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	s.cancel = cancel
	s.mu.Unlock()
	go func() {
		defer cancel()
		l, err := c.WaitLogin(ctx, dc)
		res := LoginResult{OK: err == nil}
		switch {
		case err == nil:
			s.mu.Lock()
			s.login = l
			s.mu.Unlock()
		case errors.Is(err, api.ErrLoginDenied):
			res.Error = "denied"
		case errors.Is(err, api.ErrLoginExpired), errors.Is(err, context.DeadlineExceeded):
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
	return &LoginCode{UserCode: dc.UserCode, URL: u, ExpiresIn: dc.ExpiresIn}, nil
}

// CancelLogin stops a running device login.
func (s *Service) CancelLogin() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
}

// Logout forgets the token. The event "sensormonitor-login" with ok false
// lets the frontend update menu and status bar.
func (s *Service) Logout() {
	s.endLogin()
	emit("sensormonitor-login", LoginResult{Error: "signed-out"})
}

func (s *Service) endLogin() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.login = nil
}

// Error is a refusal for the frontend to translate by Code
// (ERR_DOCUMENT_DRAFT_PENDING …; "not-signed-in" before any request).
// SignIn: the sign-in is missing or ran out; sign in and try again.
type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Params  map[string]string `json:"params"`
	SignIn  bool              `json:"signIn"`
}

func (s *Service) refusal(err error) *Error {
	if err == nil {
		return nil
	}
	e := &Error{Message: err.Error(), Params: map[string]string{}}
	var ae *api.Error
	if errors.As(err, &ae) {
		e.Code, e.Message = ae.Code, ae.Message
		for k := range ae.Params {
			e.Params[k] = ae.Param(k)
		}
	}
	if api.IsUnauthorized(err) {
		// The token expired or was revoked: forget it, the menu follows.
		e.SignIn = true
		s.endLogin()
		emit("sensormonitor-login", LoginResult{Error: "session-ended"})
	}
	return e
}

// signedInClient is the client with a valid sign-in, or the refusal.
func (s *Service) signedInClient() (*api.Client, *Error) {
	if s.signedIn() == nil {
		return nil, &Error{Code: "not-signed-in", SignIn: true, Params: map[string]string{}}
	}
	c, err := s.client()
	if err != nil {
		return nil, &Error{Message: err.Error(), Params: map[string]string{}}
	}
	return c, nil
}

// StandResult answers Lookup.
type StandResult struct {
	Stand *api.Stand `json:"stand"`
	Error *Error     `json:"error"`
}

// Lookup returns the approved version and the waiting draft of document
// id ref.
func (s *Service) Lookup(ref string) StandResult {
	c, e := s.signedInClient()
	if e != nil {
		return StandResult{Error: e}
	}
	st, err := c.ByRef(context.Background(), ref)
	return StandResult{Stand: st, Error: s.refusal(err)}
}

// SourceResult answers Source.
type SourceResult struct {
	Text  string `json:"text"`
	Error *Error `json:"error"`
}

// Source returns the text of version id to edit (front matter from the
// document's data, no signature).
func (s *Service) Source(id string) SourceResult {
	c, e := s.signedInClient()
	if e != nil {
		return SourceResult{Error: e}
	}
	text, err := c.Source(context.Background(), id)
	return SourceResult{Text: text, Error: s.refusal(err)}
}

// CompareResult answers Compare.
type CompareResult struct {
	Comparison *api.Comparison `json:"comparison"`
	Error      *Error          `json:"error"`
}

// Compare compares md with the stored versions of its doc_id. Nothing is
// saved.
func (s *Service) Compare(ref, md string) CompareResult {
	c, e := s.signedInClient()
	if e != nil {
		return CompareResult{Error: e}
	}
	r, err := c.Compare(context.Background(), ref, md)
	return CompareResult{Comparison: r, Error: s.refusal(err)}
}

// DraftResult answers SaveDraft.
type DraftResult struct {
	Draft *api.SavedDraft `json:"draft"`
	Error *Error          `json:"error"`
}

// SaveDraft stores md as a draft (never approved): a new version of the
// approved document with its doc_id, else a new document. reason and pin
// only when the organisation asks for them.
func (s *Service) SaveDraft(md, filename, reason, pin string) DraftResult {
	c, e := s.signedInClient()
	if e != nil {
		return DraftResult{Error: e}
	}
	d, err := c.SaveDraft(context.Background(), md, api.DraftOptions{Filename: filename, Reason: reason, PIN: pin})
	return DraftResult{Draft: d, Error: s.refusal(err)}
}
