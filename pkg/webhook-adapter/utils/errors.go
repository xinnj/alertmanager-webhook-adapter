package utils

import (
	"errors"
	"fmt"
	"net/url"
)

// SanitizeRequestError removes request URLs from errors returned by
// http.Client.Do. Outbound URLs can carry credentials (bot tokens, signatures,
// access tokens), so they must not be written to logs or responses.
func SanitizeRequestError(err error) error {
	var urlErr *url.Error
	if !errors.As(err, &urlErr) {
		return err
	}

	return fmt.Errorf("%s: %v", urlErr.Op, SanitizeRequestError(urlErr.Err))
}
