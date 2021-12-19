// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
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

	pkgsByImportPath := make(map[string]*packages.Package, len(pkgs))
	for _, pkg := range pkgs {
		pkgsByImportPath[pkg.PkgPath] = pkg
	}

	tcx := tyContext{
		pkgsByImportPath: pkgsByImportPath,
	}
	tcx.populateTyIndex()

	for _, spec := range apiSpecs {
		handlerTyPkgPath := spec.Func.Receiver.Ident.Pkg
		handlerTyName := spec.Func.Receiver.Ident.Name
		handlerMethName := spec.Func.Ident.Name

		fmt.Printf("- %s %s:\n", spec.HTTPMethod, spec.Path)

		meth, ok := tcx.lookupMethod(handlerTyPkgPath, handlerTyName, handlerMethName)
		if !ok {
			fmt.Printf(
				"!!! unresolved handler method (%s.%s).%s\n",
				handlerTyPkgPath,
				handlerTyName,
				handlerMethName,
			)
			continue
		}

		reqTy, ok := tcx.lookupTy(&spec.ReqType)
		if !ok {
			if isTypeEmpty(spec.ReqType) {
				reqTy = &emptyTySpec{}
			} else {
				fmt.Printf(
					"!!! unresolved request type %s.%s\n",
					spec.ReqType.Ident.Pkg,
					spec.ReqType.Ident.Name,
				)
			}
		}

		respTy, ok := tcx.lookupTy(&spec.RespType)
		if !ok {
			if isTypeEmpty(spec.RespType) {
				respTy = &emptyTySpec{}
			} else {
				fmt.Printf(
					"!!! unresolved response type %s.%s\n",
					spec.RespType.Ident.Pkg,
					spec.RespType.Ident.Name,
				)
			}
		}

		fmt.Printf("  - meth: %v\n", meth)
		fmt.Printf("  - reqTy: %s\n", reqTy.String())
		fmt.Printf("  - respTy: %s\n", respTy.String())
	}
}
