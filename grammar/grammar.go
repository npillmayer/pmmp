package grammar

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/npillmayer/gorgo/lr"
	"github.com/npillmayer/gorgo/lr/earley"
	"github.com/npillmayer/gorgo/lr/scanner"
	"github.com/npillmayer/gorgo/lr/sppf"
	"github.com/npillmayer/gorgo/terex"
	"github.com/npillmayer/gorgo/terex/termr"
	"github.com/npillmayer/schuko/gtrace"
	"github.com/npillmayer/schuko/tracing"
	"github.com/timtadh/lexmachine"
)

// --- Initialization --------------------------------------------------------

// T traces to the global syntax tracer.
func T() tracing.Trace {
	return gtrace.SyntaxTracer
}

var startOnce sync.Once // monitors one-time creation of grammar and lexer

var mpGrammar *lr.LRAnalysis
var mpLexer *scanner.LMAdapter

func initGrammar() {
	startOnce.Do(func() {
		var err error
		T().Infof("Creating lexer")
		if mpLexer, err = Lexer(); err != nil { // MUST be called before grammar builing !
			panic("Cannot create lexer")
		}
		T().Infof("Creating grammar")
		if mpGrammar, err = MakeMetaPostGrammar(); err != nil {
			panic("Cannot create global grammar")
		}
		initRewriters()
	})
}

func createParser() *earley.Parser {
	initGrammar()
	return earley.NewParser(mpGrammar, earley.GenerateTree(true))
}

// ---------------------------------------------------------------------------

// MakeMetaPostGrammar generates a grammar for the MetaPost language,
// constructs all the parsing table for it and returns
// them as a package (or an error).
//
// This is usually not called directly be clients, but rather implicitely
// envoked through Parse().
//
func MakeMetaPostGrammar() (*lr.LRAnalysis, error) {
	b := lr.NewGrammarBuilder("MetaPost")

	b.LHS("primary").N("atom").End()
	b.LHS("atom").N("variable").End()
	b.LHS("atom").T(S("NUMBER")).End()
	b.LHS("variable").T(S("TAG")).N("suffix").End()
	b.LHS("suffix").Epsilon()
	b.LHS("suffix").N("suffix").N("subscript").End()
	b.LHS("suffix").N("suffix").T(S("TAG")).End()
	b.LHS("subscript").T(S("NUMBER")).End()

	g, err := b.Grammar()
	if err != nil {
		T().Errorf("Error creating MetaPost grammar")
		return nil, err
	}
	g.Dump()
	ga := lr.Analysis(g)
	return ga, nil
}

// Parse parses an input string, given in TeREx language format. It returns the
// parse forest and a TokenReceiver, or an error in case of failure.
//
// Clients may use a terex.ASTBuilder to create an abstract syntax tree
// from the parse forest.
//
func Parse(input string) (*sppf.Forest, termr.TokenRetriever, error) {
	parser := createParser()
	scan, err := mpLexer.Scanner(input)
	if err != nil {
		return nil, nil, err
	}
	accept, err := parser.Parse(scan, nil)
	if err != nil {
		return nil, nil, err
	} else if !accept {
		return nil, nil, fmt.Errorf("Not a valid MetaPost statement")
	}
	return parser.ParseForest(), earleyTokenReceiver(parser), nil
}

func earleyTokenReceiver(parser *earley.Parser) termr.TokenRetriever {
	return func(pos uint64) interface{} {
		return parser.TokenAt(pos)
	}
}

// --- AST -------------------------------------------------------------------

// AST creates an abstract syntax tree from a parse tree/forest.
//
// Returns a homogenous AST, a TeREx-environment and an error status.
func AST(parsetree *sppf.Forest, tokRetr termr.TokenRetriever) (*terex.GCons,
	*terex.Environment, error) {
	initGrammar() // should have been done already by Parse()
	ab := newASTBuilder(mpGrammar.Grammar())
	env := ab.AST(parsetree, tokRetr)
	if env == nil {
		T().Errorf("Cannot create AST from parsetree")
		return nil, nil, fmt.Errorf("Error while creating AST")
	}
	ast := env.AST
	T().Infof("AST: %s", env.AST.ListString())
	return ast, env, nil
}

// QuoteAST returns an AST, which should be the result of parsing an s-expr, as
// pure data.
//
// If the environment contains any symbol's value, quoting will replace the symbol
// by its value. For example, if the s-expr contains a symbol 'str' with a value
// of "this is a string", the resulting data structure will contain the string,
// not the name of the symbol. If you do not have use for this kind of substitution,
// simply call Quote(…) for the global environment.
//
func QuoteAST(ast terex.Element, env *terex.Environment) (terex.Element, error) {
	// ast *terex.GCons
	if env == nil {
		env = terex.GlobalEnvironment
	}
	quEnv := terex.NewEnvironment("quoting", env)
	quEnv.Defn("list", varOp.call)
	//quEnv.Defn("quote", quoteOp.call)
	quEnv.Resolver = symbolPreservingResolver{}
	q := terex.Eval(ast, quEnv)
	T().Debugf("QuotAST returns Q = %v", q)
	q.Dump(tracing.LevelDebug)
	return q, quEnv.LastError()
}

// NewASTBuilder returns a new AST builder for the MetaPost language
func newASTBuilder(grammar *lr.Grammar) *termr.ASTBuilder {
	ab := termr.NewASTBuilder(grammar)
	ab.AddTermR(varOp)
	ab.AddTermR(suffixOp)
	ab.AddTermR(subscrOp)
	return ab
}

// === AST Rewriters =========================================================

type mpTermR struct {
	name    string
	opname  string
	rewrite func(*terex.GCons, *terex.Environment) terex.Element
	call    func(terex.Element, *terex.Environment) terex.Element
}

var _ terex.Operator = &mpTermR{}
var _ termr.TermR = &mpTermR{}

func makeASTTermR(name string, opname string) *mpTermR {
	termr := &mpTermR{
		name:   name,
		opname: opname,
	}
	return termr
}

func (trew *mpTermR) String() string {
	return trew.name
}

func (trew *mpTermR) Operator() terex.Operator {
	return trew
}

func (trew *mpTermR) Rewrite(l *terex.GCons, env *terex.Environment) terex.Element {
	T().Debugf("%s:trew.Rewrite[%s] called", trew.String(), l.ListString())
	e := trew.rewrite(l, env)
	return e
}

func (trew *mpTermR) Descend(sppf.RuleCtxt) bool {
	return true
}

func (trew *mpTermR) Call(e terex.Element, env *terex.Environment) terex.Element {
	opsym := env.FindSymbol(trew.opname, true)
	if opsym == nil {
		T().Errorf("Cannot find parsing operation %s", trew.opname)
		return e
	}
	operator, ok := opsym.Value.AsAtom().Data.(terex.Operator)
	if !ok {
		T().Errorf("Cannot call parsing operation %s", trew.opname)
		return e
	}
	return operator.Call(e, env)
}

var varOp *mpTermR    // for variable -> ... productions
var suffixOp *mpTermR // for suffix -> ... productions
var subscrOp *mpTermR // for subscript -> ... productions

func initRewriters() {
	suffixOp = makeASTTermR("suffix", "suffix")
	suffixOp.rewrite = func(l *terex.GCons, env *terex.Environment) terex.Element {
		// ⟨suffix⟩ → ε | ⟨suffix⟩ ⟨subscript⟩ | ⟨suffix⟩ TAG
		T().Debugf("suffix tree = ")
		terex.Elem(l).Dump(tracing.LevelDebug)
		if withoutArgs(l) {
			return terex.Elem(nil) // ⟨suffix⟩ → ε
		}
		var suf2 terex.Element
		tee := terex.Elem(l.Cdr).Sublist()
		if !tee.IsNil() && tee.First().AsAtom().Type() == terex.OperatorType {
			suf2 = terex.Elem(l.Cdar()) // ( ⟨subscript⟩ X ) => ( ⟨subscript⟩ X )
			if singleArg(l) {
				return suf2
			}
		}
		if singleArg(l) { // ⟨suffix⟩ → ε TAG
			ll := makeTagSuffixes(terex.Elem(l.Cdar()), env)
			l = l.Append(ll)
		} else { // ⟨suffix⟩ → ⟨suffix⟩ TAG
			ll := makeTagSuffixes(terex.Elem(l.Cddar()), env)
			l = terex.Cons(suf2.AsAtom(), ll)
		}
		return terex.Elem(l)
	}
	subscrOp = makeASTTermR("subscript", "subscript")
	subscrOp.rewrite = func(l *terex.GCons, env *terex.Environment) terex.Element {
		T().Infof("subscript tree = ")
		terex.Elem(l).Dump(tracing.LevelInfo)
		if singleArg(l) { // ⟨subscript⟩ → NUMBER
			T().Errorf("⟨subscript⟩ → NUMBER ")
			e := terex.Elem(l.Cdar())
			e = convertTerminalToken(e, env)
			return terex.Elem(l) // ( ⟨subscript⟩ NUMBER )
		}
		T().Errorf("⟨subscript⟩ → [ expr ] ")
		// ⟨subscript⟩ → '[' ⟨expr⟩ ']'
		sscr := terex.Cons(l.Car, terex.Cons(l.Cddar(), nil))
		return terex.Elem(sscr) // ( ⟨subscript⟩ expr )
	}
	varOp = makeASTTermR("variable", "variable")
	varOp.rewrite = func(l *terex.GCons, env *terex.Environment) terex.Element {
		// ⟨variable⟩ → TAG ⟨suffix⟩
		T().Debugf("variable tree = ")
		terex.Elem(l).Dump(tracing.LevelDebug)
		ll := makeTagSuffixes(terex.Elem(l.Cdar()), env)
		ll = ll.Append(l.Cddr())
		l = terex.Cons(l.Car, ll) // prepend #variable operator
		return terex.Elem(l)
	}
}

func withoutArgs(l *terex.GCons) bool {
	return l == nil || l.Length() <= 1
}

func singleArg(l *terex.GCons) bool {
	return l != nil && l.Length() == 2
}

func makeTagSuffixes(arg terex.Element, env *terex.Environment) *terex.GCons {
	tok := arg.AsAtom()
	T().Infof("TAG = %v", tok)         // TAG is []string, e.g. "a.r" ⇒ "a", "r"
	if tok.Type() != terex.TokenType { // not a TAG token
		panic("cannot make suffixes from non-token tag")
	}
	T().Infof("TAG = %v", tok) // TAG is []string, e.g. "a.r" ⇒ "a", "r"
	e := convertTerminalToken(terex.Elem(tok), env)
	tag := e.AsAtom().Data.(*terex.Token).Value.([]string)
	var l *terex.GCons           // create list of suffix nodes, one for each suffix
	for _, suffix := range tag { // TAG="a.r" ⇒ "a", "r"
		node := terex.Cons(terex.Atomize(suffixOp), terex.Cons(terex.Atomize(suffix), nil))
		l = l.Branch(node) // branch = (#suffix X)
	} // now: ( (#suffix "a") (#suffix "r") )
	return l
}

// === Terminal tokens =======================================================

func convertTerminalToken(el terex.Element, env *terex.Environment) terex.Element {
	if !el.IsAtom() {
		return el
	}
	atom := el.AsAtom()
	if atom.Type() != terex.TokenType {
		return el
	}
	t := atom.Data.(*terex.Token)
	token := t.Token.(*lexmachine.Token)
	T().Infof("Convert terminal token: '%v'", string(token.Lexeme))
	switch token.Type {
	case tokenIds["NUMBER"]:
		if f, err := strconv.ParseFloat(string(token.Lexeme), 64); err == nil {
			T().Debugf("   t.Value=%g", f)
			t.Value = f
		} else {
			T().Errorf("   %s", err.Error())
			return terex.Elem(terex.Atomize(err))
		}
	case tokenIds["STRING"]:
		if (len(token.Lexeme)) <= 2 {
			t.Value = ""
		} else { // trim off "…"
			//runes := []rune(string(token.Lexeme))  // unnecessary
			t.Value = string(token.Lexeme[1 : len(token.Lexeme)-1])
		}
	case tokenIds["TAG"]: // return []string value, split at '.'
		tag := string(token.Lexeme)
		tags, err := splitTagName(tag)
		if err != nil {
			T().Errorf("illegal tag name")
			t.Value = []string{"<illegal tag>"}
			return el
		}
		t.Value = tags
		T().Debugf("TAG = %v", t.Value)
	default:
	}
	return el
}

func splitTagName(tagname string) ([]string, error) {
	var t []string
	if tagname == "" {
		return t, fmt.Errorf("empty tag name")
	}
	t = strings.Split(tagname, ".")
	return t, nil
}

type symbolPreservingResolver struct{}

func (r symbolPreservingResolver) Resolve(atom terex.Atom, env *terex.Environment, asOp bool) (
	terex.Element, error) {
	if atom.Type() == terex.TokenType {
		t := atom.Data.(*terex.Token)
		token := t.Token.(*lexmachine.Token)
		T().Debugf("Resolve terminal token: '%v'", string(token.Lexeme))
		switch token.Type {
		case tokenIds["NUM"]:
			return terex.Elem(t.Value.(float64)), nil
		case tokenIds["STRING"]:
			return terex.Elem(t.Value.(string)), nil
		}
	}
	return terex.Elem(atom), nil
}

var _ terex.SymbolResolver = symbolPreservingResolver{}
