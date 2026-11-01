package cfg

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

const configFile = "stellation.toml"

type Config struct {
	Name        string          `toml:"name"`
	Slug        string          `toml:"slug,omitempty"`
	Path        string          `toml:"path"`
	Description string          `toml:"description,omitempty"`
	Website     string          `toml:"website,omitempty"`
	AutoJoin    string          `toml:"auto_join,omitempty"`
	SourceURL   string          `toml:"source_url,omitempty"`
	Nix         string          `toml:"nix,omitempty"`
	Splashes    []string        `toml:"splashes,omitempty"`
	Launcher    LauncherConfig  `toml:"launcher"`
	Downloads   DownloadsConfig `toml:"downloads"`
	Deploy      DeployConfig    `toml:"deploy"`
}

type LauncherConfig struct {
	BundleID    string `toml:"bundle_id,omitempty"`
	MinMemoryMB int    `toml:"min_memory_mb"`
	MaxMemoryMB int    `toml:"max_memory_mb"`
}

type DownloadsConfig struct {
	URL    string `toml:"url,omitempty"`
	Dir    string `toml:"dir,omitempty"`
	GitHub string `toml:"github,omitempty"`
}

type DeployConfig struct {
	Remote   string `toml:"remote,omitempty"`
	PackDir  string `toml:"pack_dir,omitempty"`
	Checksum bool   `toml:"checksum,omitempty"`
}

const (
	defaultPath         = "./pack"
	defaultPackDir      = "pack"
	defaultDownloadsDir = "downloads"
	defaultMinMemoryMB  = 2048
	defaultMaxMemoryMB  = 4096
)

var (
	slugPattern     = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	repoPattern     = regexp.MustCompile(`^[A-Za-z0-9-]+/[A-Za-z0-9._-]+$`)
	bundleIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)+$`)
	nonSlug         = regexp.MustCompile(`[^a-z0-9]+`)
	nonBundle       = regexp.MustCompile(`[^A-Za-z0-9-]+`)
)

func DefaultConfig() Config {
	return Config{
		Name: "My Modpack",
		Path: defaultPath,
		Launcher: LauncherConfig{
			MinMemoryMB: defaultMinMemoryMB,
			MaxMemoryMB: defaultMaxMemoryMB,
		},
	}
}

func Exists() bool {
	_, err := os.Stat(configFile)
	return err == nil
}

func SaveConfig(conf Config) error {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(conf); err != nil {
		return fmt.Errorf("encode %s: %w", configFile, err)
	}
	if err := os.WriteFile(configFile, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", configFile, err)
	}
	return nil
}

func GetConfig() (Config, error) {
	var conf Config
	meta, err := toml.DecodeFile(configFile, &conf)
	if errors.Is(err, os.ErrNotExist) {
		return conf, fmt.Errorf("no %s here: run 'stellation init' in your modpack's root directory", configFile)
	}
	if err != nil {
		return conf, fmt.Errorf("read %s: %w", configFile, err)
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, len(undecoded))
		for i, k := range undecoded {
			keys[i] = k.String()
		}
		return conf, fmt.Errorf("%s: unknown keys: %s", configFile, strings.Join(keys, ", "))
	}
	conf.applyDefaults()
	if err := conf.Validate(); err != nil {
		return conf, fmt.Errorf("%s: %w", configFile, err)
	}
	return conf, nil
}

func (c *Config) applyDefaults() {
	if c.Slug == "" {
		c.Slug = Slugify(c.Name)
	}
	if c.Path == "" {
		c.Path = defaultPath
	}
	if c.Deploy.PackDir == "" {
		c.Deploy.PackDir = defaultPackDir
	}
	if c.Downloads.Dir == "" {
		c.Downloads.Dir = defaultDownloadsDir
	}
	if c.Launcher.MinMemoryMB == 0 {
		c.Launcher.MinMemoryMB = defaultMinMemoryMB
	}
	if c.Launcher.MaxMemoryMB == 0 {
		c.Launcher.MaxMemoryMB = max(defaultMaxMemoryMB, c.Launcher.MinMemoryMB)
	}
	if c.Launcher.BundleID == "" && c.Website != "" {
		if u, err := url.Parse(c.Website); err == nil && u.Hostname() != "" {
			labels := strings.Split(u.Hostname(), ".")
			slices.Reverse(labels)
			for i, l := range labels {
				labels[i] = nonBundle.ReplaceAllString(l, "-")
			}
			c.Launcher.BundleID = strings.Join(append(labels, c.Slug), ".")
		}
	}
}

func (c Config) Validate() error {
	var errs []error
	bad := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }

	if strings.TrimSpace(c.Name) == "" {
		bad("name is required")
	}
	if !slugPattern.MatchString(c.Slug) {
		bad("slug %q must be lowercase letters, digits and single dashes (set it explicitly if name has none)", c.Slug)
	}
	urls := [][2]string{{"source_url", c.SourceURL}, {"downloads.url", c.Downloads.URL}}
	if c.Website != "" {
		urls = append(urls, [2]string{"website", c.Website})
	}
	for _, u := range urls {
		if u[1] == "" && u[0] != "website" {
			continue
		}
		if err := checkURL(u[1]); err != nil {
			bad("%s: %v", u[0], err)
		}
	}
	if c.AutoJoin != "" && strings.ContainsAny(c.AutoJoin, " \t/") {
		bad("auto_join %q must be a server address", c.AutoJoin)
	}
	if c.Downloads.GitHub != "" {
		if !repoPattern.MatchString(c.Downloads.GitHub) {
			bad(`downloads.github %q must be like "owner/name"`, c.Downloads.GitHub)
		}
		if c.Downloads.URL != "" {
			bad("downloads.github and downloads.url can't both be set")
		}
	}
	if c.Launcher.MinMemoryMB <= 0 || c.Launcher.MaxMemoryMB < c.Launcher.MinMemoryMB {
		bad("launcher memory: need 0 < min_memory_mb (%d) <= max_memory_mb (%d)", c.Launcher.MinMemoryMB, c.Launcher.MaxMemoryMB)
	}
	if c.Launcher.BundleID != "" && !bundleIDPattern.MatchString(c.Launcher.BundleID) {
		bad("launcher.bundle_id %q must be reverse-DNS, like net.example.modpack", c.Launcher.BundleID)
	}
	for _, d := range [][2]string{{"deploy.pack_dir", c.Deploy.PackDir}, {"downloads.dir", c.Downloads.Dir}} {
		if strings.Contains(d[1], "..") || strings.HasPrefix(d[1], "/") {
			bad("%s %q must be a relative path inside the remote", d[0], d[1])
		}
	}
	return errors.Join(errs...)
}

func checkURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return fmt.Errorf("%q must be an absolute http(s) URL", raw)
	}
	return nil
}

func Slugify(name string) string {
	return strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(name), "-"), "-")
}

func (c Config) website() (string, error) {
	if c.Website == "" {
		return "", errors.New("website is not set in " + configFile)
	}
	return strings.TrimSuffix(c.Website, "/"), nil
}

func (c Config) PackURL() (string, error) {
	site, err := c.website()
	if err != nil {
		return "", err
	}
	return site + "/" + strings.Trim(c.Deploy.PackDir, "/") + "/pack.toml", nil
}

func (c Config) UsesGitHubReleases() bool {
	return c.Downloads.GitHub != ""
}

func (c Config) DownloadsURL() (string, error) {
	switch {
	case c.UsesGitHubReleases():
		return "https://github.com/" + c.Downloads.GitHub + "/releases/latest/download", nil
	case c.Downloads.URL != "":
		return strings.TrimSuffix(c.Downloads.URL, "/"), nil
	}
	site, err := c.website()
	if err != nil {
		return "", fmt.Errorf("%w (or set downloads.url or downloads.github)", err)
	}
	return site + "/" + strings.Trim(c.Downloads.Dir, "/"), nil
}
