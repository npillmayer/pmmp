package generator

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/text/width"
)

var G bytes.Buffer

func TestGen(t *testing.T) {
	G = bytes.Buffer{}
	scanner, file := scannerFor("pmmp-grammar.txt")
	defer file.Close()
	out, gofile := makeOutputFile("../G.go")
	defer gofile.Close()
	out.WriteString("package grammar\n\n")
	out.WriteString(`import "github.com/npillmayer/gorgo/lr"`)
	out.WriteString("\n\n// created by TestGen(), please do not edit\n")
	out.WriteString("func createGrammarRules(b *lr.GrammarBuilder) {\n")
	initSymbols()
	var lhs string
	for scanner.Scan() {
		//fmt.Println(scanner.Text())
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		} else if strings.HasPrefix(line, "STOP") {
			break
		} else if strings.HasPrefix(line, "//") {
			out.WriteString("    " + line + "\n")
			continue
		}
		var rule string
		if strings.HasPrefix(line, "⟨") {
			a := strings.Split(line, "→")
			lhs = a[0] //[1 : len(a[0])-2]
			//t.Logf("lhs sym = %s", lhs)
			rule = makeRule(lhs, a[1])
		} else if strings.HasPrefix(line, "|") {
			//t.Logf("repeat")
			rule = makeRule(lhs, line[1:])
		}
		out.WriteString(rule + "\n")
		out.Flush()
	}
	if err := scanner.Err(); err != nil {
		t.Error(err)
	}
	out.WriteString("}\n")
	out.Flush()
	gofile.Close()
}

func makeRule(lhs string, rhs string) string {
	var r string
	//lhs = renameNonterms("⟨" + lhs + "⟩")
	lhs = renameNonterms(lhs)
	r = fmt.Sprintf("\tb.LHS(\"%s\")", lhs[3:len(lhs)-4])
	rhs = renameNonterms(rhs)
	//fmt.Printf("rhs is now %s\n", rhs)
	spl := strings.Split(rhs, " ")
	hasEps := false
	for _, x := range spl {
		x = strings.TrimSpace(x)
		if strings.HasPrefix(x, "⟨") { // non-terminal
			x = x[3 : len(x)-3]
			if x == "empty" {
				r = r + ".Epsilon()"
				hasEps = true
			} else {
				r = r + fmt.Sprintf(".N(\"%s\")", x)
			}
			// } else if strings.HasPrefix(x, "$") { // $keyword
			// 	r = r + fmt.Sprintf(`.T("$",%s)`, x[1:])
		} else { // terminal
			if len(x) == 0 {
				continue
			}
			var t string
			if len(x) == 1 {
				t = fmt.Sprintf(".T(\"%s\", %d)", x, int(x[0]))
			} else {
				//t = fmt.Sprintf(".T(\"%s\", %d)", x, tokval(x))
				t = fmt.Sprintf(".T(S(\"%s\"))", x)
			}
			r = r + t
		}
	}
	if !hasEps {
		r = r + ".End()"
	}
	return r
}

var delsp1 *regexp.Regexp = regexp.MustCompile("⟨([a-z]+) ([a-z]+)⟩")
var delsp2 *regexp.Regexp = regexp.MustCompile("⟨([a-z]+) ([a-z]+) ([a-z]+)⟩")

func renameNonterms(s string) string {
	//fmt.Printf("MATCH: %v \n", delsp1.MatchString(s))
	u1 := "⟨${1}_${2}⟩"
	res := delsp1.ReplaceAllString(s, u1)
	u2 := "⟨${1}_${2}_${3}⟩"
	res = delsp2.ReplaceAllString(res, u2)
	return res
}

var tokmap map[string]int = make(map[string]int)
var tokcnt int = 20

func tokval(t string) int {
	if v, ok := tokmap[t]; ok {
		return v
	}
	tokcnt++
	v := tokcnt
	tokmap[t] = v
	return v
}

func scannerFor(path string) (*bufio.Scanner, *os.File) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	return scanner, file
}

func makeOutputFile(path string) (*bufio.Writer, *os.File) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	writer := bufio.NewWriter(file)
	return writer, file
}

func initSymbols() {
	_, w1 := width.LookupString("⟨")
	_, w2 := width.LookupString("⟩")
	fmt.Printf("rune width ⟨ = %d, ⟩ = %d\n", w1, w2)
	tokmap["TAG"] = -2
	tokmap["NUMBER"] = -4
}

func NoTestM(t *testing.T) {
	s := "| btex ⟨typesetting commands⟩ etex"
	fmt.Printf("m => %s\n", renameNonterms(s))
}
