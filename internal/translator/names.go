package translator

import (
    "strings"
    "unicode"
)

// MakeK8sName converts each chunk of non-alphanumeric character into a single dash
// and also convert camelcase tokens into dash-delimited lowercase tokens.
func MakeK8sName(s string) string {
    var sb strings.Builder
    newToken := false
    for _, c := range s {
        if !(unicode.IsLetter(c) || unicode.IsNumber(c)) {
            newToken = true
            continue
        }
        if sb.Len() > 0 && (newToken || unicode.IsUpper(c)) {
            sb.WriteRune('-')
        }
        sb.WriteRune(unicode.ToLower(c))
        newToken = false
    }
    return sb.String()
}
