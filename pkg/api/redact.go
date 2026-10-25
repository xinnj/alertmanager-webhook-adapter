package api

import "net/url"

const redactedValue = "REDACTED"

// redactURL returns the URL with credential-bearing query parameters masked,
// so it is safe to write to debug logs.
func redactURL(u *url.URL) string {
	redacted := *u
	q := redacted.Query()
	for _, key := range []string{"token", "secret", "sign"} {
		if q.Has(key) {
			q.Set(key, redactedValue)
		}
	}
	redacted.RawQuery = q.Encode()

	return redacted.String()
}
