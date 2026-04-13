package nameparser_test

import (
	"testing"

	"github.com/ysmood/nameparser"
)

func FuzzParse(f *testing.F) {
	seeds := []string{
		"",
		"John Smith",
		"Dr. John A. Kenneth Doe, Jr.",
		"MD de la Cruz MD",
		"Jr John de la Cruz Jr",
		"II de Cruz II",
		"de la Cruz, Ana",
		"Smith, Jr, Jr",
		"Dr. and Mrs. Smith",
		"(Johnny) John Smith",
		",,,",
		"y y y",
		"V V V V",
		",", ",,", " , ", "  ", "\t",
		"and", "and and", "y y y",
		"de", "de la", "de, de",
		"Jr,", ",Jr",
		"Dr. and Mrs.", "Mr. and", "and Mr.",
		"a b c, d e f, g h i",
		"Smith, Jr, Jr",
		"Jr Jr Jr",
		"de de de",
		"V V V V",
		"Dr Mr Sr Jr",
		"(nickname) and (other)",
		"\"\"",
		"John (",
		"( John )",
		"a,b,c,d,e,f",
		"y",
		"a y",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		h := nameparser.New(input)
		_ = h.First()
		_ = h.Last()
		_ = h.Middle()
		_ = h.Suffix()
		_ = h.Title()
		_ = h.Nickname()
		_ = h.InitialsList()
		_ = h.String()
	})
}
