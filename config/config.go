// Package file provides the capability to parse from disk a configuration file.
package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ardnew/ctoldup/log"
	"github.com/ardnew/roster"

	"github.com/mholt/archiver/v3"
	"github.com/otiai10/copy"
	"gopkg.in/yaml.v3"
)

type (
	DirectoryNotFoundError string
	InvalidPathError       string
	NotRegularFileError    string
	FileExistsError        string
)

// Error returns the error message for DirectoryNotFoundError.
func (e DirectoryNotFoundError) Error() string {
	return "directory not found: " + string(e)
}

// Error returns the error message for InvalidPathError.
func (e InvalidPathError) Error() string {
	return "invalid file path: " + string(e)
}

// Error returns the error message for NotRegularFileError.
func (e NotRegularFileError) Error() string {
	return "not a regular file: " + string(e)
}

// Error returns the error message for FileExistsError.
func (e FileExistsError) Error() string {
	return "file already exists: " + string(e)
}

const (
	CtoldPathToken = "${CTOLD}"
	CtoldTagToken  = "${CTOLD.TAG}"
	CtoldLastToken = "${CTOLD.REV}"
	DateTimeToken  = "${DATETIME.NOW}"
)

const (
	//CtoldRepoDefault  = "http://rstok3-dev02:3690/svn/FSAS_CTOLD_Source"
	CtoldRepoDefault  = "https://github.com/ardnew/ctoldup"
	CtoldTagDefault   = "trunk"
	CtoldLocalDefault = ".ctoldup"
	CtoldLastDefault  = ""
)

// Permissions defines the default permissions of config files written to disk.
var Permissions os.FileMode = 0600

// Config represents a configuration file, containing the CTOLD and destination
// settings.
type Config struct {
	path     string
	Ctold    CtoldConfig `yaml:"ctold"`
	Merge    MergeMap    `yaml:"merge"`
	Compress CompressMap `yaml:"compress"`
}

type CtoldConfig struct {
	path  string
	Repo  string `yaml:"repo"`
	Tag   string `yaml:"tag"`
	Local string `yaml:"local"`
	Last  string `yaml:"last"`
}

func (c *CtoldConfig) SetPath(path string) {
	c.path = path
}

func (c *CtoldConfig) Url() string {
	return fmt.Sprintf("%s/%s", c.Repo, c.Tag)
}

func (c *CtoldConfig) Wc() string {
	return filepath.Clean(filepath.Join(c.Local, c.Tag))
}

func (c *CtoldConfig) LastValid() bool {
	return c.Last != CtoldLastDefault
}

type MergeMap map[string]MergeConfig

type MergeConfig struct {
	Into   string `yaml:"into"`
	Roster bool   `yaml:"roster"`
}

type CompressMap map[string]CompressConfig

type CompressConfig struct {
	Path      string `yaml:"path"`
	Overwrite bool   `yaml:"overwrite"`
	Method    string `yaml:"method"`
	Level     int    `yaml:"level"`
}

// New constructs a new config file at the given file path, initialized with all
// default data.
// The returned file is stored in-memory only. The Write method must be called
// to write the file to disk.
func New(filePath string) *Config {
	return &Config{
		path: filePath,
		Ctold: CtoldConfig{
			path:  "",
			Repo:  CtoldRepoDefault,
			Tag:   CtoldTagDefault,
			Local: CtoldLocalDefault,
			Last:  CtoldLastDefault,
		},
		Merge: MergeMap{
			CtoldPathToken: MergeConfig{
				Into:   "",
				Roster: true,
			},
		},
		Compress: CompressMap{
			CtoldPathToken: CompressConfig{
				Path: fmt.Sprintf("ctold-%s-r%s-%s.zip",
					CtoldTagToken, CtoldLastToken, DateTimeToken),
				Overwrite: true,
				Method:    "zip",
				Level:     9,
			},
		},
	}
}

// Parse parses the configuration file into the returned Config struct, or
// returns a Config struct with default configuration if the configuration file
// does not exist.
// Returns a nil Config and descriptive error if the given path is invalid.
func Parse(init bool, filePath string) (*Config, error) {

	dir := filepath.Dir(filePath)
	dstat, derr := os.Stat(dir)
	if os.IsNotExist(derr) {
		return nil, DirectoryNotFoundError(dir)
	} else if !dstat.IsDir() {
		return nil, InvalidPathError(dir)
	}

	fstat, ferr := os.Stat(filePath)
	if os.IsNotExist(ferr) {
		return New(filePath), nil
	} else if uint32(fstat.Mode()&os.ModeType) != 0 {
		return nil, NotRegularFileError(filePath)
	}

	if init {
		return nil, FileExistsError(filePath)
	}

	log.Msg(log.Info, "parse", "%q", filePath)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	cfg := New(filePath)
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Write formats and writes the receiver Config cfg's configuration to disk.
// Returns an error if formatting or writing fails.
func (cfg *Config) Write() error {
	data, err := yaml.Marshal(cfg)
	if nil != err {
		return err
	}
	return ioutil.WriteFile(cfg.path, data, Permissions)
}

func (cfg *Config) ReplaceTokens(str string) string {
	sub := map[string]string{
		CtoldPathToken: cfg.Ctold.path,
		CtoldTagToken:  cfg.Ctold.Tag,
		CtoldLastToken: cfg.Ctold.Last,
		DateTimeToken:  time.Now().Local().Format("20060102-150405"),
	}
	for from, to := range sub {
		str = strings.ReplaceAll(str, from, to)
	}
	return str
}

func (cfg *Config) MergeAll() error {
	for src, merge := range cfg.Merge {
		from := cfg.ReplaceTokens(src)
		dest := cfg.ReplaceTokens(merge.Into)
		if dest == "" {
			return fmt.Errorf("merge destination undefined: %q", from)
		}
		_, err := os.Stat(dest)
		if nil != err {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(dest, os.ModePerm); nil != err {
					return err
				}
			} else {
				return err
			}
		}
		log.Msg(log.Info, "copy", "%q -> %q", from, dest)
		if err := copy.Copy(from, dest, copy.Options{
			OnSymlink: func(s string) copy.SymlinkAction {
				return copy.Skip
			},
			Skip: func(s string) (bool, error) {
				return filepath.Base(s) == ".svn", nil
			},
			Sync: true,
		}); nil != err {
			return err
		}
		if merge.Roster {
			rosterFile := filepath.Join(dest, ".roster.yml")
			log.Msg(log.Info, "roster", "%q -> %q", dest, rosterFile)
			if err := roster.Take(roster.SkipTaker, ".roster.yml", true, dest); nil != err {
				return err
			}
		}
	}
	return nil
}

func (cfg *Config) CompressAll() error {
	for src, compress := range cfg.Compress {
		from := cfg.ReplaceTokens(src)
		dest := cfg.ReplaceTokens(compress.Path)
		switch compress.Method {
		case "zip":
			arc := archiver.Zip{
				CompressionLevel:       compress.Level,
				OverwriteExisting:      compress.Overwrite,
				MkdirAll:               true,
				SelectiveCompression:   true,
				ImplicitTopLevelFolder: false,
				ContinueOnError:        false,
			}
			if nil != arc.CheckExt(dest) {
				dest = fmt.Sprintf("%s.zip", dest)
			}
			log.Msg(log.Info, "zip", "%q -> %q", from, dest)
			if err := arc.Archive([]string{from}, dest); nil != err {
				return err
			}
		}
	}
	return nil
}
