package grammar

import "github.com/npillmayer/gorgo/lr"

// created by TestGen(), please do not edit
func createGrammarRules(b *lr.GrammarBuilder) {
	b.LHS("statement_list").Epsilon()
	b.LHS("statement_list").N("statement").T(";", 59).N("statement_list").End()
	b.LHS("statement").Epsilon()
	b.LHS("statement").N("equation").End()
	b.LHS("statement").N("assignment").End()
	b.LHS("statement").N("declaration").End()
	b.LHS("statement").N("command").End()
	b.LHS("statement").N("macro_definition").End()
	b.LHS("statement").N("function_definition").End()
	b.LHS("statement").N("if_statement").End()
	b.LHS("statement").N("loop_statement").End()
	b.LHS("if_statement").T(S("if")).N("boolean_expression").T(":", 58).N("statement_list").N("alternatives").T(S("fi")).End()
	b.LHS("alternatives").Epsilon()
	b.LHS("alternatives").T(S("else:")).N("statement_list").End()
	b.LHS("alternatives").T(S("elseif")).N("boolean_expression").T(":", 58).N("statement_list").N("alternatives").End()
	b.LHS("loop").N("loop_header").T(":", 58).N("statement_list").T(S("endfor")).End()
	b.LHS("loop_header").T(S("for")).N("symbolic_token").T("=", 61).N("progression").End()
	b.LHS("loop_header").T(S("for")).N("symbolic_token").T("=", 61).N("for_list").End()
	b.LHS("loop_header").T(S("forsuffixes")).N("symbolic_token").T("=", 61).N("suffix_list").End()
	b.LHS("loop_header").T(S("forever")).End()
	b.LHS("progression").N("tertiary").T(S("upto")).N("tertiary").End()
	b.LHS("progression").N("tertiary").T(S("downto")).N("tertiary").End()
	b.LHS("progression").N("tertiary").T(S("step")).N("tertiary").T(S("until")).N("tertiary").End()
	b.LHS("for_list").N("tertiary").End()
	b.LHS("for_list").N("for_list").T(",", 44).N("tertiary").End()
	b.LHS("suffix_list").N("suffix").End()
	b.LHS("suffix_list").N("suffix_list").T(",", 44).N("suffix").End()
    // --- Expressions -----------------------------------------------------------
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
	b.LHS("primary").N("conditional_primary").End()
	b.LHS("atom").N("variable").End()
	b.LHS("atom").T(S("Unsigned")).N("variable").End()
	b.LHS("atom").T(S("Unsigned")).End()
	b.LHS("atom").T(S("NullaryOp")).End()
	b.LHS("atom").T(S("begingroup")).N("statement_list").N("tertiary").T(S("endgroup")).End()
	b.LHS("atom").N("function_call").End()
	b.LHS("atom").T("(", 40).N("tertiary").T(")", 41).End()
	b.LHS("transformer").T(S("UnaryTransform")).N("primary").End()
	b.LHS("transformer").T(S("BinaryTransform")).T("(", 40).N("tertiary").T(",", 44).N("tertiary").T(")", 41).End()
	b.LHS("conditional_primary").T(S("if")).N("boolean_expression").T(":", 58).N("primary").N("primary_alternatives").T(S("fi")).End()
	b.LHS("primary_alternatives").Epsilon()
	b.LHS("primary_alternatives").T(S("else:")).N("primary").End()
	b.LHS("primary_alternatives").T(S("elseif")).N("boolean_expression").T(":", 58).N("primary").N("primary_alternatives").End()
    // --- Paths -----------------------------------------------------------------
	b.LHS("path_expression").N("tertiary").End()
	b.LHS("path_expression").N("path_expression").N("path_join").N("path_knot").End()
	b.LHS("path_expression").N("path_expression").N("conditional_path_segment").End()
	b.LHS("path_expression").N("path_expression").N("segment_loop").End()
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
	b.LHS("conditional_path_segment").T(S("if")).N("boolean_expression").T(":", 58).N("path_expression").N("segment_alternatives").T(S("fi")).End()
	b.LHS("segment_alternatives").Epsilon()
	b.LHS("segment_alternatives").T(S("else:")).N("path_expression").End()
	b.LHS("segment_alternatives").T(S("elseif")).N("boolean_expression").T(":", 58).N("path_expression").N("segment_alternatives").End()
	b.LHS("segment_loop").N("loop_header").T(":", 58).N("path_expression").T(S("endfor")).End()
    // --- Function calls --------------------------------------------------------
	b.LHS("function_call").T(S("Function")).T("(", 40).N("tertiary_list").T(")", 41).End()
	b.LHS("variable").T(S("TAG")).N("suffix").End()
	b.LHS("suffix").Epsilon()
	b.LHS("suffix").N("suffix").N("subscript").End()
	b.LHS("suffix").N("suffix").T(S("TAG")).End()
	b.LHS("subscript").T(S("Unsigned")).End()
	b.LHS("subscript").T("[", 91).N("tertiary").T("]", 93).End()
    // --- Equations -------------------------------------------------------------
	b.LHS("equation").N("tertiary").T("=", 61).N("right_hand_side").End()
	b.LHS("assignment").N("variable").T(S(":=")).N("right_hand_side").End()
	b.LHS("right_hand_side").N("tertiary").End()
	b.LHS("right_hand_side").N("equation").End()
	b.LHS("right_hand_side").N("assignment").End()
    // --- Declarations ----------------------------------------------------------
	b.LHS("declaration").T(S("Type")).N("declaration_list").End()
	b.LHS("declaration_list").N("generic_variable").End()
	b.LHS("declaration_list").N("declaration_list").T(",", 44).N("generic_variable").End()
	b.LHS("generic_variable").T(S("TAG")).N("generic_suffix").End()
	b.LHS("generic_suffix").Epsilon()
	b.LHS("generic_suffix").N("generic_suffix").T(S("TAG")).End()
	b.LHS("generic_suffix").N("generic_suffix").T(S("[]")).End()
    // --- Commands --------------------------------------------------------------
	b.LHS("command").T(S("pickup")).N("primary").End()
	b.LHS("command").T(S("save")).N("symbolic_token_list").End()
	b.LHS("command").N("drawing_command").End()
	b.LHS("command").N("show_command").End()
	b.LHS("show_command").T(S("show")).N("tertiary").End()
	b.LHS("symbolic_token_list").T(S("TAG")).End()
	b.LHS("symbolic_token_list").T(S("SymTok")).End()
	b.LHS("symbolic_token_list").T(S("TAG")).T(",", 44).N("symbolic_token_list").End()
	b.LHS("symbolic_token_list").T(S("SymTok")).T(",", 44).N("symbolic_token_list").End()
	b.LHS("drawing_command").T(S("DrawCmd")).N("path_expression").N("option_list").End()
	b.LHS("option_list").Epsilon()
	b.LHS("option_list").N("drawing_option").N("option_list").End()
	b.LHS("drawing_option").T(S("DrawOption")).N("tertiary").End()
    // --- Macros ----------------------------------------------------------------
	b.LHS("macro_definition").T(S("def")).N("parameter").T("=", 61).End()
	b.LHS("function_definition").N("function_heading").T("=", 61).N("replacement_text").T(S("enddef")).End()
	b.LHS("replacement_text").N("statement_list").N("tertiary").End()
	b.LHS("function_heading").T(S("def")).N("parameter").N("delimited_part").N("undelimited_part").End()
	b.LHS("function_heading").T(S("vardef")).N("generic_variable").N("delimited_part").N("undelimited_part").End()
	b.LHS("function_heading").T(S("vardef")).N("generic_variable").T(S("SymTok")).N("delimited_part").N("undelimited_part").End()
	b.LHS("function_heading").N("binary_def").N("parameter").T(S("TAG")).N("parameter").End()
	b.LHS("delimited_part").Epsilon()
	b.LHS("delimited_part").N("delimited_part").T("(", 40).N("parameter_type").N("symbolic_token_list").T(")", 41).End()
	b.LHS("undelimited_part").Epsilon()
	b.LHS("undelimited_part").N("parameter_type").N("parameter").End()
	b.LHS("undelimited_part").N("precedence_level").N("parameter").End()
	b.LHS("undelimited_part").T(S("expr")).T(S("SymTok")).T(S("of")).N("parameter").End()
	b.LHS("precedence_level").T(S("primary")).End()
	b.LHS("precedence_level").T(S("secondary")).End()
	b.LHS("precedence_level").T(S("tertiary")).End()
	b.LHS("binary_def").T(S("primarydef")).End()
	b.LHS("binary_def").T(S("secondarydef")).End()
	b.LHS("binary_def").T(S("tertiarydef")).End()
	b.LHS("parameter_type").T(S("expr")).End()
	b.LHS("parameter_type").T(S("suffix")).End()
	b.LHS("parameter").T(S("TAG")).End()
	b.LHS("parameter").T(S("SymTok")).End()
}
