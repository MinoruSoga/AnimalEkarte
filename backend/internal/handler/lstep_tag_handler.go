package handler

import "regexp"

// tagNamePattern remains temporarily for the L③b checkup-sync request validator.
var tagNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_\-]{1,100}$`)
