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

	b.LHS("expression").N("subexpression").End()
	b.LHS("expression").N("expression").T(S("RelationOp")).N("tertiary").End()
	b.LHS("subexpression").N("tertiary").End()
	b.LHS("tertiary").N("secondary").End()
	b.LHS("tertiary").N("tertiary").T(S("SecondaryOp")).N("secondary").End()
	b.LHS("secondary").N("primary").End()
	b.LHS("secondary").N("secondary").T(S("PrimaryOp")).N("primary").End()
	b.LHS("primary").N("atom").End()
	b.LHS("primary").T(S("UnaryOp")).N("primary").End()
	b.LHS("atom").N("variable").End()
	b.LHS("atom").T(S("NUMBER")).End()
	b.LHS("atom").T(S("NullaryOp")).End()
	b.LHS("atom").T("(", 40).N("expression").T(")", 41).End()
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
	ab.AddTermR(atomOp)
	ab.AddTermR(varOp)
	ab.AddTermR(suffixOp)
	ab.AddTermR(subscrOp)
	ab.AddTermR(primaryOp)
	ab.AddTermR(secondaryOp)
	ab.AddTermR(tertiaryOp)
	ab.AddTermR(exprOp)
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
	return callFromEnvironment(trew.opname, e, env)
}

var atomOp *mpTermR      // for atom -> ... productions
var varOp *mpTermR       // for variable -> ... productions
var suffixOp *mpTermR    // for suffix -> ... productions
var subscrOp *mpTermR    // for subscript -> ... productions
var primaryOp *mpTermR   // for primary -> ... productions
var secondaryOp *mpTermR // for secondary -> ... productions
var tertiaryOp *mpTermR  // for tertiary -> ... productions
var exprOp *mpTermR      // for expression -> ... productions

func initRewriters() {
	atomOp = makeASTTermR("atom", "atom")
	atomOp.rewrite = func(l *terex.GCons, env *terex.Environment) terex.Element {
		// ⟨atom⟩ → ⟨variable⟩ | NUMBER | NullaryOp
		//     | ( ⟨expression⟩ )
		T().Infof("atom tree = ")
		terex.Elem(l).Dump(tracing.LevelInfo)
		if keywordArg(l) { // NUMBER | NullaryOp
			convertTerminalToken(terex.Elem(l.Cdar()), env)
			return terex.Elem(l.Cdar())
		} else if singleArg(l) { // ⟨variable⟩
			return terex.Elem(l.Cdar())
		}
		return terex.Elem(l.Cddar()) // ( ⟨expression⟩ ) ⇒ ⟨expression⟩
	}
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
	primaryOp = makeASTTermR("primary", "primary")
	primaryOp.rewrite = func(l *terex.GCons, env *terex.Environment) terex.Element {
		// ⟨primary⟩ → ⟨atom⟩ | UnaryOp ⟨primary⟩
		T().Infof("primary tree = ")
		terex.Elem(l).Dump(tracing.LevelInfo)
		if !singleArg(l) { // ⟨primary⟩ → UnaryOp ⟨primary⟩
			opAtom := terex.Atomize(wrapOpToken(l.Cdar()))
			return terex.Elem(terex.Cons(opAtom, l.Cddr())) // UnaryOp ⟨primary⟩
		}
		return terex.Elem(l.Cdar()) // ⟨primary⟩ → ⟨atom⟩
	}
	secondaryOp = makeASTTermR("secondary", "secondary")
	secondaryOp.rewrite = func(l *terex.GCons, env *terex.Environment) terex.Element {
		// ⟨secondary⟩ → ⟨primary⟩ | ⟨secondary⟩ PrimaryOp ⟨primary⟩
		T().Infof("secondary tree = ")
		terex.Elem(l).Dump(tracing.LevelInfo)
		if singleArg(l) {
			return terex.Elem(l.Cdar()) // ⟨secondary⟩ → ⟨primary⟩
		}
		// ⟨secondary⟩ PrimaryOp ⟨primary⟩ ⇒ ( PrimaryOp ⟨secondary⟩ ⟨primary⟩ )
		opAtom := terex.Atomize(wrapOpToken(l.Cddar()))
		c := terex.Cons(opAtom, terex.Cons(l.Cdar(), l.Last()))
		return terex.Elem(c)
	}
	tertiaryOp = makeASTTermR("tertiary", "tertiary")
	tertiaryOp.rewrite = func(l *terex.GCons, env *terex.Environment) terex.Element {
		// ⟨tertiary⟩ → ⟨secondary⟩ | ⟨tertiary⟩  ⟨secondary binop⟩  ⟨secondary⟩
		T().Infof("tertiary tree = ")
		terex.Elem(l).Dump(tracing.LevelInfo)
		if singleArg(l) {
			return terex.Elem(l.Cdar()) // ⟨tertiary⟩ → ⟨secondary⟩
		}
		// ⟨tertiary⟩ SecondaryOp ⟨secondary⟩ ⇒ ( SecondaryOp ⟨tertiary⟩ ⟨secondary⟩ )
		opAtom := terex.Atomize(wrapOpToken(l.Cddar()))
		c := terex.Cons(opAtom, terex.Cons(l.Cdar(), l.Last()))
		return terex.Elem(c)
	}
	exprOp = makeASTTermR("expression", "expr")
	exprOp.rewrite = func(l *terex.GCons, env *terex.Environment) terex.Element {
		// ⟨expression⟩ → ⟨subexpression⟩ | ⟨expression⟩  RelationOp  ⟨tertiary⟩
		T().Infof("expression tree = ")
		terex.Elem(l).Dump(tracing.LevelInfo)
		if singleArg(l) {
			return terex.Elem(l.Cdar()) // ⟨expression⟩ → ⟨subexpression⟩
		}
		// ⟨expression⟩ RelationOp ⟨tertiary⟩ ⇒ ( RelationOp ⟨expression⟩ ⟨tertiary⟩ )
		opAtom := terex.Atomize(wrapOpToken(l.Cddar()))
		c := terex.Cons(opAtom, terex.Cons(l.Cdar(), l.Last()))
		return terex.Elem(c)
	}
}

func withoutArgs(l *terex.GCons) bool {
	return l == nil || l.Length() <= 1
}

func singleArg(l *terex.GCons) bool {
	return l != nil && l.Length() == 2
}

func tokenArg(l *terex.GCons) bool {
	return l != nil && l.Length() > 1 && l.Cdar().Type() == terex.TokenType
}

func keywordArg(l *terex.GCons) bool {
	if !tokenArg(l) {
		return false
	}
	tokname := l.Cdar().Data.(*terex.Token).Name
	return len(tokname) > 1 // keyword have at least 2 letters
}

// TAG is given by the scanner as
//
//     tag ( "." tag )*
//
// This is done to get fewer prefixes and therefore slightly simpler parse trees.
// We have to split these up into a list of suffixes.
// TODO Ignore leading and trailing dots, allow for ', e.g. a.r'
//
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

// TODO change signature to func(atom) => atom
// env not needed

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
		lexeme := string(token.Lexeme)
		if strings.ContainsRune(lexeme, rune('/')) {
			frac := strings.Split(lexeme, "/")
			nom, _ := strconv.ParseFloat(frac[0], 64)
			denom, _ := strconv.ParseFloat(frac[1], 64)
			t.Value = nom / denom
		} else if f, err := strconv.ParseFloat(string(token.Lexeme), 64); err == nil {
			T().Debugf("   t.Value=%g", f)
			t.Value = f
		} else {
			T().Errorf("   %s", err.Error())
			return terex.Elem(terex.Atomize(err))
		}
	case tokenIds["STRING"]:
		if (len(token.Lexeme)) <= 2 {
			t.Value = ""
		} else { // trim off "…" ⇒ …
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
	case tokenIds["NullaryOp"]:
		fallthrough
	case tokenIds["UnaryOp"]:
		fallthrough
	case tokenIds["PrimaryOp"]:
		fallthrough
	case tokenIds["SecondaryOp"]:
		fallthrough
	case tokenIds["RelationOp"]:
		t.Value = string(token.Lexeme)
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

// --- Operator wrapper ------------------------------------------------------

type wrapOp struct {
	terminalToken *terex.Token
}

func (w wrapOp) String() string {
	return "#" + w.Opname() + ":" + w.terminalToken.Name
	// will result in "##<opname>:<op-category>"
}

func (w wrapOp) Opname() string {
	return w.terminalToken.Value.(string)
}

func wrapOpToken(a terex.Atom) terex.Operator {
	a = convertTerminalToken(terex.Elem(a), nil).AsAtom()
	if a.Data == nil {
		tokname := a.Data.(*terex.Token).Name
		panic(fmt.Sprintf("value of token '%s' is nil, not operator name", tokname))
	}
	tok := a.Data.(*terex.Token)
	return wrapOp{terminalToken: tok}
}

// Call delegates the operator call to a symbol in the environment.
// The symbol is searched for with the literal value of the operator.
func (w wrapOp) Call(e terex.Element, env *terex.Environment) terex.Element {
	return callFromEnvironment(w.Opname(), e, env)
}

var _ terex.Operator = &wrapOp{}

func callFromEnvironment(opname string, e terex.Element, env *terex.Environment) terex.Element {
	opsym := env.FindSymbol(opname, true)
	if opsym == nil {
		T().Errorf("Cannot find parsing operation %s", opname)
		return e
	}
	operator, ok := opsym.Value.AsAtom().Data.(terex.Operator)
	if !ok {
		T().Errorf("Cannot call parsing operation %s", opname)
		return e
	}
	return operator.Call(e, env)
}

// --- Symbol resolver -------------------------------------------------------

// TODO This is probably unnecessary

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
