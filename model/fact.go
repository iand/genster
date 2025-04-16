package model

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	FactCategoryAKA                   = "Also known as"
	FactCategoryMilitaryServiceNumber = "Military service number"
	FactCategorySeamansTicket         = "Seaman's ticket"
	FactCategoryCauseOfDeath          = "Cause of death"
	FactCategoryLiteracy              = "Literacy"
)

type Fact struct {
	Category  string
	Detail    string
	Comment   string // an explanatatory comment to be used alongside or as a footnote to the detail
	Citations []*GeneralCitation
}

type AssociationKind string

const (
	AssociationKindTwin AssociationKind = "twin"
	AssociationKindDNA  AssociationKind = "DNA"
)

type Association struct {
	Kind      AssociationKind
	Other     *Person
	Citations []*GeneralCitation
}

func ParseCauseOfDeathFact(text string, citations []*GeneralCitation) *Fact {
	comment := ""
	text = strings.ToLower(text)
	switch text {
	case "paralysis":
		comment = "commonly caused by a stroke"
	case "colitis":
		comment = "inflammation of the colon"
	case "gastritis":
		comment = "inflammation of the stomach"
	case "hepatitis":
		comment = "inflammation of the liver"
	case "hysteritis":
		comment = "inflammation of the womb"
	case "peritonitis":
		comment = "inflammation of the abdomen"
	case "myocarditis":
		comment = "inflammation of the heart"
	case "dysentery":
		comment = "inflammation of the intestine"
	case "colic":
		comment = "abdominal pain and cramp"
	case "phthisis", "consumption", "marasmus":
		comment = "tuberculosis"
	case "ascites":
		comment = "a build up of fluid in the abdomen caused by heart failure or kidney disease"
	case "dropsy":
		comment = "a swelling caused by accumulation of abnormally large amounts of fluid often caused by kidney disease or congestive heart failure"
	case "lockjaw", "trismus":
		comment = "tetanus"
	case "natural decay":
		comment = "death through old age"
	case "senile decay":
		comment = "death through old age, possibly including dementia"
	case "morbus cordis":
		comment = "heart disease"
	case "albuminuria":
		comment = "kidney disease"
	case "palsy":
		comment = "paralysis"
	case "apoplexy":
		comment = "incapacity resulting from a stroke"
	case "erysipelas":
		comment = "a skin infection, commonly known as St. Anthony's Fire due to the burning sensation experienced"
	case "tabes dorsalis":
		comment = "degeneration of nerve cells in the lower back caused by untreated syphilis"
	case "dyspnoea":
		comment = "shortness of breath"
	case "placenta previa":
		comment = "a problem of pregnancy in which the placenta covers all or part of the cervix"
	case "scarlatina":
		comment = "scarlet fever"
	case "uraemia":
		comment = "high levels of urea in blood"
	case "scirrhus of the womb":
		comment = "a tumour in the womb"
	case "debility":
		comment = "weakness or feebleness"
	case "syncope":
		comment = "fainting"
	}

	if comment != "" {
		text = fmt.Sprintf("%q (%s)", text, comment)
	} else if strings.Contains(text, " ") {
		text = fmt.Sprintf("%q", text)
	}

	return &Fact{
		Category:  FactCategoryCauseOfDeath,
		Detail:    text,
		Citations: citations,
	}
}

type Link struct {
	Title string
	URL   string
}

func LinkFromURL(u string) *Link {
	pu, err := url.Parse(u)
	if err != nil {
		return &Link{
			Title: u,
			URL:   u,
		}
	}

	return &Link{
		Title: pu.Host,
		URL:   u,
	}
}
