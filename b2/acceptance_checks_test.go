//####################################################################
//
// File: b2/acceptance_checks_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// collectCollectionNames walks the schema and records the attribute names that are lists, sets or
// maps, and the ones that are plain values, wherever they are nested.
func collectCollectionNames(sm map[string]*schema.Schema, collections, plain map[string]bool) {
	for name, s := range sm {
		switch s.Type {
		case schema.TypeList, schema.TypeSet, schema.TypeMap:
			collections[name] = true
		default:
			plain[name] = true
		}
		if elem, ok := s.Elem.(*schema.Resource); ok {
			collectCollectionNames(elem.Schema, collections, plain)
		}
	}
}

// Terraform stores a list as "name.#" and a map as "name.%" in the flat state, so a key that names
// the list or map itself does not exist there. TestCheckResourceAttrPair passes when both keys are
// missing, so such a pair compares nothing and can never fail.
func TestAttrPairChecksDoNotUseCollectionRootKeys(t *testing.T) {
	// The helper behaves as described, which is why the keys must be checked
	state := &terraform.State{Modules: []*terraform.ModuleState{{
		Path: []string{"root"},
		Resources: map[string]*terraform.ResourceState{
			"a.x": {Primary: &terraform.InstanceState{ID: "1", Attributes: map[string]string{"list.#": "1", "list.0": "a"}}},
			"b.y": {Primary: &terraform.InstanceState{ID: "2", Attributes: map[string]string{"list.#": "1", "list.0": "b"}}},
		},
	}}}
	if err := resource.TestCheckResourceAttrPair("a.x", "list", "b.y", "list")(state); err != nil {
		t.Skipf("the SDK now rejects absent keys, so this guard is no longer needed: %v", err)
	}

	p := New("test", "unused")()
	collections, plain := map[string]bool{}, map[string]bool{}
	for _, res := range p.DataSourcesMap {
		collectCollectionNames(res.Schema, collections, plain)
	}
	for _, res := range p.ResourcesMap {
		collectCollectionNames(res.Schema, collections, plain)
	}

	files, err := filepath.Glob("*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, file := range files {
		parsed, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(parsed, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) != 4 {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "TestCheckResourceAttrPair" {
				return true
			}
			for _, arg := range []ast.Expr{call.Args[1], call.Args[3]} {
				lit, ok := arg.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				key, _ := strconv.Unquote(lit.Value)
				last := key[strings.LastIndex(key, ".")+1:]
				if collections[last] && !plain[last] {
					t.Errorf("%s: key %q names a list or map, which is not in the flat state; use %q or an element key",
						fset.Position(lit.Pos()), key, key+".#")
				}
			}
			return true
		})
	}
}

// exprText returns the source text of a simple expression: the unquoted value of a string literal
// or the name of an identifier.
func exprText(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.BasicLit:
		s, _ := strconv.Unquote(v.Value)
		return s
	case *ast.Ident:
		return v.Name
	}
	return ""
}

// Within one acceptance step, the elements of a data source list (file_versions.N.attr) are compared
// with one resource each. A pair that compares file_versions.N.attr with a different attribute, or
// whose index N is paired with two different resources, compares the wrong objects and passes or
// fails for the wrong reason. This is checked offline because the acceptance tests need B2.
func TestFileVersionsPairsCompareMatchingIndexAndResource(t *testing.T) {
	files, err := filepath.Glob("data_source_b2_bucket_file*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no file data source tests found")
	}
	indexKey := regexp.MustCompile(`^file_versions\.(\d+)\.(.+)$`)
	pairs := 0
	fset := token.NewFileSet()
	for _, file := range files {
		parsed, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(parsed, func(n ast.Node) bool {
			step, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := step.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "ComposeTestCheckFunc" {
				return true
			}
			resourceByIndex := map[string]string{}
			for _, arg := range step.Args {
				call, ok := arg.(*ast.CallExpr)
				if !ok || len(call.Args) != 4 {
					continue
				}
				fn, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || fn.Sel.Name != "TestCheckResourceAttrPair" {
					continue
				}
				m := indexKey.FindStringSubmatch(exprText(call.Args[1]))
				if m == nil {
					continue
				}
				pairs++
				index, attr := m[1], m[2]
				if other := exprText(call.Args[3]); other != attr {
					t.Errorf("%s: file_versions.%s.%s is compared with %q, expected %q",
						fset.Position(call.Pos()), index, attr, other, attr)
				}
				res := exprText(call.Args[2])
				if prev, seen := resourceByIndex[index]; seen && prev != res {
					t.Errorf("%s: file_versions.%s is compared with both %s and %s in one step",
						fset.Position(call.Pos()), index, prev, res)
				}
				resourceByIndex[index] = res
			}
			return true
		})
	}
	if pairs == 0 {
		t.Fatal("found no file_versions pairs, the guard no longer looks at anything")
	}
}
