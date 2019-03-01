package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/pacur/pacur/pack"
	"github.com/pacur/pacur/parse"
)

func ParsePkgBuild() (err error) {
	for i, path := range flag.Args() {
		if i == 0 {
			continue // skip "parse"
		}
		fmt.Println(path)
		pac := &pack.Pack{}
		pac.Init()
		if err = parse.PkgBuild(pac, path); err != nil {
			return
		}
		//fmt.Printf("%#v\n", *pac)
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "   ")
		enc.Encode(pac)
	}
	return
}
