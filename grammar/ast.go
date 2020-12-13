package grammar

import (
	"github.com/npillmayer/gorgo/lr/sppf"
	"github.com/npillmayer/gorgo/terex"
	"github.com/npillmayer/gorgo/terex/termr"
	"github.com/npillmayer/schuko/tracing"
)

// --- Parse tree to AST rewriters -------------------------------------------

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
	panic("pmmp term rewriter not intended to be 'call'ed")
	//return callFromEnvironment(trew.opname, e, env)
}

// --- Init global rewriters -------------------------------------------------

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
		if singleArg(l) { // ⟨variable⟩
			if keywordArg(l) { // NUMBER | NullaryOp
				convertTerminalToken(terex.Elem(l.Cdar()), env)
				return terex.Elem(l.Cdar())
			}
			return terex.Elem(l.Cdar())
		}
		if tokenArg(l) { // NUMBER ⟨variable⟩ ⇒ (* NUMBER ⟨variable⟩ )
			// invent an ad-hoc multiplication token
			op := wrapOpToken(terex.Atomize(makeLMToken("PrimaryOp", "*")))
			prefix := convertTerminalToken(terex.Elem(l.Cdar()), env)
			return terex.Elem(terex.List(op, prefix.AsAtom(), l.Cddar()))
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
		//     | ⟨scalar multiplication op⟩  ⟨primary⟩
		//     | ( ⟨numeric expression⟩ , ⟨numeric expression⟩ )
		T().Infof("primary tree = ")
		terex.Elem(l).Dump(tracing.LevelInfo)
		if !singleArg(l) {
			if l.Length() == 3 {
				// ⟨primary⟩ → UnaryOp ⟨primary⟩
				// ⟨primary⟩ → ⟨scalar multiplication op⟩  ⟨primary⟩
				opAtom := terex.Atomize(wrapOpToken(l.Cdar()))
				return terex.Elem(terex.Cons(opAtom, l.Cddr())) // UnaryOp ⟨primary⟩
			}
			// ⟨primary⟩ → ( ⟨numeric expression⟩ , ⟨numeric expression⟩ )
			// create a pseudo-op for pair creation
			//op:=makePseudoOp("")
			op := wrapOpToken(terex.Atomize(makeLMToken("PseudoOp", "make-pair")))
			return terex.Elem(terex.List(op, l.Cddar(), l.Nth(5)))
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
		// ⟨tertiary⟩ → ⟨secondary⟩
		//     | ⟨tertiary⟩  SecondaryOp  ⟨secondary⟩
		//     | ⟨tertiary⟩  PlusOrMinus  ⟨secondary⟩
		T().Infof("tertiary tree = ")
		terex.Elem(l).Dump(tracing.LevelInfo)
		if singleArg(l) {
			return terex.Elem(l.Cdar()) // ⟨tertiary⟩ → ⟨secondary⟩
		}
		// ⟨tertiary⟩ SecondaryOp ⟨secondary⟩ ⇒ ( SecondaryOp ⟨tertiary⟩ ⟨secondary⟩ )
		op := l.Cddar().Data.(*terex.Token)
		op.Name = "SecondaryOp" // PlusOrMinus ⇒ SecondaryOp
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

// ---------------------------------------------------------------------------

// WithoutArgs is a predicate: are there no arguments?
func withoutArgs(l *terex.GCons) bool {
	return l == nil || l.Length() <= 1
}

// SingleArg is a predicate: is there a single argument?
func singleArg(l *terex.GCons) bool {
	return l != nil && l.Length() == 2
}

// TokenArg is a predicate: is the argument a token?
func tokenArg(l *terex.GCons) bool {
	return l != nil && l.Length() > 1 && l.Cdar().Type() == terex.TokenType
}

// KeywordArg is a predicate: is the argument a token and its lexeme at
// least 2 characters long?
func keywordArg(l *terex.GCons) bool {
	if !tokenArg(l) {
		return false
	}
	tokname := l.Cdar().Data.(*terex.Token).Name
	return len(tokname) > 1 // keywords have at least 2 letters
}

// ---Tag part splitting -----------------------------------------------------

// TAG is given by the scanner as
//
//     tag ( "." tag )*
//
// This is done to get fewer prefixes and therefore slightly simpler parse trees.
// We have to split these up into a list of suffixes.
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

// ---------------------------------------------------------------------------

type mpPseudoOp struct {
	name string
	//call func(terex.Element, *terex.Environment) terex.Element
}

func makePseudoOp(s string) mpPseudoOp {
	return mpPseudoOp{name: s}
}

var _ terex.Operator = &mpTermR{}

func (po *mpPseudoOp) String() string {
	return po.name
}

func (po *mpPseudoOp) Call(e terex.Element, env *terex.Environment) terex.Element {
	T().Errorf("PseudoOperator not to be called")
	return terex.Elem(nil)
}
