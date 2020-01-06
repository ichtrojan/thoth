package main

import (
	"github.com/ichtrojan/thoth"
    "errors"
)

func main() {
	logger := thoth.Init("./logs/logger.log.txt")
    err := errors.New("Test Custom Error log")
    logger.Log(err)
}
