package text

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func LowerFirst(s string) string {
	r := []rune(s)

	return strings.ToLower(string(r[0])) + string(r[1:])
}

func LowerIfFirstWordIn(s string, words ...string) string {
	for _, w := range words {
		if strings.HasPrefix(s, w+" ") {
			return LowerFirst(s)
		}
	}
	return s
}

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

func RemoveRedundantWhitespace(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

func RemoveAllWhitespace(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), "")
}

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
		return "fourtieth"
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
		return "one hundreth"
	}

	return CardinalNoun(n/10) + "-" + ordinalNounUnderTen(n%10)
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

func JoinList(strs []string) string {
	var ret string
	for i, s := range strs {
		s = strings.Trim(s, " ,!.?")

		if i != 0 {
			if i == len(strs)-1 {
				ret += " and "
			} else {
				ret += ", "
			}
		}
		ret += s
	}
	return ret
}

func JoinListOr(strs []string) string {
	var ret string
	for i, s := range strs {
		s = strings.Trim(s, " ,!.?")

		if i != 0 {
			if i == len(strs)-1 {
				ret += " or "
			} else {
				ret += ", "
			}
		}
		ret += s
	}
	return ret
}

func JoinSentenceParts(parts ...string) string {
	var ret string
	for _, s := range parts {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			continue
		}
		if ret != "" && s != ":" {
			ret += " "
		}
		ret += LowerIfFirstWordIn(s, CommonSentenceStarts...)
	}
	return ret
}

func JoinSentences(ss ...string) string {
	var ret string
	for _, s := range ss {
		s = FormatSentence(s)
		// s = strings.TrimSpace(s)
		if len(s) == 0 {
			continue
		}
		if ret != "" {
			ret += " "
		}
		ret += s
	}
	return ret
}

func AppendSentence(base, s string) string {
	s = FormatSentence(s)
	if len(base) == 0 {
		return s
	}
	if !strings.HasSuffix(base, " ") {
		base += " "
	}
	return base + s
}

func CardinalWithUnit(n int, singular string, plural string) string {
	if n == 1 {
		return "one " + singular
	}
	return CardinalNoun(n) + " " + plural
}

func CardinalSuffix(n int) string {
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

func FormatSentence(s string) string {
	return UpperFirst(FinishSentence(s))
}

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

func AppendAside(s, clause string) string {
	return AppendClause(s, clause) + ","
}

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

func StripTerminator(s string) string {
	return strings.TrimRight(s, ",:;.!?")
}

func AppendIndependentClause(s, clause string) string {
	if s == "" {
		return clause
	}
	if clause == "" {
		return s
	}
	s = strings.TrimSpace(s)
	s = StripTerminator(s)
	return s + "; " + clause
}

var startsWithNumeral = regexp.MustCompile(`^[0-9]`)

func StartsWithNumeral(s string) bool {
	return startsWithNumeral.MatchString(s)
}

var startsWithVowel = regexp.MustCompile(`^[aeiouAEIOU]`)

func MaybeAn(s string) string {
	if startsWithVowel.MatchString(s) {
		return "n " + s
	}
	return " " + s
}

func MaybePluralise(s string, quantity int) string {
	if quantity != 1 {
		return s + "s"
	}
	return s
}

func MaybePossessiveSuffix(s string) string {
	if strings.HasSuffix(s, "s") {
		return s + "'"
	}
	return s + "'s"
}

// TODO: remove MaybeWasVerb
func MaybeWasVerb(verb string) string {
	fs := strings.Fields(verb)
	if len(fs) == 0 {
		return verb
	}
	switch fs[0] {
	case "born", "baptised", "buried", "cremated", "executed", "lost", "killed", "promoted", "demoted":
		return "was " + verb
	default:
		return verb
	}
}

func MaybeHaveBeenVerb(verb string) string {
	st := MaybeWasVerb(verb)
	if strings.HasPrefix(st, "was ") {
		return "have been " + st[4:]
	}
	return "have " + st
}

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

var CommonSentenceStarts = []string{"He", "She", "They", "His", "Her", "Their", "The", "It"}

func StripNewlines(s string) string {
	return strings.Join(strings.Split(s, "\n"), " ")
}

func PrefixLines(s string, prefix string) string {
	return prefix + strings.Join(strings.Split(s, "\n"), "\n"+prefix)
}
