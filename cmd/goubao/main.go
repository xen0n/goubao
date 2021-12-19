// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"path"
	"sort"

	"golang.org/x/tools/go/packages"
)

func readAPISpecs(filename string) ([]routeDescOutput, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	istream := bufio.NewReader(f)

	decoder := json.NewDecoder(istream)

	var result []routeDescOutput
	err = decoder.Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func gatherPackagePatterns(specs []routeDescOutput) []string {
	var result []string
	seenPkgs := make(map[string]struct{})

	add := func(s string) {
		if len(s) == 0 {
			// skip empty types
			return
		}
		if _, seen := seenPkgs[s]; seen {
			return
		}
		result = append(result, s)
		seenPkgs[s] = struct{}{}
	}

	for _, spec := range specs {
		add(spec.Func.Ident.Pkg)
		add(spec.ReqType.Ident.Pkg)
		add(spec.RespType.Ident.Pkg)
	}

	sort.Strings(result)

	return result
}

func main() {

	apiSpecFilename := os.Args[1]

	apiSpecs, err := readAPISpecs(apiSpecFilename)
	if err != nil {
		panic(err)
	}
	fmt.Printf("read %d API specs from %s\n", len(apiSpecs), apiSpecFilename)

	pkgPatterns := gatherPackagePatterns(apiSpecs)
	fmt.Printf("will parse these packages:\n\n")
	for _, pkgPath := range pkgPatterns {
		fmt.Printf("  - %s\n", pkgPath)
	}
	fmt.Printf("\n")

	conf := packages.Config{
		Mode: (packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo),
	}
	pkgs, err := packages.Load(&conf, pkgPatterns...)
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgs {
		fmt.Printf("Package %s:\n\n", pkg)
		tcx := tyContext{
			tyInfo: pkg.TypesInfo,
		}
		_ = tcx

		for i, fileNode := range pkg.Syntax {
			filePath := pkg.CompiledGoFiles[i]
			basename := path.Base(filePath)
			_ = fileNode
			fmt.Printf(">>> %s\t%s\n", basename, filePath)
		}

		fmt.Printf("\n")
	}
}

type tyContext struct {
	tyInfo *types.Info
}

//nolint:unused // WIP
func (tcx *tyContext) typeOf(e ast.Expr) types.Type {
	return tcx.tyInfo.TypeOf(e)
}
