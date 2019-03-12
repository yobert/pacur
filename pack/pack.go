package pack

import (
	"github.com/dropbox/godropbox/errors"
	//"github.com/pacur/pacur/constants"
	//"github.com/pacur/pacur/resolver"
	//"strings"
)

type Pack struct {
	// Metadata
	Targets     []string
	Distro      string
	Release     string
	FullRelease string
	Root        string
	Home        string
	SrcDir      string
	PkgDir      string
	PkgName     string
	PkgVer      string
	PkgRel      string
	PkgDesc     string
	PkgDescLong []string
	Maintainer  string
	Arch        string
	License     []string
	Section     string
	Priority    string
	Url         string
	Depends     []string
	OptDepends  []string
	MakeDepends []string
	Provides    []string
	Conflicts   []string
	Sources     []string
	HashSums    []string
	Backup      []string

	// Function bodies (lists of command lines)
	Build    []string
	Package  []string
	PreInst  []string
	PostInst []string
	PreRm    []string
	PostRm   []string
}

func (p *Pack) Init() {
	p.FullRelease = p.Distro
	if p.Release != "" {
		p.FullRelease += "-" + p.Release
	}
}

func (p *Pack) Validate() (err error) {
	if len(p.Sources) == len(p.HashSums) {
	} else if len(p.Sources) > len(p.HashSums) {
		err = &ValidationError{
			errors.New("pack: Missing hash sum for source"),
		}
		return
	} else {
		err = &ValidationError{
			errors.New("pack: Too many hash sums for sources"),
		}
		return
	}

	return
}

func (p *Pack) Compile() error {
	return p.Validate()
}
