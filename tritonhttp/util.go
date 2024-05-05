package tritonhttp

import (
	"net/textproto"
	"time"
)

// CanonicalHeaderKey returns the canonical format of the
// header key s. The canonicalization converts the first
// letter and any letter following a hyphen to upper case;
// the rest are converted to lowercase. For example, the
// canonical key for "content-type" is "Content-Type".
// You should store header keys in the canonical format
// in internal data structures.
func CanonicalHeaderKey(s string) string {
	return textproto.CanonicalMIMEHeaderKey(s)
}

// FormatTime formats time according to the HTTP spec.
// It is like time.RFC1123 but hard-codes GMT as the time zone.
// You should use this function for the "Date" and "Last-Modified"
// headers.
func FormatTime(t time.Time) string {
	s := t.UTC().Format(time.RFC1123)
	s = s[:len(s)-3] + "GMT"
	return s
}
