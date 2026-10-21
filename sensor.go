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
