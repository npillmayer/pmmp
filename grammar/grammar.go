package grammar

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/npillmayer/pmmp"

	"github.com/npillmayer/gorgo/lr"
	"github.com/npillmayer/gorgo/lr/earley"
	"github.com/npillmayer/gorgo/lr/scanner"
	"github.com/npillmayer/gorgo/lr/sppf"
	"github.com/npillmayer/gorgo/terex"
	"github.com/npillmayer/gorgo/terex/termr"
	"github.com/timtadh/lexmachine"
)

// --- Initialization --------------------------------------------------------

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

	b.LHS("S").N("statement_list").End()
	b.LHS("statement_list").Epsilon()
	b.LHS("statement_list").N("statement").T(";", 59).N("statement_list").End()
	b.LHS("statement").Epsilon()
	b.LHS("statement").N("equation").End()
	b.LHS("statement").N("assignment").End()
	b.LHS("statement").N("declaration").End()
	b.LHS("statement").N("path_expression").End()
	b.LHS("boolean_expression").N("tertiary").T(S("RelationOp")).N("tertiary").End()
	b.LHS("tertiary_list").N("tertiary").End()
	b.LHS("tertiary_list").N("tertiary_list").T(",", 44).N("tertiary").End()
	b.LHS("tertiary").N("secondary").End()
	b.LHS("tertiary").N("tertiary").T(S("SecondaryOp")).N("secondary").End()
	b.LHS("tertiary").N("tertiary").T(S("PlusOrMinus")).N("secondary").End()
	b.LHS("secondary").N("primary").End()
	b.LHS("secondary").N("secondary").T(S("PrimaryOp")).N("primary").End()
	b.LHS("secondary").N("secondary").N("transformer").End()
	b.LHS("primary").N("atom").End()
	b.LHS("primary").T("(", 40).N("tertiary").T(",", 44).N("tertiary").T(")", 41).End()
	b.LHS("primary").T(S("UnaryOp")).N("primary").End()
	b.LHS("primary").T(S("PlusOrMinus")).N("primary").End()
	b.LHS("primary").T(S("OfOp")).N("tertiary").T(S("of")).N("primary").End()
	b.LHS("primary").N("atom").T("[", 91).N("tertiary").T(",", 44).N("tertiary").T("]", 93).End()
	b.LHS("atom").N("variable").End()
	b.LHS("atom").T(S("Unsigned")).N("variable").End()
	b.LHS("atom").T(S("Unsigned")).End()
	b.LHS("atom").T(S("NullaryOp")).End()
	b.LHS("atom").T(S("begingroup")).N("statement_list").N("tertiary").T(S("endgroup")).End()
	b.LHS("atom").N("function_call").End()
	b.LHS("atom").T("(", 40).N("tertiary").T(")", 41).End()
	b.LHS("transformer").T(S("UnaryTransform")).N("primary").End()
	b.LHS("transformer").T(S("BinaryTransform")).T("(", 40).N("tertiary").T(",", 44).N("tertiary").T(")", 41).End()
	b.LHS("path_expression").N("tertiary").End()
	b.LHS("path_expression").N("path_expression").N("path_join").N("path_knot").End()
	b.LHS("path_knot").N("tertiary").End()
	b.LHS("path_join").T(S("--")).End()
	b.LHS("path_join").N("direction_specifier").N("basic_path_join").N("direction_specifier").End()
	b.LHS("direction_specifier").Epsilon()
	b.LHS("direction_specifier").T("{", 123).T(S("curl")).N("tertiary").T("}", 125).End()
	b.LHS("direction_specifier").T("{", 123).N("tertiary").T("}", 125).End()
	b.LHS("direction_specifier").T("{", 123).N("tertiary").T(",", 44).N("tertiary").T("}", 125).End()
	b.LHS("basic_path_join").T(S("..")).End()
	b.LHS("basic_path_join").T(S("...")).End()
	b.LHS("basic_path_join").T(S("..")).N("tension").T(S("..")).End()
	b.LHS("basic_path_join").T(S("..")).N("controls").T(S("..")).End()
	b.LHS("tension").T(S("tension")).N("primary").End()
	b.LHS("tension").T(S("tension")).N("primary").T(S("and")).N("primary").End()
	b.LHS("controls").T(S("controls")).N("primary").End()
	b.LHS("controls").T(S("controls")).N("primary").T(S("and")).N("primary").End()
	b.LHS("function_call").T(S("Function")).T("(", 40).N("tertiary_list").T(")", 41).End()
	b.LHS("variable").T(S("TAG")).N("suffix").End()
	b.LHS("suffix").Epsilon()
	b.LHS("suffix").N("suffix").N("subscript").End()
	b.LHS("suffix").N("suffix").T(S("TAG")).End()
	b.LHS("subscript").T(S("Unsigned")).End()
	b.LHS("subscript").T("[", 91).N("tertiary").T("]", 93).End()
	b.LHS("equation").N("tertiary").T("=", 61).N("right_hand_side").End()
	b.LHS("assignment").N("variable").T(S(":=")).N("right_hand_side").End()
	b.LHS("right_hand_side").N("tertiary").End()
	b.LHS("right_hand_side").N("equation").End()
	b.LHS("right_hand_side").N("assignment").End()
	b.LHS("declaration").T(S("Type")).N("declaration_list").End()
	b.LHS("declaration_list").N("generic_variable").End()
	b.LHS("declaration_list").N("declaration_list").T(",", 44).N("generic_variable").End()
	b.LHS("generic_variable").T(S("TAG")).N("generic_suffix").End()
	b.LHS("generic_suffix").Epsilon()
	b.LHS("generic_suffix").N("generic_suffix").T(S("TAG")).End()
	b.LHS("generic_suffix").N("generic_suffix").T(S("[]")).End()

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
// func QuoteAST(ast terex.Element, env *terex.Environment) (terex.Element, error) {
// 	// ast *terex.GCons
// 	if env == nil {
// 		env = terex.GlobalEnvironment
// 	}
// 	quEnv := terex.NewEnvironment("quoting", env)
// 	quEnv.Defn("list", varOp.call)
// 	//quEnv.Defn("quote", quoteOp.call)
// 	quEnv.Resolver = symbolPreservingResolver{}
// 	q := terex.Eval(ast, quEnv)
// 	T().Debugf("QuotAST returns Q = %v", q)
// 	q.Dump(tracing.LevelDebug)
// 	return q, quEnv.LastError()
// }

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
	ab.AddTermR(declOp)
	ab.AddTermR(declvarOp)
	ab.AddTermR(declsuffixOp)
	ab.AddTermR(eqOp)
	ab.AddTermR(assignOp)
	ab.AddTermR(transformOp)
	ab.AddTermR(funcallOp)
	ab.AddTermR(tertiaryListOp)
	ab.AddTermR(stmtOp)
	ab.AddTermR(stmtListOp)
	ab.AddTermR(basicJoinOp)
	ab.AddTermR(tensionOp)
	ab.AddTermR(controlsOp)
	ab.AddTermR(dirOp)
	ab.AddTermR(joinOp)
	ab.AddTermR(pathExprOp)
	return ab
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
	default:
		t.Value = string(token.Lexeme)
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

// type wrapOp struct {
// 	terminalToken *terex.Token
// }

// func (w wrapOp) String() string {
// 	// will result in "##<opname>:<op-category>"
// 	return "#" + w.Opname() + ":" + w.terminalToken.Name
// }

// func (w wrapOp) Opname() string {
// 	return w.terminalToken.Value.(string)
// }

func wrapOpToken(a terex.Atom) terex.Operator {
	a = convertTerminalToken(terex.Elem(a), nil).AsAtom()
	if a.Data == nil || a.Data.(*terex.Token).Value == nil {
		tokname := a.Data.(*terex.Token).Name
		panic(fmt.Sprintf("value of token '%s' is nil, not operator name", tokname))
	}
	tok := a.Data.(*terex.Token)
	return pmmp.NewTokenOperator(tok)
}

// Call delegates the operator call to a symbol in the environment.
// The symbol is searched for with the literal value of the operator.
// func (w wrapOp) Call(e terex.Element, env *terex.Environment) terex.Element {
// 	return callFromEnvironment(w.Opname(), e, env)
// }

// var _ terex.Operator = &wrapOp{}

// func callFromEnvironment(opname string, e terex.Element, env *terex.Environment) terex.Element {
// 	opsym := env.FindSymbol(opname, true)
// 	if opsym == nil {
// 		T().Errorf("Cannot find parsing operation %s", opname)
// 		return e
// 	}
// 	operator, ok := opsym.Value.AsAtom().Data.(terex.Operator)
// 	if !ok {
// 		T().Errorf("Cannot call parsing operation %s", opname)
// 		return e
// 	}
// 	return operator.Call(e, env)
// }

// --- Symbol resolver -------------------------------------------------------

// TODO This is probably unnecessary

// type symbolPreservingResolver struct{}

// func (r symbolPreservingResolver) Resolve(atom terex.Atom, env *terex.Environment, asOp bool) (
// 	terex.Element, error) {
// 	if atom.Type() == terex.TokenType {
// 		t := atom.Data.(*terex.Token)
// 		token := t.Token.(*lexmachine.Token)
// 		T().Debugf("Resolve terminal token: '%v'", string(token.Lexeme))
// 		switch token.Type {
// 		case tokenIds["NUM"]:
// 			return terex.Elem(t.Value.(float64)), nil
// 		case tokenIds["STRING"]:
// 			return terex.Elem(t.Value.(string)), nil
// 		}
// 	}
// 	return terex.Elem(atom), nil
// }

// var _ terex.SymbolResolver = symbolPreservingResolver{}
