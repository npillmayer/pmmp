package variables_test

import (
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/npillmayer/gorgo/runtime"
	"github.com/npillmayer/pmmp/variables"
	"github.com/npillmayer/schuko/gtrace"
	"github.com/npillmayer/schuko/tracing/gologadapter"
)

type TestingErrorListener struct {
	*antlr.DefaultErrorListener // use default as base class
}

var parserr string

/* Our error listener prints an error and set a global error flag.
 */
func (c *TestingErrorListener) SyntaxError(r antlr.Recognizer, sym interface{},
	line, column int, msg string, e antlr.RecognitionException) {
	//
	at := fmt.Sprintf("%s:%s", strconv.Itoa(line), strconv.Itoa(column))
	log.Printf("line %s: %.44s", at, msg)
	parserr = msg
}

func newErrL() antlr.ErrorListener {
	return &TestingErrorListener{}
}

func checkErr(t *testing.T) {
	if parserr != "" {
		t.Errorf(parserr)
	}
}

// Init the global tracers.
func TestInit0(t *testing.T) {
	gtrace.InterpreterTracer = gologadapter.New()
	gtrace.SyntaxTracer = gologadapter.New()
}

func TestVarDecl1(t *testing.T) {
	symtab := runtime.NewSymbolTable()
	symtab.DefineTag("x")
}

func TestVarDecl2(t *testing.T) {
	symtab := runtime.NewSymbolTable()
	x := variables.NewVarDecl("x")
	symtab.InsertTag(x.Tag())
	variables.CreateSuffix("r", variables.SuffixType, x.AsSuffix())
}

/*
func TestVarDecl3(t *testing.T) {
	symtab := runtime.NewSymbolTable()
	x, _ := symtab.DefineTag("x")
	var v *variables.VarDecl = x.(*variables.VarDecl)
	variables.CreateVarDecl("r", variables.ComplexSuffix, v)
	arr := variables.CreateVarDecl("<array>", variables.ComplexArray, x.(*variables.VarDecl))
	variables.CreateVarDecl("a", variables.ComplexSuffix, arr)
	//var b *bytes.Buffer
	//b = v.ShowVariable(b)
	//fmt.Printf("## showvariable %s;\n%s\n", v.BaseTag.GetName(), b.String())
}

func TestVarRef1(t *testing.T) {
	x := variables.CreateVarDecl("x", variables.NumericType, nil)
	var v *variables.VarRef = variables.CreateVarRef(x, 1, nil)
	t.Logf("var ref: %v\n", v)
}

func TestVarRef2(t *testing.T) {
	x := variables.CreateVarDecl("x", variables.NumericType, nil)
	r := variables.CreateVarDecl("r", variables.ComplexSuffix, x)
	var v *variables.VarRef = variables.CreateVarRef(r, 1, nil)
	t.Logf("var ref: %v\n", v)
}

func TestVarRef3(t *testing.T) {
	x := variables.CreateVarDecl("x", variables.NumericType, nil)
	arr := variables.CreateVarDecl("<[]>", variables.ComplexArray, x)
	subs := []decimal.Decimal{decimal.New(7, 0)}
	var v *variables.VarRef = variables.CreateVarRef(arr, 7, subs)
	t.Logf("var ref: %v\n", v)
}
*/

/*
func TestVarRefParser1(t *testing.T) {
	listener.ParseVariableFromString("x@", newErrL())
	checkErr(t)
}

func TestVarRefParser2(t *testing.T) {
	listener.ParseVariableFromString("x1@", newErrL())
	checkErr(t)
}

func TestVarRefParser3(t *testing.T) {
	listener.ParseVariableFromString("x.a@", newErrL())
	checkErr(t)
}

func TestVarRefParser4(t *testing.T) {
	listener.ParseVariableFromString("xyz18abc@", newErrL())
	checkErr(t)
}

func TestVarRefParser5(t *testing.T) {
	listener.ParseVariableFromString("xyz18.abc@", newErrL())
	checkErr(t)
}

func TestVarRefParser6(t *testing.T) {
	listener.ParseVariableFromString("x1a[2]b@", newErrL())
	checkErr(t)
}
*/
