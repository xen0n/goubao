// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"path"

	"golang.org/x/tools/go/packages"
)

func main() {
	conf := packages.Config{
		Mode: (packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo),
	}

	srcPackageName := os.Args[1]
	fmt.Printf("reading %s\n", srcPackageName)

	pkgs, err := packages.Load(&conf, srcPackageName)
	if err != nil {
		panic(err)
	}

	pkg := pkgs[0]
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
}

type tyContext struct {
	tyInfo *types.Info
}

//nolint:unused // WIP
func (tcx *tyContext) typeOf(e ast.Expr) types.Type {
	return tcx.tyInfo.TypeOf(e)
}
