// SPDX-License-Identifier: MIT

package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type tyContext struct {
	pkgsByImportPath map[string]*packages.Package

	// map[pkgImportPath]map[tyName]decl
	types map[string]map[string]iTySpec

	// map[pkgImportPath]map[tyName]map[methodName]method
	methods map[string]map[string]map[string]*ast.FuncDecl
}

func (tcx *tyContext) lookupTy(
	apiTy *goType,
) (spec iTySpec, ok bool) {
	if len(apiTy.Ident.Pkg) == 0 {
		// assume basic type
		obj := types.Universe.Lookup(apiTy.Ident.Name)
		if obj == nil {
			return nil, false
		}

		return &builtinTySpec{
			inner: obj.Type(),
		}, true
	}

	pkg, ok := tcx.types[apiTy.Ident.Pkg]
	if !ok {
		return nil, false
	}

	spec, ok = pkg[apiTy.Ident.Name]
	return spec, ok
}

func (tcx *tyContext) lookupMethod(
	pkgImportPath string,
	receiverTyName string,
	methodName string,
) (decl *ast.FuncDecl, ok bool) {
	pkg, ok := tcx.methods[pkgImportPath]
	if !ok {
		return nil, false
	}

	recvTy, ok := pkg[receiverTyName]
	if !ok {
		return nil, false
	}

	decl, ok = recvTy[methodName]
	return decl, ok
}

func (tcx *tyContext) populateTyIndex() {
	tcx.types = make(map[string]map[string]iTySpec)
	tcx.methods = make(map[string]map[string]map[string]*ast.FuncDecl)

	for _, pkg := range tcx.pkgsByImportPath {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				switch decl := decl.(type) {
				case *ast.GenDecl:
					for _, spec := range decl.Specs {
						// we're only interested in type declarations
						spec, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}

						tcx.recordTy(pkg, spec)
					}

				case *ast.FuncDecl:
					if decl.Recv == nil {
						// this is a function, yet we only care about methods
						// at this point
						continue
					}
					tcx.recordMethod(pkg, decl)

				default:
					// do nothing
				}
			}
		}
	}
}

func (tcx *tyContext) recordTy(pkg *packages.Package, ty *ast.TypeSpec) {
	tyPkg := pkg.PkgPath
	tyName := ty.Name.Name

	tcx.addTyToIndex(tyPkg, tyName, ty)
}

func (tcx *tyContext) recordMethod(pkg *packages.Package, decl *ast.FuncDecl) {
	if decl.Recv == nil {
		panic("function passed to recordMethod")
	}
	if len(decl.Recv.List) != 1 {
		panic("len(receiver list) not 1")
	}

	receiverTy := pkg.TypesInfo.TypeOf(decl.Recv.List[0].Type)
	receiverTyPkg, receiverTyName := getPurifiedTypeForReceiver(receiverTy)

	tcx.addMethodToIndex(receiverTyPkg, receiverTyName, decl.Name.Name, decl)
}

func (tcx *tyContext) addTyToIndex(
	pkgImportPath string,
	tyName string,
	spec *ast.TypeSpec,
) {
	pkg, ok := tcx.types[pkgImportPath]
	if !ok {
		pkg = make(map[string]iTySpec)
		tcx.types[pkgImportPath] = pkg
	}

	if _, ok := pkg[tyName]; ok {
		panic("attempted to add duplicate type")
	}

	pkg[tyName] = &namedTySpec{inner: spec}
}

func (tcx *tyContext) addMethodToIndex(
	pkgImportPath string,
	tyName string,
	methodName string,
	decl *ast.FuncDecl,
) {
	pkg, ok := tcx.methods[pkgImportPath]
	if !ok {
		pkg = make(map[string]map[string]*ast.FuncDecl)
		tcx.methods[pkgImportPath] = pkg
	}

	ty, ok := pkg[tyName]
	if !ok {
		ty = make(map[string]*ast.FuncDecl)
		pkg[tyName] = ty
	}

	if _, ok := ty[methodName]; ok {
		panic("attempted to add duplicate method")
	}
	ty[methodName] = decl
}

// get (package import path, typename identifier) of receiver type, without
// caring if it is pointer receiver or not
func getPurifiedTypeForReceiver(ty types.Type) (pkgImportPath string, name string) {
	switch ty := ty.(type) {
	case *types.Pointer:
		return getPurifiedTypeForReceiver(ty.Elem())

	case *types.Named:
		name := ty.Obj()
		return name.Pkg().Path(), name.Name()
	}

	panic("not implemented")
}
