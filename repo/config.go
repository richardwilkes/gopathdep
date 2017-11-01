package repo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v2"

	"github.com/richardwilkes/gopathdep/util"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
)

const (
	// CurrentVersion of the configuration file.
	CurrentVersion = "1.0"
	// ConfigFileName is the configuration file name.
	ConfigFileName = "pathdep.yaml"
)

// Config holds the configuration.
type Config struct {
	Dir          string `yaml:"-"`
	Version      string
	Dependencies Dependencies
}

// NewConfigFromDir creates a new configuration from the configuration file in the directory.
func NewConfigFromDir(dir string) (*Config, error) {
	cfg := &Config{Dir: util.MustGitRootOrDir(dir)}
	path := filepath.ToSlash(filepath.Join(cfg.Dir, ConfigFileName))

	file, err := os.Open(path)
	if err == nil {
		var data []byte
		if data, err = ioutil.ReadAll(file); err == nil {
			err = yaml.Unmarshal(data, cfg)
		}
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		err = errs.Wrap(err)
	} else {
		const msg = "Unable to open %s\nTry running '%s record' to create one." // Just here to fool the linter, as I really do want an error message with punctuation.
		err = fmt.Errorf(msg, path, cmdline.AppCmdName)
	}
	return cfg, err
}

// Save the configuration file.
func (cfg *Config) Save() error {
	cfg.Version = CurrentVersion
	sort.Sort(cfg.Dependencies)
	file, err := os.Create(filepath.Join(cfg.Dir, ConfigFileName))
	if err == nil {
		var data []byte
		if data, err = yaml.Marshal(cfg); err == nil {
			_, err = file.Write(data)
		}
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return errs.Wrap(err)
}
