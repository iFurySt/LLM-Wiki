package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type storedConfig struct {
	CurrentProfile string                   `json:"current_profile"`
	Profiles       map[string]storedProfile `json:"profiles"`
}

type storedProfile struct {
	BaseURL      string `json:"base_url"`
	TenantID     string `json:"tenant_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
	PrincipalID  string `json:"principal_id"`
	DisplayName  string `json:"display_name"`
}

type resolvedOptions struct {
	BaseURL      string
	TenantID     string
	AccessToken  string
	RefreshToken string
	ProfileName  string
	ExpiresAt    time.Time
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".llm-wiki", "config.json"), nil
}

func loadStoredConfig() (storedConfig, error) {
	path, err := configPath()
	if err != nil {
		return storedConfig{}, err
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return storedConfig{CurrentProfile: "default", Profiles: map[string]storedProfile{}}, nil
		}
		return storedConfig{}, err
	}
	var cfg storedConfig
	if err := json.Unmarshal(payload, &cfg); err != nil {
		return storedConfig{}, err
	}
	if cfg.CurrentProfile == "" {
		cfg.CurrentProfile = "default"
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]storedProfile{}
	}
	return cfg, nil
}

func saveStoredConfig(cfg storedConfig) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o600)
}

func resolveOptions(baseURL string, tenantID string, accessToken string, refreshToken string, profileName string) (resolvedOptions, error) {
	cfg, err := loadStoredConfig()
	if err != nil {
		return resolvedOptions{}, err
	}
	if profileName == "" {
		profileName = firstNonEmpty(os.Getenv("LLM_WIKI_PROFILE"), cfg.CurrentProfile)
	}
	profile := cfg.Profiles[profileName]
	options := resolvedOptions{
		BaseURL:      firstNonEmpty(baseURL, os.Getenv("LLM_WIKI_BASE_URL"), os.Getenv("LLM_WIKI_CLI_BASE_URL"), profile.BaseURL, "https://llm-wiki.ifuryst.com"),
		TenantID:     firstNonEmpty(tenantID, os.Getenv("LLM_WIKI_TENANT"), profile.TenantID, "default"),
		AccessToken:  firstNonEmpty(accessToken, os.Getenv("LLM_WIKI_TOKEN"), profile.AccessToken),
		RefreshToken: firstNonEmpty(refreshToken, os.Getenv("LLM_WIKI_REFRESH_TOKEN"), profile.RefreshToken),
		ProfileName:  profileName,
	}
	if profile.ExpiresAt != "" {
		if parsed, err := time.Parse(time.RFC3339, profile.ExpiresAt); err == nil {
			options.ExpiresAt = parsed
		}
	}
	return options, nil
}

func persistProfile(name string, profile storedProfile) error {
	cfg, err := loadStoredConfig()
	if err != nil {
		return err
	}
	cfg.CurrentProfile = name
	cfg.Profiles[name] = profile
	return saveStoredConfig(cfg)
}

func clearProfileTokens(name string) error {
	cfg, err := loadStoredConfig()
	if err != nil {
		return err
	}
	profile := cfg.Profiles[name]
	profile.AccessToken = ""
	profile.RefreshToken = ""
	profile.ExpiresAt = ""
	profile.PrincipalID = ""
	profile.DisplayName = ""
	cfg.Profiles[name] = profile
	return saveStoredConfig(cfg)
}

func firstNonEmpty(values ...string) string {
	for _, item := range values {
		value := strings.TrimSpace(item)
		if value != "" {
			return value
		}
	}
	return ""
}
