package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pacur/pacur/pack"
	"github.com/pacur/pacur/parse"
)

func ParsePkgBuild() (err error) {
	var (
		distro  string
		release string
	)

	for i, path := range flag.Args() {
		if i == 0 {
			continue // skip "parse"
		}
		if i == 1 {
			split := strings.Split(flag.Arg(1), "-")
			distro = split[0]
			release = ""
			if len(split) > 1 {
				release = split[1]
			}
			continue
		}

		fmt.Println(path)
		pac := &pack.Pack{
			Distro:  distro,
			Release: release,
		}
		pac.Init()
		if err = parse.PkgBuild(pac, path); err != nil {
			return
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "   ")
		enc.Encode(pac)
	}
	return
}
