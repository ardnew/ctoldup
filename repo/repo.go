package repo

import (
	"fmt"

	"github.com/ardnew/ctoldup/config"
	"github.com/ardnew/ctoldup/log"

	"github.com/Masterminds/vcs"
)

type Repo struct {
	*vcs.SvnRepo
	cfg config.Config
}

func New(cfg *config.Config) (*Repo, error) {
	svn, err := vcs.NewSvnRepo(cfg.Ctold.Url(), cfg.Ctold.Wc())
	if nil != err {
		return nil, err
	}
	return &Repo{
		SvnRepo: svn,
		cfg:     *cfg,
	}, nil
}

func (r *Repo) Fetch() (version, local string, err error) {
	log.Msg(log.Info, "ping", "%q", r.Remote())
	if !r.Ping() {
		return "", "", fmt.Errorf("cannot connect to repository: %s", r.Remote())
	}
	if r.CheckLocal() {
		log.Msg(log.Info, "update", "%q -> %q", r.Remote(), r.LocalPath())
		if err := r.Update(); nil != err {
			return "", "", err
		}
	} else {
		log.Msg(log.Info, "checkout", "%q -> %q", r.Remote(), r.LocalPath())
		if err := r.Get(); nil != err {
			return "", "", err
		}
	}
	version, err = r.Version()
	return version, r.LocalPath(), err
}
