package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func GetVersion() string {
	if e, err := strconv.ParseInt(buildTimestamp, 10, 64); err == nil {
		buildTimestamp = time.Unix(e, 0).Format(time.RFC3339)
	}
	return fmt.Sprintf("%s built from %s at %s",
		strings.TrimPrefix(version, "v"), gitVersion,
		buildTimestamp)
}

var (
	version        = "v1.2.0"
	gitVersion     = "unknown-git-version"
	buildTimestamp = "unknown-time"
)
