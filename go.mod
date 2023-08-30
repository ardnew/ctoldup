module github.com/ardnew/ctoldup

go 1.15

replace (
	github.com/Masterminds/vcs => /Users/andrew/Development/go/src/github.com/Masterminds/vcs
	github.com/ardnew/roster => /Users/andrew/Development/go/src/github.com/ardnew/roster
)

require (
	github.com/Masterminds/vcs v1.13.1
	github.com/ardnew/roster v0.0.0-20200915183949-0a672b963792
	github.com/ardnew/version v0.2.0
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/mholt/archiver/v3 v3.3.0
	github.com/otiai10/copy v1.2.0
	gopkg.in/yaml.v3 v3.0.0
)
