package nameparser

import (
	"regexp"
	"strings"
)

const DefaultEncoding = "UTF-8"

var (
	reSpaces         = regexp.MustCompile(`\s+`)
	reWord           = regexp.MustCompile(`([\p{L}\p{N}_]|\.)+`)
	reMac            = regexp.MustCompile(`(?i)^(ma?c)([\p{L}\p{N}_]{2,})`)
	reInitial        = regexp.MustCompile(`^([\p{L}\p{N}_]\.|[A-Z])?$`)
	reQuotedWord     = regexp.MustCompile(`(^|[^\p{L}\p{N}_])'([^\s]*?)'([^\p{L}\p{N}_]|$)`)
	reDoubleQuotes   = regexp.MustCompile(`"(.*?)"`)
	reParenthesis    = regexp.MustCompile(`\((.*?)\)`)
	reRomanNumeral   = regexp.MustCompile(`(?i)^(X|IX|IV|V?I{0,3})$`)
	reNoVowels       = regexp.MustCompile(`(?i)^[^aeyiuo]+$`)
	rePeriodNotAtEnd = regexp.MustCompile(`(?i).*\..+$`)
	reEmoji          = regexp.MustCompile(`[\x{1F300}-\x{1F64F}\x{1F680}-\x{1F6FF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}]+`)
	rePhD            = regexp.MustCompile(`(?i)\s(ph\.?\s+d\.?)`)
)

type StringSet struct {
	values   map[string]struct{}
	onChange func()
}

func NewStringSet(items []string) *StringSet {
	s := &StringSet{values: map[string]struct{}{}}
	for _, item := range items {
		s.values[lc(item)] = struct{}{}
	}
	return s
}

func (s *StringSet) WithOnChange(fn func()) *StringSet {
	s.onChange = fn
	return s
}

func (s *StringSet) Clone() *StringSet {
	out := &StringSet{values: map[string]struct{}{}}
	for item := range s.values {
		out.values[item] = struct{}{}
	}
	return out
}

func (s *StringSet) Add(items ...string) *StringSet {
	for _, item := range items {
		n := lc(item)
		if n == "" {
			continue
		}
		s.values[n] = struct{}{}
	}
	if s.onChange != nil {
		s.onChange()
	}
	return s
}

func (s *StringSet) Remove(items ...string) *StringSet {
	for _, item := range items {
		delete(s.values, lc(item))
	}
	if s.onChange != nil {
		s.onChange()
	}
	return s
}

func (s *StringSet) AddWithEncoding(input []byte, encoding string) error {
	decoded, err := decodeBytes(input, encoding)
	if err != nil {
		return err
	}
	s.Add(decoded)
	return nil
}

func (s *StringSet) Contains(item string) bool {
	_, ok := s.values[lc(item)]
	return ok
}

func (s *StringSet) Len() int {
	return len(s.values)
}

func (s *StringSet) Values() []string {
	out := make([]string, 0, len(s.values))
	for item := range s.values {
		out = append(out, item)
	}
	return out
}

type Regexes struct {
	Spaces         *regexp.Regexp
	Word           *regexp.Regexp
	Mac            *regexp.Regexp
	Initial        *regexp.Regexp
	QuotedWord     *regexp.Regexp
	DoubleQuotes   *regexp.Regexp
	Parenthesis    *regexp.Regexp
	RomanNumeral   *regexp.Regexp
	NoVowels       *regexp.Regexp
	PeriodNotAtEnd *regexp.Regexp
	Emoji          *regexp.Regexp
	PhD            *regexp.Regexp
}

func defaultRegexes() Regexes {
	return Regexes{
		Spaces:         reSpaces,
		Word:           reWord,
		Mac:            reMac,
		Initial:        reInitial,
		QuotedWord:     reQuotedWord,
		DoubleQuotes:   reDoubleQuotes,
		Parenthesis:    reParenthesis,
		RomanNumeral:   reRomanNumeral,
		NoVowels:       reNoVowels,
		PeriodNotAtEnd: rePeriodNotAtEnd,
		Emoji:          reEmoji,
		PhD:            rePhD,
	}
}

func (r Regexes) Clone() Regexes {
	return r
}

type Constants struct {
	Prefixes                 *StringSet
	SuffixAcronyms           *StringSet
	SuffixNotAcronyms        *StringSet
	Titles                   *StringSet
	FirstNameTitles          *StringSet
	Conjunctions             *StringSet
	CapitalizationExceptions map[string]string
	Regexes                  Regexes

	StringFormat                 string
	InitialsFormat               string
	InitialsDelimiter            string
	EmptyAttributeDefault        string
	CapitalizeName               bool
	ForceMixedCaseCapitalization bool

	suffixesPrefixesTitles *StringSet
}

func NewConstants() *Constants {
	c := &Constants{
		Prefixes:                     NewStringSet(defaultPrefixesData),
		SuffixAcronyms:               NewStringSet(defaultSuffixAcronymsData),
		SuffixNotAcronyms:            NewStringSet(defaultSuffixNotAcronymsData),
		Titles:                       NewStringSet(defaultTitlesData),
		FirstNameTitles:              NewStringSet(defaultFirstNameTitlesData),
		Conjunctions:                 NewStringSet(defaultConjunctionsData),
		CapitalizationExceptions:     map[string]string{},
		Regexes:                      defaultRegexes(),
		StringFormat:                 "{title} {first} {middle} {last} {suffix} ({nickname})",
		InitialsFormat:               "{first} {middle} {last}",
		InitialsDelimiter:            ".",
		EmptyAttributeDefault:        "",
		CapitalizeName:               false,
		ForceMixedCaseCapitalization: false,
	}

	for k, v := range defaultCapitalizationExceptionsData {
		c.CapitalizationExceptions[k] = v
	}

	invalidate := func() { c.suffixesPrefixesTitles = nil }
	c.Prefixes.WithOnChange(invalidate)
	c.SuffixAcronyms.WithOnChange(invalidate)
	c.SuffixNotAcronyms.WithOnChange(invalidate)
	c.Titles.WithOnChange(invalidate)
	return c
}

func (c *Constants) Clone() *Constants {
	out := &Constants{
		Prefixes:                     c.Prefixes.Clone(),
		SuffixAcronyms:               c.SuffixAcronyms.Clone(),
		SuffixNotAcronyms:            c.SuffixNotAcronyms.Clone(),
		Titles:                       c.Titles.Clone(),
		FirstNameTitles:              c.FirstNameTitles.Clone(),
		Conjunctions:                 c.Conjunctions.Clone(),
		CapitalizationExceptions:     map[string]string{},
		Regexes:                      c.Regexes.Clone(),
		StringFormat:                 c.StringFormat,
		InitialsFormat:               c.InitialsFormat,
		InitialsDelimiter:            c.InitialsDelimiter,
		EmptyAttributeDefault:        c.EmptyAttributeDefault,
		CapitalizeName:               c.CapitalizeName,
		ForceMixedCaseCapitalization: c.ForceMixedCaseCapitalization,
	}
	for k, v := range c.CapitalizationExceptions {
		out.CapitalizationExceptions[k] = v
	}
	invalidate := func() { out.suffixesPrefixesTitles = nil }
	out.Prefixes.WithOnChange(invalidate)
	out.SuffixAcronyms.WithOnChange(invalidate)
	out.SuffixNotAcronyms.WithOnChange(invalidate)
	out.Titles.WithOnChange(invalidate)
	return out
}

func (c *Constants) SuffixesPrefixesTitles() *StringSet {
	if c.suffixesPrefixesTitles != nil {
		return c.suffixesPrefixesTitles
	}
	all := NewStringSet(nil)
	all.Add(c.Prefixes.Values()...)
	all.Add(c.SuffixAcronyms.Values()...)
	all.Add(c.SuffixNotAcronyms.Values()...)
	all.Add(c.Titles.Values()...)
	c.suffixesPrefixesTitles = all
	return c.suffixesPrefixesTitles
}

func (c *Constants) SetRegex(name string, re *regexp.Regexp) {
	switch strings.ToLower(name) {
	case "spaces":
		c.Regexes.Spaces = re
	case "word":
		c.Regexes.Word = re
	case "mac":
		c.Regexes.Mac = re
	case "initial":
		c.Regexes.Initial = re
	case "quoted_word":
		c.Regexes.QuotedWord = re
	case "double_quotes":
		c.Regexes.DoubleQuotes = re
	case "parenthesis":
		c.Regexes.Parenthesis = re
	case "roman_numeral":
		c.Regexes.RomanNumeral = re
	case "no_vowels":
		c.Regexes.NoVowels = re
	case "period_not_at_end":
		c.Regexes.PeriodNotAtEnd = re
	case "emoji":
		c.Regexes.Emoji = re
	case "phd":
		c.Regexes.PhD = re
	}
}

var CONSTANTS = NewConstants()
