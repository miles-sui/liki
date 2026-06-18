// Package qiming provides起名 computation.
//
// Types
//   Wuxing
//   WuGeData
//   NameCandidate, Evaluation
//   WuGe, SanCai
//   StrokeCombo
//   Character, CharLite, Phonetic
//
// Functions
//   PrepareWuGe(surname, yongShen, xiShen) → (WuGeData, error)
//   ComposeNames(surname, combos, yongChars, xiChars) → []string
//   DetailNames(surname, names) → ([]NameCandidate, error)
//   EvaluateName(surname, given, yong) → (Evaluation, error)
package qiming
