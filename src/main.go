package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Config mirrors config.json exactly.
type Config struct {
	BaseURL              string   `json:"BASE_URL"`
	APIToken             string   `json:"API_TOKEN"`
	TargetClientPrefixes []string `json:"TARGET_CLIENT_PREFIXES"`
	GroupIDMode          bool     `json:"GROUP_ID_MODE"`
	TargetGroupPrefixes  []string `json:"TARGET_GROUP_PREFIXES"`
	TargetGroupIDs       []int    `json:"TARGET_GROUP_IDS"`
}

// app bundles information that needs to be used by many functions
type app struct {
	http           *http.Client
	cfg            Config
	sid            string
	targetGroupIDs []int
}

func (a *app) headers() map[string]string {
	return map[string]string{
		"X-FTL-SID":    a.sid,
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}
}

// doRequest is used as a helper for sending HTTP requests to Pi-Hole
func (a *app) doRequest(method, path string, body any) ([]byte, int, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshaling request body: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, a.cfg.BaseURL+path, reader)
	if err != nil {
		return nil, 0, err
	}
	for k, v := range a.headers() {
		req.Header.Set(k, v)
	}

	resp, err := a.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s %s failed: %w", method, path, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode >= 400 {
		return raw, resp.StatusCode, fmt.Errorf("%s %s returned %d: %s", method, path, resp.StatusCode, string(raw))
	}

	return raw, resp.StatusCode, nil
}

// getAuth authenticates with Pi-hole and returns the session ID.
func (a *app) getAuth() (string, error) {
	raw, _, err := a.doRequest(http.MethodPost, "/auth", map[string]string{"password": a.cfg.APIToken})
	if err != nil {
		return "", err
	}

	var parsed struct {
		Session struct {
			SID string `json:"sid"`
		} `json:"session"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("parsing auth response: %w", err)
	}
	return parsed.Session.SID, nil
}

// getAllClients fetches every known client
func (a *app) getAllClients() ([]map[string]any, error) {
	raw, _, err := a.doRequest(http.MethodGet, "/clients", nil)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Clients []map[string]any `json:"clients"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parsing clients response: %w", err)
	}
	return parsed.Clients, nil
}

// getGroups returns a name id map of all groups
func (a *app) getGroups() (map[string]int, error) {
	raw, _, err := a.doRequest(http.MethodGet, "/groups", nil)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Groups []struct {
			Name string `json:"name"`
			ID   int    `json:"id"`
		} `json:"groups"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parsing groups response: %w", err)
	}

	result := make(map[string]int, len(parsed.Groups))
	for _, g := range parsed.Groups {
		result[g.Name] = g.ID
	}
	return result, nil
}

// nameMatchesPrefix checks whether a client's comment starts with any of the configured target prefixes
func (a *app) nameMatchesPrefix(client map[string]any) bool {
	comment, _ := client["comment"].(string)
	for _, prefix := range a.cfg.TargetClientPrefixes {
		if strings.HasPrefix(comment, prefix) {
			return true
		}
	}
	return false
}

// resetClientsWithPrefixes finds every client whose comment matches a target prefix and resets its groups
func (a *app) resetClientsWithPrefixes() error {
	clients, err := a.getAllClients()
	if err != nil {
		return err
	}

	var matched []map[string]any
	for _, c := range clients {
		if a.nameMatchesPrefix(c) {
			matched = append(matched, c)
		} else {
			fmt.Printf("Skipping %v\n", c["comment"])
		}
	}

	for _, c := range matched {
		clientID, _ := c["client"].(string)
		payload := map[string]any{
			"client":  clientID,
			"name":    c["name"],
			"comment": c["comment"],
			"groups":  a.targetGroupIDs,
		}
		if _, _, err := a.doRequest(http.MethodPut, "/clients/"+clientID, payload); err != nil {
			return err
		}
	}

	fmt.Println("Clients with matching prefixes have been reset.")
	return nil
}

func loadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("opening config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

// prefixSliceStartsWith checks prefixes against client comments
func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func run() error {
	cfg, err := loadConfig("config.json")
	if err != nil {
		return err
	}

	if cfg.BaseURL == "" || cfg.APIToken == "" {
		fmt.Println("You must provide your BASE_URL and API_TOKEN.")
		os.Exit(0)
	}

	a := &app{
		http: &http.Client{},
		cfg:  cfg,
	}

	sid, err := a.getAuth()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	a.sid = sid

	if cfg.GroupIDMode {
		a.targetGroupIDs = cfg.TargetGroupIDs
	} else {
		groupNameDict, err := a.getGroups()
		if err != nil {
			return err
		}
		var ids []int
		for name, id := range groupNameDict {
			if hasAnyPrefix(name, cfg.TargetGroupPrefixes) {
				ids = append(ids, id)
			}
		}
		a.targetGroupIDs = ids
	}

	if err := a.resetClientsWithPrefixes(); err != nil {
		return err
	}

	// API session close
	if _, _, err := a.doRequest(http.MethodDelete, "/auth/", nil); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
