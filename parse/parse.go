package parse

import (
	//	"bufio"
	"bytes"
	"context"
	//	"fmt"
	"github.com/dropbox/godropbox/errors"
	"github.com/mitchellh/mapstructure"
	"github.com/pacur/pacur/pack"
	"github.com/pacur/pacur/utils"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/shell"
	"mvdan.cc/sh/v3/syntax"
	//	"os"
	"path/filepath"
	"regexp"
	//	"strings"
)

const (
	root      = "/pacur_build"
	blockList = 1
	blockFunc = 2
)

var (
	itemReg = regexp.MustCompile("(\"[^\"]+\")|(`[^`]+`)")
)

func File(distro, release, home string) (pac *pack.Pack, err error) {
	home, err = filepath.Abs(home)
	if err != nil {
		err = &FileError{
			errors.Wrapf(err, "parse: Failed to get root directory from '%s'",
				home),
		}
	}

	err = utils.ExistsMakeDir(root)
	if err != nil {
		return
	}

	err = utils.CopyFiles(home, root, false)
	if err != nil {
		return
	}
	path := filepath.Join(root, "PKGBUILD")

	pac = &pack.Pack{
		Distro:  distro,
		Release: release,
		Root:    root,
		Home:    home,
		SrcDir:  filepath.Join(root, "src"),
		PkgDir:  filepath.Join(root, "pkg"),
	}

	pac.Init()

	if err = PkgBuild(pac, path); err != nil {
		return
	}

	return
}

func PkgBuild(pac *pack.Pack, path string) (err error) {
	file, err := utils.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	rootNode, err := syntax.NewParser().Parse(file, path)
	if err != nil {
		return errors.Newf("parse: File %#v: %v", path, err)
	}

	vars, err := shell.SourceNode(context.Background(), rootNode)
	if err != nil {
		return errors.Newf("parse: SourceNode on file %#v: %v", path, err)
	}

	m := make(map[string]interface{})

	for key, val := range vars {
		switch val.Kind {
		case expand.String:
			fallthrough
		case expand.NameRef:
			m[key] = val.Str
		case expand.Indexed:
			m[key] = val.List
		case expand.Associative:
			m[key] = val.Map
		default:
			return errors.Newf("parse: Unhandled shell variable type %#v", val.Kind)
		}
	}

	syntax.Walk(rootNode, func(node syntax.Node) bool {
		switch x := node.(type) {
		case *syntax.FuncDecl:
			if x.Name == nil {
				return false
			}
			cmd := x.Body.Cmd
			if cmd == nil {
				return false
			}
			block, ok := cmd.(*syntax.Block)
			if !ok {
				return false
			}
			lines := []string{}
			printer := syntax.NewPrinter()
			for _, stmt := range block.Stmts {
				var buf bytes.Buffer
				printer.Print(&buf, stmt)
				lines = append(lines, buf.String())
			}
			m[x.Name.Value] = lines
		}
		return true
	})

	// TODO Handle distro/release suffixed fields somehow.
	// Not sure the best way to do this. I'd just run through and
	// copy them here, but then shell variable substitution wouldn't
	// work right. I'm going to think about a bash-ey way to do it.

	if err := mapstructure.Decode(m, &pac); err != nil {
		return errors.Wrapf(err, "parse: Failed to map input structure to pack structure")
	}

	return
}
