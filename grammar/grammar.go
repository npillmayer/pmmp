package grammar

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/npillmayer/gorgo"
	"github.com/npillmayer/pmmp"
	"github.com/npillmayer/pmmp/sframe"

	"github.com/npillmayer/gorgo/lr"
	"github.com/npillmayer/gorgo/lr/earley"
	"github.com/npillmayer/gorgo/lr/sppf"
	"github.com/npillmayer/gorgo/terex"
	"github.com/npillmayer/gorgo/terex/termr"
)

// --- Initialization --------------------------------------------------------

var startOnce sync.Once // monitors one-time creation of grammar and lexer

var mpGrammar *lr.LRAnalysis

//var mpLexer *scanner.LMAdapter

func initGlobalGrammar() {
	startOnce.Do(func() {
		mpGrammar = initGrammar("S")
	})
}

func initGrammar(startSymbol string) *lr.LRAnalysis {
	var err error
	var g *lr.LRAnalysis
	tracer().Infof("Creating MetaPost lexer")
	initTokens() // MUST be called before grammar builing !
	tracer().Infof("Creating MetaPost grammar with start-symbol %q", startSymbol)
	if g, err = MakeMetaPostGrammar(startSymbol); err != nil {
		panic("Cannot create global grammar")
	}
	initRewriters()
	return g
}

func createDefaultParser() *earley.Parser {
	initGlobalGrammar()
	return earley.NewParser(mpGrammar, earley.GenerateTree(true))
}

// createParser is mainly provided for testing
func createParser(startSymbol string) (*earley.Parser, *lr.LRAnalysis) {
	g := initGrammar(startSymbol)
	return earley.NewParser(g, earley.GenerateTree(true)), g
}

// ---------------------------------------------------------------------------

// MakeMetaPostGrammar generates a grammar for the MetaPost language,
// constructs all the parsing table for it and returns
// them as a package (or an error).
//
// This is usually not called directly be clients, but rather implicitely
// envoked through Parse().
//
func MakeMetaPostGrammar(startSymbol string) (*lr.LRAnalysis, error) {
	gramName := "MetaPost"
	if startSymbol != "" {
		gramName += "/[" + startSymbol + "]"
	}
	b := lr.NewGrammarBuilder(gramName)
	if startSymbol == "" {
		b.LHS("S").N("statement_list").T(";", 59).End()
	} else {
		b.LHS("S").N(startSymbol).End()
	}
	createGrammarRules(b)

	g, err := b.Grammar()
	if err != nil {
		tracer().Errorf("Error creating MetaPost grammar")
		return nil, err
	}
	g.Dump()
	ga := lr.Analysis(g)
	return ga, nil
}

// parseStatement parses an input string, given in MetaPost language format. It returns the
// parse forest and a TokenRetriever, or an error in case of failure.
//
// Clients may use a terex.ASTBuilder to create an abstract syntax tree
// from the parse forest.
//
func parseStatement(lex *lexer, scopes *sframe.ScopeFrameTree) (*sppf.Forest, gorgo.TokenRetriever, error) {
	parser := createDefaultParser()
	accept, err := parser.Parse(lex, nil)
	if err != nil {
		return nil, nil, err
	} else if !accept {
		return nil, nil, fmt.Errorf("not a valid MetaPost statement")
	}
	return parser.ParseForest(), earleyTokenReceiver(parser), nil
}

func earleyTokenReceiver(parser *earley.Parser) gorgo.TokenRetriever {
	return func(pos uint64) gorgo.Token {
		return parser.TokenAt(pos)
	}
}

// --- AST -------------------------------------------------------------------

var ErrGrammarNotInitialized = errors.New("MetaPost grammar not initialized")

// AST creates an abstract syntax tree from a parse tree/forest.
//
// Returns a homogenous AST, a TeREx-environment and an error status.
func AST(parsetree *sppf.Forest, tokRetr gorgo.TokenRetriever) (*terex.GCons,
	*terex.Environment, error) {
	//
	// grammar init should have been done already by Parse()
	if mpGrammar == nil || mpGrammar.Grammar() == nil {
		return nil, nil, ErrGrammarNotInitialized
	}
	return makeAST(mpGrammar, parsetree, tokRetr)
	// ab := newASTBuilder(mpGrammar.Grammar())
	// env := ab.AST(parsetree, tokRetr)
	// if env == nil {
	// 	tracer().Errorf("Cannot create AST from parsetree")
	// 	return nil, nil, errors.New("error while creating AST")
	// }
	// ast := env.AST
	// tracer().Infof("AST: %s", env.AST.ListString())
	// return ast, env, nil
}

func makeAST(g *lr.LRAnalysis, parsetree *sppf.Forest, tokRetr gorgo.TokenRetriever) (*terex.GCons,
	*terex.Environment, error) {
	//
	// grammar init should have been done already by parser
	if g == nil || g.Grammar() == nil {
		return nil, nil, ErrGrammarNotInitialized
	}
	ab := newASTBuilder(g.Grammar())
	env := ab.AST(parsetree, tokRetr)
	if env == nil {
		tracer().Errorf("Cannot create AST from parsetree")
		return nil, nil, errors.New("error while creating AST")
	}
	ast := env.AST
	tracer().Infof("AST: %s", env.AST.ListString())
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
//     // ast *terex.GCons
//     if env == nil {
//         env = terex.GlobalEnvironment
//     }
//     quEnv := terex.NewEnvironment("quoting", env)
//     quEnv.Defn("list", varOp.call)
//     //quEnv.Defn("quote", quoteOp.call)
//     quEnv.Resolver = symbolPreservingResolver{}
//     q := terex.Eval(ast, quEnv)
//     T().Debugf("QuotAST returns Q = %v", q)
//     q.Dump(tracing.LevelDebug)
//     return q, quEnv.LastError()
// }

// NewASTBuilder returns a new AST builder for the MetaPost language
func newASTBuilder(grammar *lr.Grammar) *termr.ASTBuilder {
	ab := termr.NewASTBuilder(grammar)
	ab.AddRewriter(atomOp.name, atomOp)
	ab.AddRewriter(varOp.name, varOp)
	ab.AddRewriter(suffixOp.name, suffixOp)
	ab.AddRewriter(subscrOp.name, subscrOp)
	ab.AddRewriter(primaryOp.name, primaryOp)
	ab.AddRewriter(secondaryOp.name, secondaryOp)
	ab.AddRewriter(tertiaryOp.name, tertiaryOp)
	ab.AddRewriter(exprOp.name, exprOp)
	ab.AddRewriter(declOp.name, declOp)
	ab.AddRewriter(declvarOp.name, declvarOp)
	ab.AddRewriter(declsuffixOp.name, declsuffixOp)
	ab.AddRewriter(eqOp.name, eqOp)
	ab.AddRewriter(assignOp.name, assignOp)
	ab.AddRewriter(transformOp.name, transformOp)
	ab.AddRewriter(funcallOp.name, funcallOp)
	ab.AddRewriter(tertiaryListOp.name, tertiaryListOp)
	ab.AddRewriter(stmtOp.name, stmtOp)
	ab.AddRewriter(stmtListOp.name, stmtListOp)
	ab.AddRewriter(basicJoinOp.name, basicJoinOp)
	ab.AddRewriter(tensionOp.name, tensionOp)
	ab.AddRewriter(controlsOp.name, controlsOp)
	ab.AddRewriter(dirOp.name, dirOp)
	ab.AddRewriter(joinOp.name, joinOp)
	ab.AddRewriter(pathExprOp.name, pathExprOp)
	ab.AddRewriter(commandOp.name, commandOp)
	ab.AddRewriter(drawOptOp.name, drawOptOp)
	return ab
}

// === Terminal tokens =======================================================

// TODO change signature to func(atom) => atom
// env not needed

func setTerminalTokenValue(el terex.Element, env *terex.Environment) terex.Element {
	if !el.IsAtom() {
		return el
	}
	atom := el.AsAtom()
	if atom.Type() != terex.TokenType {
		return el
	}
	token := atom.Data.(MPToken)
	tracer().Infof("setting value of terminal token: '%v'", string(token.Lexeme()))
	switch token.TokType() {
	/*
		case tokenTypeFromLexeme["NUMBER"]:
			lexeme := string(token.Lexeme)
			if strings.ContainsRune(lexeme, rune('/')) {
				frac := strings.Split(lexeme, "/")
				nom, _ := strconv.ParseFloat(frac[0], 64)
				denom, _ := strconv.ParseFloat(frac[1], 64)
				t.Value = nom / denom
			} else if f, err := strconv.ParseFloat(string(token.Lexeme), 64); err == nil {
				tracer().Debugf("   t.Value=%g", f)
				t.Value = f
			} else {
				tracer().Errorf("   %s", err.Error())
				return terex.Elem(terex.Atomize(err))
			}
		case tokenTypeFromLexeme["STRING"]:
			if (len(token.Lexeme)) <= 2 {
				t.Value = ""
			} else { // trim off "…" ⇒ …
				t.Value = string(token.Lexeme[1 : len(token.Lexeme)-1])
			}
	*/
	case tokenTypeFromLexeme["TAG"]: // return []string value, split at '.'
		tag := string(token.Lexeme())
		tags, err := splitTagName(tag)
		if err != nil {
			tracer().Errorf("illegal tag name")
			token.Val = []string{"<illegal tag>"}
			return el
		}
		token.Val = tags
		tracer().Infof("TAG = %v  ⇒ %v", token.Value(), token)
	default:
		token.Val = string(token.Lexeme())
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
//     terminalToken *terex.Token
// }

// func (w wrapOp) String() string {
//     // will result in "##<opname>:<op-category>"
//     return "#" + w.Opname() + ":" + w.terminalToken.Name
// }

// func (w wrapOp) Opname() string {
//     return w.terminalToken.Value.(string)
// }

func wrapOpToken(a terex.Atom) terex.Operator {
	a = setTerminalTokenValue(terex.Elem(a), nil).AsAtom()
	if a.Data == nil || a.Data.(MPToken).Value() == nil {
		panic(fmt.Sprintf("value of token-atom is nil, not an operator name"))
	}
	token := a.Data.(MPToken)
	return pmmp.NewTokenOperator(token)
}

// Call delegates the operator call to a symbol in the environment.
// The symbol is searched for with the literal value of the operator.
// func (w wrapOp) Call(e terex.Element, env *terex.Environment) terex.Element {
//     return callFromEnvironment(w.Opname(), e, env)
// }

// var _ terex.Operator = &wrapOp{}

// func callFromEnvironment(opname string, e terex.Element, env *terex.Environment) terex.Element {
//     opsym := env.FindSymbol(opname, true)
//     if opsym == nil {
//         T().Errorf("Cannot find parsing operation %s", opname)
//         return e
//     }
//     operator, ok := opsym.Value.AsAtom().Data.(terex.Operator)
//     if !ok {
//         T().Errorf("Cannot call parsing operation %s", opname)
//         return e
//     }
//     return operator.Call(e, env)
// }

// --- Symbol resolver -------------------------------------------------------

// TODO This is probably unnecessary

// type symbolPreservingResolver struct{}

// func (r symbolPreservingResolver) Resolve(atom terex.Atom, env *terex.Environment, asOp bool) (
//     terex.Element, error) {
//     if atom.Type() == terex.TokenType {
//         t := atom.Data.(*terex.Token)
//         token := t.Token.(*lexmachine.Token)
//         T().Debugf("Resolve terminal token: '%v'", string(token.Lexeme))
//         switch token.Type {
//         case tokenIds["NUM"]:
//             return terex.Elem(t.Value.(float64)), nil
//         case tokenIds["STRING"]:
//             return terex.Elem(t.Value.(string)), nil
//         }
//     }
//     return terex.Elem(atom), nil
// }

// var _ terex.SymbolResolver = symbolPreservingResolver{}
