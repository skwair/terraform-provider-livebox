package livebox

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
)

// Client used to interact with the Livebox API.
// Note that for now, this client is not designed to be long-lived.
// It uses the same cookie-based session mechanism as the official web interface which keeps active connections for
// roughly 5 minutes.
type Client struct {
	host  string
	token string

	httpClient *http.Client
}

// NewClient returns a new client.
// The host parameter must contain one of the following schemes: http, https.
// Of course, https is strongly recommended even if the Livebox serves a self-signed certificate,
// at least the connection will be encrypted.
func NewClient(host, password string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	c := &Client{
		host: host,
		httpClient: &http.Client{
			Jar: jar,
			Transport: &cookieNamePatcher{
				transport: http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			},
		},
	}

	if err := c.login(password); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) login(password string) error {
	payload := &apiRequest{
		Service: "sah.Device.Information",
		Method:  "createContext",
		Parameters: map[string]any{
			"applicationName": "webui",
			"username":        "admin",
			"password":        password,
		},
	}

	resp, err := c.doAuthReq(payload)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	var lr struct {
		Data struct {
			ContextID string `json:"contextID"`
		} `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return err
	}

	c.token = lr.Data.ContextID

	return nil
}
