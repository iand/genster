package text

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// LowerFirst lowercases the first rune of s and returns the result.
// It panics if s is empty.
func LowerFirst(s string) string {
	r := []rune(s)

	return strings.ToLower(string(r[0])) + string(r[1:])
}

// LowerIfFirstWordIn lowercases the first rune of s if s begins with one of
// the given words followed by a space. The match is an exact word check - a
// bare prefix is not enough. If no word matches, s is returned unchanged.
func LowerIfFirstWordIn(s string, words ...string) string {
	for _, w := range words {
		if strings.HasPrefix(s, w+" ") {
			return LowerFirst(s)
		}
	}
	return s
}

// UpperFirst trims surrounding whitespace from s, then uppercases its first
// rune. Returns an empty string if s is blank after trimming.
func UpperFirst(s string) string {
	s = strings.TrimFunc(s, unicode.IsSpace)
	if len(s) == 0 {
		return ""
	} else if len(s) == 1 {
		return strings.ToUpper(s)
	}

	r := []rune(s)
	return strings.ToUpper(string(r[0])) + string(r[1:])
}

// RemoveRedundantWhitespace collapses each run of whitespace to a single space
// and trims leading and trailing whitespace.
func RemoveRedundantWhitespace(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

// RemoveAllWhitespace removes every whitespace character from s.
func RemoveAllWhitespace(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), "")
}

// CardinalNoun returns the English word for n. Values 0-199 are returned as
// words (e.g. "forty-two"); values above 199 fall back to a decimal digit
// string.
func CardinalNoun(n int) string {
	noun := cardinalNounUnderTwenty(n)
	if noun != "" {
		return noun
	}
	if n > 199 {
		return fmt.Sprintf("%d", n)
	}

	if n == 100 {
		return "one hundred"
	}
	if n > 99 {
		noun = "one hundred and "
		n -= 100
	}
	if n < 20 {
		return noun + cardinalNounUnderTwenty(n)
	}

	switch n / 10 {
	case 2:
		noun += "twenty"
	case 3:
		noun += "thirty"
	case 4:
		noun += "forty"
	case 5:
		noun += "fifty"
	case 6:
		noun += "sixty"
	case 7:
		noun += "seventy"
	case 8:
		noun += "eighty"
	case 9:
		noun += "ninety"
	}

	switch n % 10 {
	case 1:
		noun += "-one"
	case 2:
		noun += "-two"
	case 3:
		noun += "-three"
	case 4:
		noun += "-four"
	case 5:
		noun += "-five"
	case 6:
		noun += "-six"
	case 7:
		noun += "-seven"
	case 8:
		noun += "-eight"
	case 9:
		noun += "-nine"

	}

	return noun
}

func cardinalNounUnderTwenty(n int) string {
	switch n {
	case 0:
		return "no"
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	case 4:
		return "four"
	case 5:
		return "five"
	case 6:
		return "six"
	case 7:
		return "seven"
	case 8:
		return "eight"
	case 9:
		return "nine"
	case 10:
		return "ten"
	case 11:
		return "eleven"
	case 12:
		return "twelve"
	case 13:
		return "thirteen"
	case 14:
		return "fourteen"
	case 15:
		return "fifteen"
	case 16:
		return "sixteen"
	case 17:
		return "seventeen"
	case 18:
		return "eighteen"
	case 19:
		return "nineteen"
	}
	return ""
}

// SmallCardinalNoun returns the English word for n when n is 0-5 ("no",
// "one", ..., "five"), and a decimal digit string for larger values. Use this
// instead of CardinalNoun when the number is expected to be small but a
// fallback to digits is acceptable for prose readability.
func SmallCardinalNoun(n int) string {
	switch n {
	case 0:
		return "no"
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	case 4:
		return "four"
	case 5:
		return "five"
	default:
		return strconv.Itoa(n)
	}
}

// OrdinalNoun returns the English ordinal word for n (e.g. 1 -> "first",
// 21 -> "twenty-first"). Values above 100 are not handled and produce
// unexpected results.
func OrdinalNoun(n int) string {
	if n < 10 {
		return ordinalNounUnderTen(n)
	}
	switch n {
	case 10:
		return "tenth"
	case 11:
		return "eleventh"
	case 12:
		return "twelfth"
	case 13:
		return "thirteenth"
	case 14:
		return "fourteenth"
	case 15:
		return "fifteenth"
	case 16:
		return "sixteenth"
	case 17:
		return "seventeenth"
	case 18:
		return "eighteenth"
	case 19:
		return "nineteenth"
	case 20:
		return "twentieth"
	case 30:
		return "thirtieth"
	case 40:
		return "fortieth"
	case 50:
		return "fiftieth"
	case 60:
		return "sixtieth"
	case 70:
		return "seventieth"
	case 80:
		return "eightieth"
	case 90:
		return "ninetieth"
	case 100:
		return "one hundredth"
	}

	return CardinalNoun((n/10)*10) + "-" + ordinalNounUnderTen(n%10)
}

func ordinalNounUnderTen(n int) string {
	switch n {
	case 0:
		return "zeroeth"
	case 1:
		return "first"
	case 2:
		return "second"
	case 3:
		return "third"
	case 4:
		return "fourth"
	case 5:
		return "fifth"
	case 6:
		return "sixth"
	case 7:
		return "seventh"
	case 8:
		return "eighth"
	case 9:
		return "ninth"
	default:
		return ""
	}
}

// MultiplicativeAdverb returns the English adverb for n occurrences: "no"
// for 0, "once" for 1, "twice" for 2, and "N times" for larger values.
func MultiplicativeAdverb(n int) string {
	switch n {
	case 0:
		return "no"
	case 1:
		return "once"
	case 2:
		return "twice"
	default:
		return fmt.Sprintf("%s times", CardinalNoun(n))
	}
}

// JoinList joins elements with commas and a final " and ", producing natural
// English list prose (e.g. "one, two and three"). Leading/trailing
// whitespace and punctuation are stripped from each element.
func JoinList(strs []string) string {
	var ret strings.Builder
	for i, s := range strs {
		s = strings.Trim(s, " ,!.?")

		if i != 0 {
			if i == len(strs)-1 {
				ret.WriteString(" and ")
			} else {
				ret.WriteString(", ")
			}
		}
		ret.WriteString(s)
	}
	return ret.String()
}

// JoinListOr is like JoinList but uses " or " before the final element.
func JoinListOr(strs []string) string {
	var ret strings.Builder
	for i, s := range strs {
		s = strings.Trim(s, " ,!.?")

		if i != 0 {
			if i == len(strs)-1 {
				ret.WriteString(" or ")
			} else {
				ret.WriteString(", ")
			}
		}
		ret.WriteString(s)
	}
	return ret.String()
}

// JoinSentenceParts concatenates non-empty parts with a single space between
// them. Empty and whitespace-only parts are skipped. The colon ":" is
// special-cased to receive no leading space.
func JoinSentenceParts(parts ...string) string {
	var p Para
	p.Continue(parts...)
	return p.Current()
}

// JoinSentences formats each argument as a complete sentence and joins them
// with a single space. Empty arguments are skipped.
func JoinSentences(ss ...string) string {
	var p Para
	for _, s := range ss {
		p.AddCompleteSentence(s)
	}
	return p.Text()
}

// CardinalWithUnit returns "one <singular>" when n is 1, and
// "<CardinalNoun(n)> <plural>" otherwise (e.g. "one child", "two children").
func CardinalWithUnit(n int, singular string, plural string) string {
	if n == 1 {
		return "one " + singular
	}
	return CardinalNoun(n) + " " + plural
}

// CardinalSuffix returns the English ordinal suffix for n: "st", "nd", "rd",
// or "th". It correctly handles the 11/12/13 exceptions (e.g. 11 -> "th",
// not "st").
func CardinalSuffix(n int) string {
	// 11th, 12th, 13th are irregular regardless of their last digit.
	switch n % 100 {
	case 11, 12, 13:
		return "th"
	}
	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

// FinishSentence trims trailing whitespace and soft punctuation (",", ":",
// ";") from s and appends a period if s does not already end with ".", "!",
// or "?". Returns an empty string for blank input.
func FinishSentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	s = strings.TrimRight(s, ",:;")
	if !strings.HasSuffix(s, ".") && !strings.HasSuffix(s, "!") && !strings.HasSuffix(s, "?") {
		return s + "."
	}
	return s
}

// FormatSentence formats s as a complete, well-formed sentence.
func FormatSentence(s string) string {
	var p Para
	p.Continue(s)
	return p.Text()
}

// AppendClause appends clause to s, inserting a comma separator when s is
// non-empty and does not already end with one. Words from CommonSentenceStarts
// at the start of clause are lowercased to avoid mid-sentence capitalisation.
func AppendClause(s, clause string) string {
	if s == "" {
		return clause
	}
	if clause == "" {
		return s
	}
	s = strings.TrimSpace(s)
	if !strings.HasSuffix(s, ",") {
		s += ","
	}
	return s + " " + LowerIfFirstWordIn(clause, CommonSentenceStarts...)
}

// AppendAside is like AppendClause but also appends a trailing comma,
// producing a parenthetical aside: "base, clause,".
func AppendAside(s, clause string) string {
	return AppendClause(s, clause) + ","
}

// AppendRelated joins s and clause with an HTML em-dash (&mdash;), stripping
// any sentence terminator from s first. The first letter of clause is
// lowercased unconditionally.
func AppendRelated(s, clause string) string {
	if s == "" {
		return clause
	}
	if clause == "" {
		return s
	}

	s = StripTerminator(s)
	return s + "&mdash;" + LowerFirst(clause)
}

// StripTerminator removes all trailing whitespace and punctuation
// (" ", ",", ":", ";", ".", "!", "?") from s.
func StripTerminator(s string) string {
	return strings.TrimRight(s, " ,:;.!?")
}

var startsWithNumeral = regexp.MustCompile(`^[0-9]`)

// StartsWithNumeral reports whether s begins with an ASCII digit.
func StartsWithNumeral(s string) bool {
	return startsWithNumeral.MatchString(s)
}

var startsWithVowel = regexp.MustCompile(`^[aeiouAEIOU]`)

// MaybeAn returns a string intended to be appended to a literal "a" to form
// the correct indefinite article. It returns "n <s>" when s starts with a
// vowel (so "a"+"n apple" = "an apple") and " <s>" otherwise (so "a"+" book"
// = "a book").
func MaybeAn(s string) string {
	if startsWithVowel.MatchString(s) {
		return "n " + s
	}
	return " " + s
}

// MaybePluralise appends "s" to s when quantity is not 1.
func MaybePluralise(s string, quantity int) string {
	if quantity != 1 {
		return s + "s"
	}
	return s
}

// MaybePossessiveSuffix appends the correct English possessive suffix: "'"
// when s already ends in "s" (e.g. "Jones'"), or "'s" otherwise.
func MaybePossessiveSuffix(s string) string {
	if strings.HasSuffix(s, "s") {
		return s + "'"
	}
	return s + "'s"
}

// StripWasIs removes a leading "was " or "is " auxiliary verb from st,
// returning the remainder. Used to convert passive voice fragments into bare
// verb phrases when tense needs to change.
func StripWasIs(st string) string {
	if strings.HasPrefix(st, "was ") {
		return st[4:]
	}
	if strings.HasPrefix(st, "is ") {
		return st[3:]
	}
	return st
}

var containsIsolatedNumber = regexp.MustCompile(`^(.*)\b([0-9]+)\b(.*)$`)

// ReplaceFirstNumberWithCardinalNoun finds the first isolated digit sequence
// in s and replaces it with its English word form via CardinalNoun. If no
// digit sequence is found, s is returned unchanged.
func ReplaceFirstNumberWithCardinalNoun(s string) string {
	matches := containsIsolatedNumber.FindStringSubmatch(s)
	if len(matches) < 4 {
		return s
	}

	n, err := strconv.Atoi(matches[2])
	if err != nil {
		return s
	}

	return JoinSentenceParts(matches[1], CardinalNoun(n), matches[3])
}

// CommonSentenceStarts lists pronouns and articles that JoinSentenceParts and
// AppendClause lowercase when they appear at the start of a non-first part,
// preventing awkward mid-sentence capitalisation.
var CommonSentenceStarts = []string{"He", "She", "They", "His", "Her", "Their", "The", "It"}

// StripNewlines replaces every newline character in s with a space.
func StripNewlines(s string) string {
	return strings.Join(strings.Split(s, "\n"), " ")
}

// PrefixLines prepends prefix to every line in s, including the first.
func PrefixLines(s string, prefix string) string {
	return prefix + strings.Join(strings.Split(s, "\n"), "\n"+prefix)
}
