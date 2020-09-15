// Package file provides the capability to parse from disk a configuration file.
package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ardnew/ctoldup/log"

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
	CtoldRepoDefault    = "http://RSTOK3-DEV02:3690/svn/FSAS_CTOLD_Source"
	CtoldTagDefault     = "trunk"
	CtoldLocalDefault   = ".ctoldup"
	CtoldLastDefault    = ""
	CopyDestDefault     = ""
	CopyZipDefault      = "ctold-trunk-latest.zip"
	CopyZipLevelDefault = 9
)

// Permissions defines the default permissions of config files written to disk.
var Permissions os.FileMode = 0600

// Config represents a configuration file, containing the CTOLD and destination
// settings.
type Config struct {
	path  string
	Ctold CtoldConfig `yaml:"ctold"`
	Copy  CopyConfig  `yaml:"copy"`
}

type CtoldConfig struct {
	Repo  string `yaml:"repo"`
	Tag   string `yaml:"tag"`
	Local string `yaml:"local"`
	Last  string `yaml:"last"`
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

type CopyConfig struct {
	Dest  string `yaml:"dest"`
	Zip   string `yaml:"zip"`
	Level int    `yaml:"compression"`
}

// New constructs a new config file at the given file path, initialized with all
// default data.
// The returned file is stored in-memory only. The Write method must be called
// to write the file to disk.
func New(filePath string) *Config {
	return &Config{
		path: filePath,
		Ctold: CtoldConfig{
			Repo:  CtoldRepoDefault,
			Tag:   CtoldTagDefault,
			Local: CtoldLocalDefault,
			Last:  CtoldLastDefault,
		},
		Copy: CopyConfig{
			Dest:  CopyDestDefault,
			Zip:   CopyZipDefault,
			Level: CopyZipLevelDefault,
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
