// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"go/ast"
	"go/types"
)

func isTypeEmpty(ty goType) bool {
	return ty.Ident == goIdent{}
}

type iTySpec interface {
	tySpecMarker()

	String() string
}

type emptyTySpec struct{}

var _ iTySpec = (*emptyTySpec)(nil)

func (*emptyTySpec) tySpecMarker() {}

func (*emptyTySpec) String() string { return "<empty type>" }

type builtinTySpec struct {
	inner types.Type
}

var _ iTySpec = (*builtinTySpec)(nil)

func (*builtinTySpec) tySpecMarker() {}

func (s *builtinTySpec) String() string {
	return fmt.Sprintf("<built-in type %s>", s.inner.String())
}

type namedTySpec struct {
	inner *ast.TypeSpec
}

var _ iTySpec = (*namedTySpec)(nil)

func (*namedTySpec) tySpecMarker() {}

func (s *namedTySpec) String() string {
	return fmt.Sprintf("<named type %s>", s.inner.Name.String())
}
