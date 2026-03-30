package nameparser

import (
	"regexp"
	"strings"
	"sync"
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
	mu       sync.RWMutex
	values   map[string]struct{}
	onChange func()
}

func NewStringSet(items []string) *StringSet {
	s := &StringSet{values: make(map[string]struct{}, len(items))}
	for _, item := range items {
		s.values[lc(item)] = struct{}{}
	}
	return s
}

func (s *StringSet) WithOnChange(fn func()) *StringSet {
	s.mu.Lock()
	s.onChange = fn
	s.mu.Unlock()
	return s
}

func (s *StringSet) Clone() *StringSet {
	s.mu.RLock()
	out := &StringSet{values: make(map[string]struct{}, len(s.values))}
	for item := range s.values {
		out.values[item] = struct{}{}
	}
	s.mu.RUnlock()
	return out
}

func (s *StringSet) Add(items ...string) *StringSet {
	changed := false
	s.mu.Lock()
	for _, item := range items {
		n := lc(item)
		if n == "" {
			continue
		}
		if _, ok := s.values[n]; !ok {
			changed = true
		}
		s.values[n] = struct{}{}
	}
	onChange := s.onChange
	s.mu.Unlock()
	if changed && onChange != nil {
		onChange()
	}
	return s
}

func (s *StringSet) Remove(items ...string) *StringSet {
	changed := false
	s.mu.Lock()
	for _, item := range items {
		n := lc(item)
		if _, ok := s.values[n]; ok {
			delete(s.values, n)
			changed = true
		}
	}
	onChange := s.onChange
	s.mu.Unlock()
	if changed && onChange != nil {
		onChange()
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
	s.mu.RLock()
	_, ok := s.values[lc(item)]
	s.mu.RUnlock()
	return ok
}

func (s *StringSet) Len() int {
	s.mu.RLock()
	n := len(s.values)
	s.mu.RUnlock()
	return n
}

func (s *StringSet) Values() []string {
	s.mu.RLock()
	out := make([]string, 0, len(s.values))
	for item := range s.values {
		out = append(out, item)
	}
	s.mu.RUnlock()
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

	mu                     sync.RWMutex
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

	invalidate := c.invalidateSuffixesPrefixesTitles
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
	invalidate := out.invalidateSuffixesPrefixesTitles
	out.Prefixes.WithOnChange(invalidate)
	out.SuffixAcronyms.WithOnChange(invalidate)
	out.SuffixNotAcronyms.WithOnChange(invalidate)
	out.Titles.WithOnChange(invalidate)
	return out
}

func (c *Constants) invalidateSuffixesPrefixesTitles() {
	c.mu.Lock()
	c.suffixesPrefixesTitles = nil
	c.mu.Unlock()
}

func (c *Constants) SuffixesPrefixesTitles() *StringSet {
	c.mu.RLock()
	cached := c.suffixesPrefixesTitles
	c.mu.RUnlock()
	if cached != nil {
		return cached
	}

	all := NewStringSet(nil)
	all.Add(c.Prefixes.Values()...)
	all.Add(c.SuffixAcronyms.Values()...)
	all.Add(c.SuffixNotAcronyms.Values()...)
	all.Add(c.Titles.Values()...)

	c.mu.Lock()
	if c.suffixesPrefixesTitles == nil {
		c.suffixesPrefixesTitles = all
	}
	cached = c.suffixesPrefixesTitles
	c.mu.Unlock()
	return cached
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
