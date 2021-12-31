package sframe

import (
	"testing"

	"github.com/npillmayer/schuko/tracing/gotestingadapter"
)

func TestDeclareNumericVariable(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.runtime")
	defer teardown()
	//
	decl := MakeTagDecl(TagNumeric, "a")
	if decl.Name() != "a" {
		t.Errorf("expected declared symbol to be named 'a', is %q", decl.Name())
	}
	if decl.TagType() != TagNumeric {
		t.Errorf("expected declared symbol to be of type numeric, is %s", decl.TagType().String())
	}
	if !decl.IsTag() {
		t.Error("expected declared symbol to be a tag, isn't")
	}
	if decl.IsArray() {
		t.Error("expected declared symbol to be scalar, is array")
	}
}

func TestDeclareArrayVariable(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.runtime")
	defer teardown()
	//
	decl := MakeTagDecl(TagNumeric, "a", "r", "[]")
	if decl.Name() != "a.r[]" {
		t.Errorf("expected declared symbol to be named 'a.r[]', is %q", decl.Name())
	}
	if decl.TagType() != TagNumeric {
		t.Errorf("expected declared symbol to be of type numeric, is %s", decl.TagType().String())
	}
	if !decl.IsTag() {
		t.Error("expected declared symbol to be a tag, isn't")
	}
	if !decl.IsArray() {
		t.Error("expected declared symbol to be scalar, is array")
	}
	if decl.arraycnt != 1 {
		t.Errorf("expected declared symbol to have 1 array, has %d", decl.arraycnt)
	}
}
