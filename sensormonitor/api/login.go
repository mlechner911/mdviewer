package sensormonitor

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// DeviceCode is what the user sees and confirms in the browser.
type DeviceCode struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// Login is the result of a confirmed device login. The token is valid
// until Expires and cannot be renewed: then sign in again.
type Login struct {
	Token   string
	Email   string
	Expires time.Time
}

// The user denied the sign-in, or waited too long.
var (
	ErrLoginDenied  = errors.New("the sign-in was denied")
	ErrLoginExpired = errors.New("the code expired")
)

// StartLogin begins the device login; clientName appears on the
// confirmation page ("MarkSafe on <computer>").
func (c *Client) StartLogin(ctx context.Context, clientName string) (*DeviceCode, error) {
	var dc DeviceCode
	if err := c.do(ctx, http.MethodPost, "/mlcauth/device/code", map[string]string{"client_name": clientName}, &dc, nil); err != nil {
		return nil, err
	}
	return &dc, nil
}

// WaitLogin polls until the code is confirmed, denied or expired. Cancel
// through ctx.
func (c *Client) WaitLogin(ctx context.Context, dc *DeviceCode) (*Login, error) {
	minPoll := c.MinPoll
	if minPoll <= 0 {
		minPoll = 5 * time.Second
	}
	interval := max(time.Duration(dc.Interval)*time.Second, minPoll)
	deadline := time.Now().Add(time.Duration(dc.ExpiresIn) * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
		var tok struct {
			AccessToken string `json:"access_token"`
			ExpiresIn   int    `json:"expires_in"`
			Email       string `json:"email"`
		}
		err := c.do(ctx, http.MethodPost, "/mlcauth/device/token", map[string]string{"device_code": dc.DeviceCode}, &tok, nil)
		if err == nil {
			l := &Login{Token: tok.AccessToken, Email: tok.Email}
			if tok.ExpiresIn > 0 {
				l.Expires = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
			}
			return l, nil
		}
		switch ErrorCode(err) {
		case "authorization_pending":
		case "slow_down":
			interval += minPoll
		case "access_denied":
			return nil, ErrLoginDenied
		case "expired_token":
			return nil, ErrLoginExpired
		default:
			return nil, err
		}
	}
	return nil, ErrLoginExpired
}
