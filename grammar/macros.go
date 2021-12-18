package grammar

type macro struct {
	variable    string
	replacement string
}

type macroStorage struct {
	macros []macro
}

type nestedReader struct {
	current []rune
}
