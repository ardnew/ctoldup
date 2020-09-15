package main

import (
	"flag"
	"os"

	"github.com/ardnew/ctoldup/config"
	"github.com/ardnew/ctoldup/log"
	"github.com/ardnew/ctoldup/repo"
	"github.com/ardnew/version"
)

func init() {
	version.ChangeLog = []version.Change{
		{
			Package: "ctoldup",
			Version: "0.1.0",
			Date:    "Sept 10, 2020",
			Description: []string{
				`initial commit`,
			},
		},
	}
}

const (
	configFilePathDefault = "ctoldup.yml"
	initConfigFileDefault = false
)

func main() {

	var (
		configFilePath string
		initConfigFile bool
		copySource     bool
		copyForce      bool
	)

	flag.StringVar(&configFilePath, "f", configFilePathDefault, "configuration file path")
	flag.BoolVar(&initConfigFile, "n", initConfigFileDefault, "initialize config file with default settings")
	flag.BoolVar(&copySource, "c", false, "if repo updated, create copy (and optional zip package)")
	flag.BoolVar(&copyForce, "d", false, "force create copy (and optional zip package)")
	flag.Parse()

	if initConfigFile {
		log.Msg(log.Info, "init", "%q", configFilePath)
	}

	cfg, err := config.Parse(initConfigFile, configFilePath)
	if nil != err {
		log.Msg(log.Error, "error", "config.Parse(): %s", err.Error())
		os.Exit(1)
	}

	write := func(c *config.Config) {
		if err := c.Write(); nil != err {
			log.Msg(log.Error, "error", "%s", err)
			os.Exit(2)
		}
	}

	if initConfigFile {
		write(cfg)

	} else {
		svn, err := repo.New(cfg)
		if nil != err {
			log.Msg(log.Error, "error", "repo.New(): %s", err.Error())
			os.Exit(3)
		}

		ver, err := svn.Fetch()
		if nil != err {
			log.Msg(log.Error, "error", "svn.Fetch(): %s", err.Error())
			os.Exit(4)
		}

		updated := (ver != cfg.Ctold.Last) || !cfg.Ctold.LastValid()
		if updated {
			if cfg.Ctold.LastValid() {
				log.Msg(log.Info, "revision", "%s -> %s", cfg.Ctold.Last, ver)
			} else {
				log.Msg(log.Info, "revision", "%s", ver)
			}
			cfg.Ctold.Last = ver
			write(cfg)
		} else {
			log.Msg(log.Info, "revision", "%s (no change)", ver)
		}

		if copyForce || (copySource && updated) {
			if "" == cfg.Copy.Dest {
				log.Msg(log.Error, "error", "invalid copy destination: (empty)")
				os.Exit(5)
			}

			if err := svn.Copy(cfg.Copy.Dest, cfg.Copy.Zip, cfg.Copy.Level); nil != err {
				log.Msg(log.Error, "error", "svn.Copy(): %s", err.Error())
				os.Exit(6)
			}
		}
	}

	log.Msg(log.Info, "exit", "success!")
}
