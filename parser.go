package nameparser

import (
	"fmt"
	"strings"
)

var memberOrder = []string{"title", "first", "middle", "last", "suffix", "nickname"}

type HumanNameOptions struct {
	Constants         *Constants
	Encoding          string
	StringFormat      string
	InitialsFormat    string
	InitialsDelimiter string

	First    any
	Middle   any
	Last     any
	Title    any
	Suffix   any
	Nickname any
}

type HumanName struct {
	C                 *Constants
	Encoding          string
	StringFormat      string
	InitialsFormat    string
	InitialsDelimiter string

	Original     string
	fullName     string
	Unparsable   bool
	hasOwnConfig bool

	titleList    []string
	firstList    []string
	middleList   []string
	lastList     []string
	suffixList   []string
	nicknameList []string

	inferredTitles      map[string]struct{}
	inferredConjunction map[string]struct{}
	inferredSuffixes    map[string]struct{}
}

func New(fullName string) *HumanName {
	return NewWithConstants(fullName, CONSTANTS)
}

func NewWithConstants(fullName string, constants *Constants) *HumanName {
	h := newHumanNameBase(constants)
	h.SetFullName(fullName)
	return h
}

func NewWithOptions(fullName string, opts HumanNameOptions) (*HumanName, error) {
	h := newHumanNameBase(opts.Constants)
	if opts.Encoding != "" {
		h.Encoding = opts.Encoding
	}
	if opts.StringFormat != "" {
		h.StringFormat = opts.StringFormat
	}
	if opts.InitialsFormat != "" {
		h.InitialsFormat = opts.InitialsFormat
	}
	if opts.InitialsDelimiter != "" {
		h.InitialsDelimiter = opts.InitialsDelimiter
	}

	hasParts := opts.First != nil || opts.Middle != nil || opts.Last != nil || opts.Title != nil || opts.Suffix != nil || opts.Nickname != nil
	if hasParts {
		if err := h.SetFirst(opts.First); err != nil {
			return nil, err
		}
		if err := h.SetMiddle(opts.Middle); err != nil {
			return nil, err
		}
		if err := h.SetLast(opts.Last); err != nil {
			return nil, err
		}
		if err := h.SetTitle(opts.Title); err != nil {
			return nil, err
		}
		if err := h.SetSuffix(opts.Suffix); err != nil {
			return nil, err
		}
		if err := h.SetNickname(opts.Nickname); err != nil {
			return nil, err
		}
		h.Unparsable = false
		return h, nil
	}

	h.SetFullName(fullName)
	return h, nil
}

func newHumanNameBase(constants *Constants) *HumanName {
	c := constants
	hasOwn := false
	if c == nil {
		c = NewConstants()
		hasOwn = true
	}

	h := &HumanName{
		C:                 c,
		Encoding:          DefaultEncoding,
		StringFormat:      c.StringFormat,
		InitialsFormat:    c.InitialsFormat,
		InitialsDelimiter: c.InitialsDelimiter,
		Unparsable:        true,
		hasOwnConfig:      hasOwn,
	}
	return h
}

func (h *HumanName) HasOwnConfig() bool {
	return h.hasOwnConfig
}

func (h *HumanName) SetFullName(value string) {
	h.Original = value
	h.fullName = value
	h.parseFullName()
}

func (h *HumanName) SetFullNameBytes(value []byte) error {
	decoded, err := decodeBytes(value, h.Encoding)
	if err != nil {
		return err
	}
	h.SetFullName(decoded)
	return nil
}

func (h *HumanName) FullName() string {
	return h.String()
}

func (h *HumanName) String() string {
	if h.StringFormat != "" {
		d := h.AsDict(true)
		s := h.StringFormat
		for _, key := range memberOrder {
			s = strings.ReplaceAll(s, "{"+key+"}", d[key])
		}
		s = sanitizeFormatOutput(s, h.C.EmptyAttributeDefault)
		return strings.Trim(h.collapseWhitespace(s), ", ")
	}
	return strings.Join(h.Members(), " ")
}

func (h *HumanName) AsDict(includeEmpty bool) map[string]string {
	d := map[string]string{}
	for _, key := range memberOrder {
		val, _ := h.Member(key)
		if includeEmpty || val != "" {
			d[key] = val
		}
	}
	return d
}

func (h *HumanName) Members() []string {
	out := make([]string, 0, len(memberOrder))
	for _, key := range memberOrder {
		val, _ := h.Member(key)
		if val != "" {
			out = append(out, val)
		}
	}
	return out
}

func normalizeIndex(i, n int) int {
	if i < 0 {
		i += n
	}
	if i < 0 {
		i = 0
	}
	if i > n {
		i = n
	}
	return i
}

func (h *HumanName) MembersSlice(start int, end *int) []string {
	members := []string{h.Title(), h.First(), h.Middle(), h.Last(), h.Suffix(), h.Nickname()}
	n := len(members)
	s := normalizeIndex(start, n)
	e := n
	if end != nil {
		e = normalizeIndex(*end, n)
	}
	if e < s {
		return []string{}
	}
	out := make([]string, 0, e-s)
	for _, item := range members[s:e] {
		if item == "" {
			out = append(out, h.C.EmptyAttributeDefault)
		} else {
			out = append(out, item)
		}
	}
	return out
}

func (h *HumanName) Len() int {
	hasValue := func(items []string) bool {
		for _, item := range items {
			if item != "" {
				return true
			}
		}
		return false
	}

	count := 0
	if hasValue(h.titleList) {
		count++
	}
	if hasValue(h.firstList) {
		count++
	}
	if hasValue(h.middleList) {
		count++
	}
	if hasValue(h.lastList) {
		count++
	}
	if hasValue(h.suffixList) {
		count++
	}
	if hasValue(h.nicknameList) {
		count++
	}
	return count
}

func (h *HumanName) EqualString(other string) bool {
	return strings.EqualFold(h.String(), other)
}

func (h *HumanName) EqualName(other *HumanName) bool {
	if other == nil {
		return false
	}
	return strings.EqualFold(h.String(), other.String())
}

func (h *HumanName) Title() string {
	val := joinSpace(h.titleList)
	if val == "" {
		return h.C.EmptyAttributeDefault
	}
	return val
}

func (h *HumanName) First() string {
	val := joinSpace(h.firstList)
	if val == "" {
		return h.C.EmptyAttributeDefault
	}
	return val
}

func (h *HumanName) Middle() string {
	val := joinSpace(h.middleList)
	if val == "" {
		return h.C.EmptyAttributeDefault
	}
	return val
}

func (h *HumanName) Last() string {
	val := joinSpace(h.lastList)
	if val == "" {
		return h.C.EmptyAttributeDefault
	}
	return val
}

func (h *HumanName) Suffix() string {
	val := strings.Join(h.suffixList, ", ")
	if val == "" {
		return h.C.EmptyAttributeDefault
	}
	return val
}

func (h *HumanName) Nickname() string {
	val := joinSpace(h.nicknameList)
	if val == "" {
		return h.C.EmptyAttributeDefault
	}
	return val
}

func (h *HumanName) SurnamesList() []string {
	out := make([]string, 0, len(h.middleList)+len(h.lastList))
	out = append(out, h.middleList...)
	out = append(out, h.lastList...)
	return out
}

func (h *HumanName) Surnames() string {
	val := joinSpace(h.SurnamesList())
	if val == "" {
		return h.C.EmptyAttributeDefault
	}
	return val
}

func (h *HumanName) TitleList() []string    { return copyStrings(h.titleList) }
func (h *HumanName) FirstList() []string    { return copyStrings(h.firstList) }
func (h *HumanName) MiddleList() []string   { return copyStrings(h.middleList) }
func (h *HumanName) LastList() []string     { return copyStrings(h.lastList) }
func (h *HumanName) SuffixList() []string   { return copyStrings(h.suffixList) }
func (h *HumanName) NicknameList() []string { return copyStrings(h.nicknameList) }

func (h *HumanName) Member(key string) (string, error) {
	switch key {
	case "title":
		return h.Title(), nil
	case "first":
		return h.First(), nil
	case "middle":
		return h.Middle(), nil
	case "last":
		return h.Last(), nil
	case "suffix":
		return h.Suffix(), nil
	case "nickname":
		return h.Nickname(), nil
	default:
		return "", fmt.Errorf("not a valid HumanName attribute: %s", key)
	}
}

func (h *HumanName) SetMember(key string, value any) error {
	switch key {
	case "title":
		return h.SetTitle(value)
	case "first":
		return h.SetFirst(value)
	case "middle":
		return h.SetMiddle(value)
	case "last":
		return h.SetLast(value)
	case "suffix":
		return h.SetSuffix(value)
	case "nickname":
		return h.SetNickname(value)
	default:
		return fmt.Errorf("not a valid HumanName attribute: %s", key)
	}
}

func (h *HumanName) setList(attr string, value any) error {
	var val []string
	switch v := value.(type) {
	case nil:
		val = []string{}
	case string:
		val = []string{v}
	case []string:
		val = v
	default:
		return fmt.Errorf("can only assign strings, lists or nil to name attributes: got %T", value)
	}

	parsed, err := h.parsePieces(val, 0)
	if err != nil {
		return err
	}

	switch attr {
	case "title":
		h.titleList = parsed
	case "first":
		h.firstList = parsed
	case "middle":
		h.middleList = parsed
	case "last":
		h.lastList = parsed
	case "suffix":
		h.suffixList = parsed
	case "nickname":
		h.nicknameList = parsed
	default:
		return fmt.Errorf("unknown attribute: %s", attr)
	}
	return nil
}

func (h *HumanName) SetTitle(value any) error    { return h.setList("title", value) }
func (h *HumanName) SetFirst(value any) error    { return h.setList("first", value) }
func (h *HumanName) SetMiddle(value any) error   { return h.setList("middle", value) }
func (h *HumanName) SetLast(value any) error     { return h.setList("last", value) }
func (h *HumanName) SetSuffix(value any) error   { return h.setList("suffix", value) }
func (h *HumanName) SetNickname(value any) error { return h.setList("nickname", value) }

func hasInferred(set map[string]struct{}, value string) bool {
	if len(set) == 0 {
		return false
	}
	_, ok := set[lc(value)]
	return ok
}

func addInferred(set *map[string]struct{}, value string) {
	normalized := lc(value)
	if normalized == "" {
		return
	}
	if *set == nil {
		*set = map[string]struct{}{}
	}
	(*set)[normalized] = struct{}{}
}

func (h *HumanName) IsTitle(value string) bool {
	return hasInferred(h.inferredTitles, value) || h.C.Titles.Contains(value)
}

func (h *HumanName) IsConjunction(piece any) bool {
	switch v := piece.(type) {
	case []string:
		for _, item := range v {
			if h.IsConjunction(item) {
				return true
			}
		}
		return false
	case string:
		return (hasInferred(h.inferredConjunction, v) || h.C.Conjunctions.Contains(v)) && !h.IsAnInitial(v)
	default:
		return false
	}
}

func (h *HumanName) IsPrefix(piece any) bool {
	switch v := piece.(type) {
	case []string:
		for _, item := range v {
			if h.IsPrefix(item) {
				return true
			}
		}
		return false
	case string:
		return h.C.Prefixes.Contains(v)
	default:
		return false
	}
}

func (h *HumanName) IsRomanNumeral(value string) bool {
	if h.C.Regexes.RomanNumeral == nil {
		return false
	}
	return h.C.Regexes.RomanNumeral.MatchString(value)
}

func (h *HumanName) IsSuffix(piece any) bool {
	switch v := piece.(type) {
	case []string:
		for _, item := range v {
			if h.IsSuffix(item) {
				return true
			}
		}
		return false
	case string:
		normalized := lc(v)
		isAcronym := h.C.SuffixAcronyms.Contains(strings.ReplaceAll(normalized, ".", ""))
		isNotAcronym := hasInferred(h.inferredSuffixes, normalized) || h.C.SuffixNotAcronyms.Contains(normalized)
		return (isAcronym || isNotAcronym) && !h.IsAnInitial(v)
	default:
		return false
	}
}

func (h *HumanName) AreSuffixes(pieces []string) bool {
	for _, piece := range pieces {
		if !h.IsSuffix(piece) {
			return false
		}
	}
	return true
}

func (h *HumanName) IsRootname(piece string) bool {
	return !h.C.SuffixesPrefixesTitles().Contains(piece) &&
		!hasInferred(h.inferredTitles, piece) &&
		!hasInferred(h.inferredSuffixes, piece) &&
		!h.IsAnInitial(piece)
}

func (h *HumanName) IsAnInitial(value string) bool {
	if h.C.Regexes.Initial == nil {
		return false
	}
	return h.C.Regexes.Initial.MatchString(value)
}

func (h *HumanName) collapseWhitespace(value string) string {
	if h.C.Regexes.Spaces != nil {
		value = h.C.Regexes.Spaces.ReplaceAllString(strings.TrimSpace(value), " ")
	} else {
		value = strings.TrimSpace(value)
	}
	value = strings.TrimSuffix(value, ",")
	return value
}

func (h *HumanName) preProcess() {
	h.fixPhD()
	h.parseNicknames()
	h.squashEmoji()
}

func (h *HumanName) postProcess() {
	h.handleFirstnames()
	h.handleCapitalization()
}

func (h *HumanName) fixPhD() {
	re := h.C.Regexes.PhD
	if re == nil {
		return
	}
	match := re.FindStringSubmatch(h.fullName)
	if len(match) > 1 {
		h.suffixList = append(h.suffixList, match[1])
		h.fullName = re.ReplaceAllString(h.fullName, "")
	}
}

func (h *HumanName) parseNicknames() {
	reQuoted := h.C.Regexes.QuotedWord
	if reQuoted != nil && reQuoted.MatchString(h.fullName) {
		matches := reQuoted.FindAllStringSubmatch(h.fullName, -1)
		for _, m := range matches {
			if len(m) > 2 {
				h.nicknameList = append(h.nicknameList, m[2])
			}
		}
		h.fullName = reQuoted.ReplaceAllString(h.fullName, "$1$3")
	}

	reDouble := h.C.Regexes.DoubleQuotes
	if reDouble != nil && reDouble.MatchString(h.fullName) {
		matches := reDouble.FindAllStringSubmatch(h.fullName, -1)
		for _, m := range matches {
			if len(m) > 1 {
				h.nicknameList = append(h.nicknameList, m[1])
			}
		}
		h.fullName = reDouble.ReplaceAllString(h.fullName, "")
	}

	reParen := h.C.Regexes.Parenthesis
	if reParen != nil && reParen.MatchString(h.fullName) {
		matches := reParen.FindAllStringSubmatch(h.fullName, -1)
		for _, m := range matches {
			if len(m) > 1 {
				h.nicknameList = append(h.nicknameList, m[1])
			}
		}
		h.fullName = reParen.ReplaceAllString(h.fullName, "")
	}
}

func (h *HumanName) squashEmoji() {
	re := h.C.Regexes.Emoji
	if re != nil && re.MatchString(h.fullName) {
		h.fullName = re.ReplaceAllString(h.fullName, "")
	}
}

func (h *HumanName) handleFirstnames() {
	title := joinSpace(h.titleList)
	if title != "" && len(h.firstList) > 0 && h.Len() == 2 && !h.C.FirstNameTitles.Contains(title) {
		h.firstList, h.lastList = h.lastList, h.firstList
	}
}

type integerRange struct {
	start int
	end   int
}

func groupContiguousIntegers(data []int) []integerRange {
	if len(data) == 0 {
		return nil
	}
	ranges := []integerRange{}
	start := data[0]
	prev := data[0]
	for i := 1; i < len(data); i++ {
		if data[i] == prev+1 {
			prev = data[i]
			continue
		}
		if prev > start {
			ranges = append(ranges, integerRange{start: start, end: prev})
		}
		start = data[i]
		prev = data[i]
	}
	if prev > start {
		ranges = append(ranges, integerRange{start: start, end: prev})
	}
	return ranges
}

func (h *HumanName) parseFullName() {
	h.titleList = []string{}
	h.firstList = []string{}
	h.middleList = []string{}
	h.lastList = []string{}
	h.suffixList = []string{}
	h.nicknameList = []string{}
	h.inferredTitles = nil
	h.inferredConjunction = nil
	h.inferredSuffixes = nil
	h.Unparsable = true

	h.preProcess()
	h.fullName = h.collapseWhitespace(h.fullName)

	parts := strings.Split(h.fullName, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	if len(parts) == 1 {
		pieces, _ := h.parsePieces(parts, 0)
		pLen := len(pieces)
		for i, piece := range pieces {
			nxt := ""
			hasNext := i+1 < pLen
			if hasNext {
				nxt = pieces[i+1]
			}

			if len(h.firstList) == 0 && (hasNext || pLen == 1) && h.IsTitle(piece) {
				h.titleList = append(h.titleList, piece)
				continue
			}
			if len(h.firstList) == 0 {
				if pLen == 1 && len(h.nicknameList) > 0 {
					h.lastList = append(h.lastList, piece)
					continue
				}
				h.firstList = append(h.firstList, piece)
				continue
			}

			if h.AreSuffixes(pieces[i+1:]) ||
				(hasNext && h.IsRomanNumeral(nxt) && i == pLen-2 && !h.IsAnInitial(piece)) {
				h.lastList = append(h.lastList, piece)
				h.suffixList = append(h.suffixList, pieces[i+1:]...)
				break
			}

			if !hasNext {
				h.lastList = append(h.lastList, piece)
				continue
			}
			h.middleList = append(h.middleList, piece)
		}
	} else {
		postCommaSplit := strings.Split(parts[1], " ")
		firstPartSplit := strings.Split(parts[0], " ")

		postCommaPieces, _ := h.parsePieces(postCommaSplit, 1)
		if h.AreSuffixes(postCommaSplit) && len(firstPartSplit) > 1 {
			h.suffixList = append(h.suffixList, parts[1:]...)
			pieces, _ := h.parsePieces(firstPartSplit, 0)
			for i, piece := range pieces {
				hasNext := i+1 < len(pieces)
				if len(h.firstList) == 0 && (hasNext || len(pieces) == 1) && h.IsTitle(piece) {
					h.titleList = append(h.titleList, piece)
					continue
				}
				if len(h.firstList) == 0 {
					h.firstList = append(h.firstList, piece)
					continue
				}
				if h.AreSuffixes(pieces[i+1:]) {
					h.lastList = append(h.lastList, piece)
					h.suffixList = append(append([]string{}, pieces[i+1:]...), h.suffixList...)
					break
				}
				if !hasNext {
					h.lastList = append(h.lastList, piece)
					continue
				}
				h.middleList = append(h.middleList, piece)
			}
		} else {
			lastnamePieces, _ := h.parsePieces(firstPartSplit, 1)
			for _, piece := range lastnamePieces {
				if h.IsSuffix(piece) && len(h.lastList) > 0 {
					h.suffixList = append(h.suffixList, piece)
				} else {
					h.lastList = append(h.lastList, piece)
				}
			}

			for i, piece := range postCommaPieces {
				hasNext := i+1 < len(postCommaPieces)
				if len(h.firstList) == 0 && (hasNext || len(postCommaPieces) == 1) && h.IsTitle(piece) {
					h.titleList = append(h.titleList, piece)
					continue
				}
				if len(h.firstList) == 0 {
					h.firstList = append(h.firstList, piece)
					continue
				}
				if h.IsSuffix(piece) {
					h.suffixList = append(h.suffixList, piece)
					continue
				}
				h.middleList = append(h.middleList, piece)
			}
			if len(parts) > 2 && parts[2] != "" {
				h.suffixList = append(h.suffixList, parts[2:]...)
			}
		}
	}

	h.Unparsable = h.Len() == 0
	h.postProcess()
}

func (h *HumanName) parsePieces(parts []string, additionalPartsCount int) ([]string, error) {
	output := []string{}
	for _, part := range parts {
		chunks := strings.Split(part, " ")
		for _, chunk := range chunks {
			chunk = strings.Trim(chunk, " ,")
			if chunk == "" {
				continue
			}
			output = append(output, chunk)
		}
	}

	for _, part := range output {
		if h.C.Regexes.PeriodNotAtEnd != nil && h.C.Regexes.PeriodNotAtEnd.MatchString(part) {
			periodChunks := strings.Split(part, ".")
			titles := []string{}
			suffixes := []string{}
			for _, chunk := range periodChunks {
				if h.IsTitle(chunk) {
					titles = append(titles, chunk)
				}
				if h.IsSuffix(chunk) {
					suffixes = append(suffixes, chunk)
				}
			}
			if len(titles) > 0 {
				addInferred(&h.inferredTitles, part)
				continue
			}
			if len(suffixes) > 0 {
				addInferred(&h.inferredSuffixes, part)
				continue
			}
		}
	}

	return h.joinOnConjunctions(output, additionalPartsCount), nil
}

func (h *HumanName) joinOnConjunctions(pieces []string, additionalPartsCount int) []string {
	length := len(pieces) + additionalPartsCount
	if length < 3 {
		return pieces
	}

	rootnamePieces := []string{}
	for _, p := range pieces {
		if h.IsRootname(p) {
			rootnamePieces = append(rootnamePieces, p)
		}
	}
	totalLength := len(rootnamePieces) + additionalPartsCount

	conjIndex := []int{}
	for i, piece := range pieces {
		if h.IsConjunction(piece) {
			conjIndex = append(conjIndex, i)
		}
	}

	contiguous := groupContiguousIntegers(conjIndex)
	deleteIndex := []int{}
	for _, r := range contiguous {
		newPiece := strings.Join(pieces[r.start:r.end+1], " ")
		for i := r.start + 1; i <= r.end; i++ {
			deleteIndex = append(deleteIndex, i)
		}
		pieces[r.start] = newPiece
		addInferred(&h.inferredConjunction, newPiece)
	}

	for i := len(deleteIndex) - 1; i >= 0; i-- {
		idx := deleteIndex[i]
		if idx >= 0 && idx < len(pieces) {
			pieces = append(pieces[:idx], pieces[idx+1:]...)
		}
	}

	if len(pieces) == 1 {
		return pieces
	}

	conjIndex = []int{}
	for i, piece := range pieces {
		if h.IsConjunction(piece) {
			conjIndex = append(conjIndex, i)
		}
	}

	for iPos := 0; iPos < len(conjIndex); iPos++ {
		i := conjIndex[iPos]
		if i < 0 || i >= len(pieces) {
			continue
		}
		if len(pieces[i]) == 1 && totalLength < 4 {
			continue
		}

		if i == 0 {
			if len(pieces) < 2 {
				continue
			}
			newPiece := strings.Join(pieces[i:i+2], " ")
			if h.IsTitle(pieces[i+1]) {
				addInferred(&h.inferredTitles, newPiece)
			}
			pieces[i] = newPiece
			pieces = append(pieces[:i+1], pieces[i+2:]...)
			for j, v := range conjIndex {
				if v > i {
					conjIndex[j] = v - 1
				}
			}
		} else {
			left := i - 1
			right := i + 2
			if right > len(pieces) {
				right = len(pieces)
			}
			newPiece := strings.Join(pieces[left:right], " ")
			if h.IsTitle(pieces[left]) {
				addInferred(&h.inferredTitles, newPiece)
			}
			pieces[left] = newPiece
			if i < len(pieces) {
				pieces = append(pieces[:i], pieces[i+1:]...)
			}
			rmCount := 2
			if i < len(pieces) {
				pieces = append(pieces[:i], pieces[i+1:]...)
			} else {
				rmCount = 1
			}
			for j, v := range conjIndex {
				if v > i {
					conjIndex[j] = v - rmCount
				}
			}
		}
	}

	prefixes := []string{}
	for _, piece := range pieces {
		if h.IsPrefix(piece) {
			prefixes = append(prefixes, piece)
		}
	}
	if len(prefixes) == 0 {
		return pieces
	}

	i := -1
	for _, prefix := range prefixes {
		idx := indexOf(pieces, prefix, 0)
		if idx >= 0 {
			i = idx
		}
		if i < 0 || i >= len(pieces) {
			continue
		}
		if i == 0 && totalLength >= 1 {
			continue
		}

		nextPrefix, hasNextPrefix := firstMatch(pieces[i+1:], func(s string) bool { return h.IsPrefix(s) })
		if hasNextPrefix {
			j := indexOf(pieces, nextPrefix, i+1)
			if j == i+1 {
				j++
			}
			if j > len(pieces) {
				j = len(pieces)
			}
			newPiece := strings.Join(pieces[i:j], " ")
			pieces = append(append(copyStrings(pieces[:i]), newPiece), pieces[j:]...)
			continue
		}

		stopAt, hasSuffix := firstMatch(pieces[i+1:], func(s string) bool { return h.IsSuffix(s) })
		if hasSuffix {
			j := indexOf(pieces, stopAt, i+1)
			newPiece := strings.Join(pieces[i:j], " ")
			pieces = append(append(copyStrings(pieces[:i]), newPiece), pieces[j:]...)
			continue
		}

		newPiece := strings.Join(pieces[i:], " ")
		pieces = append(copyStrings(pieces[:i]), newPiece)
	}

	return pieces
}

func (h *HumanName) processInitial(namePart string, firstName bool) string {
	parts := strings.Split(namePart, " ")
	initials := []string{}
	for _, part := range parts {
		if part == "" {
			continue
		}
		if (!h.IsPrefix(part) && !h.IsConjunction(part)) || firstName {
			initials = append(initials, string([]rune(part)[0]))
		}
	}
	if len(initials) > 0 {
		return strings.Join(initials, " ")
	}
	return h.C.EmptyAttributeDefault
}

func (h *HumanName) InitialsList() []string {
	first := []string{}
	for _, name := range h.firstList {
		if name != "" {
			first = append(first, h.processInitial(name, true))
		}
	}
	middle := []string{}
	for _, name := range h.middleList {
		if name != "" {
			middle = append(middle, h.processInitial(name, false))
		}
	}
	last := []string{}
	for _, name := range h.lastList {
		if name != "" {
			last = append(last, h.processInitial(name, false))
		}
	}
	out := append(first, middle...)
	out = append(out, last...)
	return out
}

func (h *HumanName) Initials() string {
	firstInitials := []string{}
	for _, name := range h.firstList {
		if name != "" {
			firstInitials = append(firstInitials, h.processInitial(name, true))
		}
	}
	middleInitials := []string{}
	for _, name := range h.middleList {
		if name != "" {
			middleInitials = append(middleInitials, h.processInitial(name, false))
		}
	}
	lastInitials := []string{}
	for _, name := range h.lastList {
		if name != "" {
			lastInitials = append(lastInitials, h.processInitial(name, false))
		}
	}

	joiner := h.InitialsDelimiter + " "
	parts := map[string]string{
		"first":  h.C.EmptyAttributeDefault,
		"middle": h.C.EmptyAttributeDefault,
		"last":   h.C.EmptyAttributeDefault,
	}
	if len(firstInitials) > 0 {
		parts["first"] = strings.Join(firstInitials, joiner) + h.InitialsDelimiter
	}
	if len(middleInitials) > 0 {
		parts["middle"] = strings.Join(middleInitials, joiner) + h.InitialsDelimiter
	}
	if len(lastInitials) > 0 {
		parts["last"] = strings.Join(lastInitials, joiner) + h.InitialsDelimiter
	}

	s := h.InitialsFormat
	s = strings.ReplaceAll(s, "{first}", parts["first"])
	s = strings.ReplaceAll(s, "{middle}", parts["middle"])
	s = strings.ReplaceAll(s, "{last}", parts["last"])
	return h.collapseWhitespace(s)
}

func (h *HumanName) capWord(word, attribute string) string {
	if (h.IsPrefix(word) && (attribute == "last" || attribute == "middle")) || h.IsConjunction(word) {
		return strings.ToLower(word)
	}
	if exception, ok := h.C.CapitalizationExceptions[lc(word)]; ok {
		return exception
	}
	if h.C.Regexes.Mac != nil {
		match := h.C.Regexes.Mac.FindStringSubmatch(word)
		if len(match) == 3 {
			return titleCaseWord(match[1]) + titleCaseWord(match[2])
		}
	}
	return titleCaseWord(word)
}

func (h *HumanName) capPiece(piece, attribute string) string {
	if piece == "" {
		return ""
	}
	re := h.C.Regexes.Word
	if re == nil {
		return piece
	}
	return re.ReplaceAllStringFunc(piece, func(m string) string {
		return h.capWord(m, attribute)
	})
}

func (h *HumanName) Capitalize(force *bool) {
	name := h.String()
	applyForce := h.C.ForceMixedCaseCapitalization
	if force != nil {
		applyForce = *force
	}
	if !applyForce {
		if name != strings.ToUpper(name) && name != strings.ToLower(name) {
			return
		}
	}

	h.titleList = strings.Split(h.capPiece(h.Title(), "title"), " ")
	h.firstList = strings.Split(h.capPiece(h.First(), "first"), " ")
	h.middleList = strings.Split(h.capPiece(h.Middle(), "middle"), " ")
	h.lastList = strings.Split(h.capPiece(h.Last(), "last"), " ")
	h.suffixList = strings.Split(h.capPiece(h.Suffix(), "suffix"), ", ")
}

func (h *HumanName) handleCapitalization() {
	if h.C.CapitalizeName {
		h.Capitalize(nil)
	}
}
