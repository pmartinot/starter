package packs

import (
	"fmt"
	"path/filepath"

	"github.com/cloud66/starter/common"
)

type Ruby struct {
	WorkDir string

	Gemfile  string
	Version  string
	Packages *common.Lister
}

func (r *Ruby) Name() string {
	return "ruby"
}

func (r *Ruby) PackVersion() string {
	return "0.1"
}

func (r *Ruby) Detect() (bool, error) {
	r.Gemfile = filepath.Join(r.WorkDir, "Gemfile")

	// TODO: fetch git url and branch from the director
	return common.FileExists(r.Gemfile), nil
}

func (r *Ruby) OutputFolder() string {
	return r.WorkDir
}

func (r *Ruby) Compile() (*common.ParseContext, error) {
	// we have a ruby app

	foundRuby, rubyVersion := common.GetRubyVersion(r.Gemfile)
	if foundRuby {
		r.Version = fmt.Sprintf("%s-onbuild", rubyVersion)
	} else {
		r.Version = "onbuild"
	}

	service := &common.Service{Name: "web"}

	// port depends on the application server. for now we are going to fix to 3000
	if runsUnicorn, _ := common.GetGemVersion(r.Gemfile, "unicorn", "thin"); runsUnicorn {
		fmt.Println(common.MsgL2, "----> Found non Webrick application server", common.MsgReset)
		// The command here will be found in the Procfile
		service.Ports = []string{"9292:80:443"}
	} else {
		service.Ports = []string{"3000:80:443"}
		service.Command = "bundle exec rails s _env:$RAILS_ENV"
	}

	// add packages based on any other findings in the Gemfile
	r.Packages = common.NewLister()
	if hasRmagick, _ := common.GetGemVersion(r.Gemfile, "rmagick"); hasRmagick {
		fmt.Println(common.MsgL2, "----> Found Image Magick", common.MsgReset)
		r.Packages.Add("imagemagick", "libmagickwand-dev")
	}

	if hasSqlite, _ := common.GetGemVersion(r.Gemfile, "sqlite"); hasSqlite {
		fmt.Println(common.MsgL2, "----> Found sqlite", common.MsgReset)
		r.Packages.Add("libsqlite3-dev")
	}

	// look for DB
	dbs := common.NewLister()
	if hasMysql, _ := common.GetGemVersion(r.Gemfile, "mysql2"); hasMysql {
		fmt.Println(common.MsgL2, "----> Found Mysql", common.MsgReset)
		dbs.Add("mysql")
	}

	if hasPg, _ := common.GetGemVersion(r.Gemfile, "pg"); hasPg {
		fmt.Println(common.MsgL2, "----> Found PostgreSQL", common.MsgReset)
		dbs.Add("postgresql")
	}

	if hasRedis, _ := common.GetGemVersion(r.Gemfile, "redis"); hasRedis {
		fmt.Println(common.MsgL2, "----> Found Redis", common.MsgReset)
		dbs.Add("redis")
	}

	parseContext := &common.ParseContext{Services: []*common.Service{service}, Dbs: dbs.Items}

	return parseContext, nil
}
