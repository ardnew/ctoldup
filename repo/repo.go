package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ardnew/ctoldup/config"
	"github.com/ardnew/ctoldup/log"

	"github.com/Masterminds/vcs"
	"github.com/mholt/archiver/v3"
	"github.com/otiai10/copy"
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

func (r *Repo) Fetch() (version string, err error) {
	log.Msg(log.Info, "ping", "%q", r.Remote())
	if !r.Ping() {
		return "", fmt.Errorf("cannot connect to repository: %s", r.Remote())
	}
	if r.CheckLocal() {
		log.Msg(log.Info, "update", "%q -> %q", r.Remote(), r.LocalPath())
		if err := r.Update(); nil != err {
			return "", err
		}
	} else {
		log.Msg(log.Info, "checkout", "%q -> %q", r.Remote(), r.LocalPath())
		if err := r.Get(); nil != err {
			return "", err
		}
	}
	version, err = r.Version()
	return
}

func (r *Repo) Copy(dest string, zip string, level int) error {
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
	log.Msg(log.Info, "copy", "%q -> %q", r.LocalPath(), dest)
	if err := copy.Copy(r.LocalPath(), dest, copy.Options{
		OnSymlink: func(src string) copy.SymlinkAction {
			return copy.Skip
		},
		Skip: func(src string) (bool, error) {
			return filepath.Base(src) == ".svn", nil
		},
		Sync: true,
	}); nil != err {
		return err
	}
	if "" != zip {
		arc := archiver.Zip{
			CompressionLevel:       level,
			OverwriteExisting:      true,
			MkdirAll:               true,
			SelectiveCompression:   true,
			ImplicitTopLevelFolder: false,
			ContinueOnError:        false,
		}
		if nil != arc.CheckExt(zip) {
			zip = fmt.Sprint("%s.zip", zip)
		}
		log.Msg(log.Info, "zip", "%q -> %q", dest, zip)
		return arc.Archive([]string{dest}, zip)
	}
	return nil
}
