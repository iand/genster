package model

const (
	InferenceTypeYearOfBirth = "Year of birth"
	InferenceTypeYearOfDeath = "Year of death"
)

type Inference struct {
	Type   string
	Value  string
	Reason string
}

func (inf *Inference) AsCitation() *GeneralCitation {
	return &GeneralCitation{
		Detail: inf.Type + " inferred to be " + inf.Value + " because " + inf.Reason,
	}
}
