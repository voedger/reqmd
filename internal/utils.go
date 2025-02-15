package internal

import (
	"fmt"
	"log"
	"strings"
)

var IsVerbose bool

func Verbose(msg string, kv ...string) {
	if !IsVerbose {
		return
	}

	pairs := make([]string, 0, len(kv)/2)
	for i := 0; i < len(kv)-1; i += 2 {
		pairs = append(pairs, fmt.Sprintf("%s=%s", kv[i], kv[i+1]))
	}

	if len(kv)%2 != 0 {
		pairs = append(pairs, fmt.Sprintf("%s=<missing>", kv[len(kv)-1]))
	}

	if len(pairs) > 0 {
		log.Printf("%s [%s]", msg, strings.Join(pairs, ", "))
	} else {
		log.Println(msg)
	}
}
