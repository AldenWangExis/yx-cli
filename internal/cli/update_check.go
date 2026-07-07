package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	updateCheckOptOutEnv = "YX_NO_UPDATE_CHECK"
	updateCheckLatestURL = "https://registry.npmjs.org/@aldenwangexis%2fyx-cli/latest"
	defaultNPMPackage    = "@aldenwangexis/yx-cli"
)

type UpdateChecker interface {
	Check(ctx UpdateCheckContext) (UpdateNotice, error)
}

type UpdateCheckContext struct {
	Context context.Context
}

type UpdateNotice struct {
	Available bool
	Current   string
	Latest    string
	Command   string
}

type UpdateCheckerConfig struct {
	LatestURL     string
	CachePath     string
	Current       string
	Now           func() time.Time
	HTTPClient    *http.Client
	CacheDuration time.Duration
}

type RegistryUpdateChecker struct {
	config UpdateCheckerConfig
}

func NewRegistryUpdateChecker(config UpdateCheckerConfig) *RegistryUpdateChecker {
	if config.LatestURL == "" {
		config.LatestURL = updateCheckLatestURL
	}
	if config.Current == "" {
		config.Current = Version
	}
	if config.Now == nil {
		config.Now = time.Now
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 1500 * time.Millisecond}
	}
	if config.CacheDuration == 0 {
		config.CacheDuration = 24 * time.Hour
	}
	if config.CachePath == "" {
		config.CachePath = defaultUpdateCachePath()
	}
	return &RegistryUpdateChecker{config: config}
}

func NewGitHubUpdateChecker(config UpdateCheckerConfig) *RegistryUpdateChecker {
	return NewRegistryUpdateChecker(config)
}

func (c *RegistryUpdateChecker) Check(ctx UpdateCheckContext) (UpdateNotice, error) {
	latest, ok := c.cachedLatest()
	if !ok {
		var err error
		latest, err = c.fetchLatest(ctx.Context)
		if err != nil {
			return UpdateNotice{}, err
		}
		_ = c.writeCache(latest)
	}
	if !isNewerVersion(c.config.Current, latest) {
		return UpdateNotice{}, nil
	}
	return UpdateNotice{
		Available: true,
		Current:   c.config.Current,
		Latest:    latest,
		Command:   updateCommand(latest),
	}, nil
}

func updateCommand(latest string) string {
	if os.Getenv("YX_INSTALL_CHANNEL") == "npm" {
		pkg := strings.TrimSpace(os.Getenv("YX_NPM_PACKAGE"))
		if pkg == "" {
			pkg = defaultNPMPackage
		}
		return fmt.Sprintf("npm update -g %s", pkg)
	}
	return fmt.Sprintf("YX_INSTALL_VERSION=%s curl -fsSL https://raw.githubusercontent.com/AldenWangExis/yx-cli/master/scripts/install.sh | sh", latest)
}

func (c *RegistryUpdateChecker) cachedLatest() (string, bool) {
	data, err := os.ReadFile(c.config.CachePath)
	if err != nil {
		return "", false
	}
	var cache struct {
		CheckedAt time.Time `json:"checkedAt"`
		Latest    string    `json:"latest"`
	}
	if err := json.Unmarshal(data, &cache); err != nil || cache.Latest == "" {
		return "", false
	}
	if c.config.Now().Sub(cache.CheckedAt) > c.config.CacheDuration {
		return "", false
	}
	return cache.Latest, true
}

func (c *RegistryUpdateChecker) writeCache(latest string) error {
	if err := os.MkdirAll(filepath.Dir(c.config.CachePath), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(struct {
		CheckedAt time.Time `json:"checkedAt"`
		Latest    string    `json:"latest"`
	}{CheckedAt: c.config.Now().UTC(), Latest: latest})
	if err != nil {
		return err
	}
	return os.WriteFile(c.config.CachePath, data, 0o600)
}

func (c *RegistryUpdateChecker) fetchLatest(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.LatestURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "yx-cli/"+c.config.Current)
	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("latest version lookup returned status %d", resp.StatusCode)
	}
	var payload struct {
		TagName string `json:"tag_name"`
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	return normalizeVersionTag(firstStatusValue(payload.Version, payload.TagName, payload.Name)), nil
}

func normalizeVersionTag(version string) string {
	value := strings.TrimSpace(version)
	if value == "" || strings.HasPrefix(value, "v") {
		return value
	}
	return "v" + value
}

func maybeRunUpdateCheck(cmdCtx context.Context, cmdErrWriter interface{ Write([]byte) (int, error) }, opts Options, ctx Context, commandName string) {
	if shouldSkipUpdateCheck(ctx, commandName) {
		return
	}
	checker := opts.UpdateChecker
	if checker == nil {
		checker = NewRegistryUpdateChecker(UpdateCheckerConfig{})
	}
	notice, err := checker.Check(UpdateCheckContext{Context: cmdCtx})
	if err != nil || !notice.Available {
		return
	}
	_, _ = fmt.Fprintf(cmdErrWriter, "A newer yx is available: %s -> %s\nUpdate: %s\n", notice.Current, notice.Latest, notice.Command)
}

func shouldSkipUpdateCheck(ctx Context, commandName string) bool {
	if os.Getenv(updateCheckOptOutEnv) != "" || os.Getenv("CI") != "" {
		return true
	}
	if ctx.JSON {
		return true
	}
	switch commandName {
	case "completion", "help":
		return true
	default:
		return false
	}
}

func defaultUpdateCachePath() string {
	dir, err := os.UserCacheDir()
	if err != nil || dir == "" {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil || home == "" {
			return filepath.Join(os.TempDir(), "yx", "update-check.json")
		}
		dir = filepath.Join(home, ".cache")
	}
	return filepath.Join(dir, "yx", "update-check.json")
}

func isNewerVersion(current, latest string) bool {
	currentParts, ok := parseSemver(current)
	if !ok {
		return false
	}
	latestParts, ok := parseSemver(latest)
	if !ok {
		return false
	}
	for i := range currentParts {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}
	return false
}

func parseSemver(version string) ([3]int, bool) {
	var parts [3]int
	value := strings.TrimPrefix(strings.TrimSpace(version), "v")
	fields := strings.Split(value, ".")
	if len(fields) != 3 {
		return parts, false
	}
	for i, field := range fields {
		if idx := strings.IndexAny(field, "-+"); idx >= 0 {
			field = field[:idx]
		}
		n, err := strconv.Atoi(field)
		if err != nil || n < 0 {
			return parts, false
		}
		parts[i] = n
	}
	return parts, true
}
