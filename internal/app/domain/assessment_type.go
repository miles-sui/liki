package domain

// AssessmentType classifies an assessment as self or peer evaluation.
type AssessmentType string

const (
	AssessSelf AssessmentType = "self"
	AssessPeer AssessmentType = "peer"
)
