package main

import (
    "fmt"
    "knflows/sw/internal/translator"
    "os"
)

func main() {
    if len(os.Args) != 2 {
        fmt.Println("missing file argument")
        os.Exit(1)
    }

    filename := os.Args[1]

    err := translator.Translate(filename)
    if err != nil {
        fmt.Println(err)
    }
}
