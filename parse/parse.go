package parse

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/dropbox/godropbox/errors"
	"github.com/pacur/pacur/pack"
	"github.com/pacur/pacur/utils"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/shell"
	"mvdan.cc/sh/v3/syntax"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
		Distro:     distro,
		Release:    release,
		Root:       root,
		Home:       home,
		SourceDir:  filepath.Join(root, "src"),
		PackageDir: filepath.Join(root, "pkg"),
	}

	pac.Init()

	if err = PkgBuild(pac, path); err != nil {
		return
	}

	return
}

func bashit(path string) error {
	ctx := context.Background()

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open: %v", err)
	}
	defer f.Close()
	file, err := syntax.NewParser().Parse(f, path)
	if err != nil {
		return fmt.Errorf("could not parse: %v", err)
	}

	vars, err := shell.SourceNode(ctx, file)
	if err != nil {
		return err
	}

	/*	for key, val := range vars {
		fmt.Printf("%#v\t", key)
		switch val.Kind {
		case expand.String:
			fallthrough
		case expand.NameRef:
			fmt.Printf("%#v\n", val.Str)
		case expand.Indexed:
			fmt.Println()
			for i, v := range val.List {
				fmt.Printf("\t\t%d: %#v\n", i, v)
			}
		case expand.Associative:
			fmt.Println()
			for k, v := range val.Map {
				fmt.Printf("\t\t%#v: %#v\n", k, v)
			}
		}
	}*/
	_ = expand.String
	_ = vars

	syntax.Walk(file, func(node syntax.Node) bool {
		switch x := node.(type) {
		case *syntax.FuncDecl:
			if x.Name != nil {
				fmt.Println("---")
				fmt.Println(x.Name.Value)
				fmt.Println("===")
				cmd := x.Body.Cmd

				block, ok := cmd.(*syntax.Block)
				if !ok {
					return false
				}

				buf := &bytes.Buffer{}
				printer := syntax.NewPrinter()
				for _, stmt := range block.Stmts {
					printer.Print(buf, stmt)
					fmt.Fprintln(buf)
				}
				fmt.Println(buf.String())
			}
		}
		return true
	})

	//syntax.NewPrinter().Print(os.Stdout, f)

	return nil
}

func PkgBuild(pac *pack.Pack, path string) (err error) {
	if err := bashit(path); err != nil {
		return err
	}

	file, err := utils.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	return nil

	n := 0
	blockType := 0
	blockKey := ""
	blockData := ""
	blockItems := []string{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		n += 1

		if line == "" || line[:1] == "#" {
			continue
		}

		if blockType == blockList {
			if line == ")" {
				for _, item := range itemReg.FindAllString(blockData, -1) {
					blockItems = append(blockItems, item[1:len(item)-1])
				}
				err = pac.AddItem(blockKey, blockItems, n, line)
				if err != nil {
					return
				}
				blockType = 0
				blockKey = ""
				blockData = ""
				blockItems = []string{}
				continue
			}

			blockData += strings.TrimSpace(line)
		} else if blockType == blockFunc {
			if line == "}" {
				err = pac.AddItem(blockKey, blockItems, n, line)
				if err != nil {
					return
				}
				blockType = 0
				blockKey = ""
				blockItems = []string{}
				continue
			}

			blockItems = append(blockItems, strings.TrimSpace(line))
		} else {
			if strings.Contains(line, "() {") {
				blockType = blockFunc
				blockKey = strings.Split(line, "() {")[0]
			} else {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					err = &SyntaxError{
						errors.Newf("parse: Line missing '=' (%d: %s)",
							n, line),
					}
					return
				}

				key := parts[0]
				val := parts[1]

				if key[:1] == " " {
					err = &SyntaxError{
						errors.Newf("parse: Extra space padding (%d: %s)",
							n, line),
					}
					return
				} else if key[len(key)-1:] == " " {
					err = &SyntaxError{
						errors.Newf(
							"parse: Extra space before '=' (%d: %s)",
							n, line),
					}
					return
				}

				valLen := len(val)
				switch val[:1] {
				case `"`, "`":
					if val[valLen-1:] != val[:1] {
						err = &SyntaxError{
							errors.Newf("parse: Unexpected char '%s' "+
								"expected '%s' (%d: %s)",
								val[valLen-1:], val[:1], n, line),
						}
						return
					}

					err = pac.AddItem(key, val[1:valLen-1], n, line)
					if err != nil {
						return
					}
				case "(":
					if val[valLen-1:] == ")" {
						if val[1:2] != `"` && val[1:2] != "`" {
							err = &SyntaxError{
								errors.Newf("parse: Unexpected char '%s' "+
									"expected '\"' or '`' (%d: %s)",
									val[1:2], n, line),
							}
							return
						}

						if val[valLen-2:valLen-1] != val[1:2] {
							err = &SyntaxError{
								errors.Newf("parse: Unexpected char '%s' "+
									"expected '%s' (%d: %s)",
									val[valLen-2:valLen-1], val[1:2],
									n, line),
							}
							return
						}

						val = val[2 : len(val)-2]
						err = pac.AddItem(key, []string{val}, n, line)
						if err != nil {
							return
						}
					} else {
						blockType = blockList
						blockKey = key
					}
				case " ":
					err = &SyntaxError{
						errors.Newf(
							"parse: Extra space after '=' (%d: %s)",
							n, line),
					}
					return
				default:
					err = &SyntaxError{
						errors.Newf(
							"parse: Unexpected char '%s' expected "+
								"'\"' or '`' (%d: %s)", val[:1], n, line),
					}
					return
				}
			}
		}
	}

	return
}
