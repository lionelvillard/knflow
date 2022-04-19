package types

type SW struct {
    ID string `yaml:"id"`
    Start interface{} `yaml:"start"`
    States []State `yaml:"states"`
}

type State struct {
    Name string
    Type string
    Data interface{}
    End interface{}
}
