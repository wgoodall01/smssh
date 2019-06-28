package util

import "log"

func Fatal(msg string, err error) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}
