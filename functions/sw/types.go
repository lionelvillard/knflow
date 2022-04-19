package function

import (
    "math/rand"
    "strings"
    "unicode"
)

type SW struct {
    ID     string `json:"id"`
    Start  string
    States []State `json:"states"`
}

type State struct {
    Name string
    Type string
    Data interface{}
    End  interface{}
}

const (
    letterBytes   = "abcdefghijklmnopqrstuvwxyz"
    randSuffixLen = 8
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

// RandomString will generate a random string.
func RandomString() string {
    suffix := make([]byte, randSuffixLen)

    for i := range suffix {
        suffix[i] = letterBytes[rand.Intn(len(letterBytes))]
    }
    return string(suffix)
}
