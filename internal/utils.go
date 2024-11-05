package internal

import (
    "fmt"
)

func LogInfo(message string) {
    fmt.Println("[INFO]", message)
}

func LogError(err error) {
    fmt.Println("[ERROR]", err)
}

