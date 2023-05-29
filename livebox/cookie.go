package livebox

import (
	"net/http"
	"strings"
)

// The Livebox API uses and invalid character ('/') as the session cookie name.
// Since the standard implementation of http.CookieJar (rightfully) discards invalid cookies,
// cookieNamePatcher renames cookies on the fly when receiving responses and sending requests.
type cookieNamePatcher struct {
	transport http.Transport
}

// RoundTrip implements the http.RoundTripper interface.
func (i *cookieNamePatcher) RoundTrip(req *http.Request) (*http.Response, error) {
	const (
		realCookieName    = "/sessid"
		patchedCookieName = "-sessid"
	)

	if strings.Contains(req.Header.Get("Cookie"), patchedCookieName) {
		req.Header["Cookie"][0] = strings.ReplaceAll(req.Header["Cookie"][0], patchedCookieName, realCookieName)
	}

	resp, err := i.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if strings.Contains(resp.Header.Get("Set-Cookie"), realCookieName) {
		resp.Header["Set-Cookie"][0] = strings.ReplaceAll(resp.Header["Set-Cookie"][0], realCookieName, patchedCookieName)
	}

	return resp, nil
}
