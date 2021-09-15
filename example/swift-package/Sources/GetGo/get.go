package Go

import (
	"io"
	"net/http"
)

// Get makes an HTTP(S) GET request to url,
// returning the resulting content or an error.
func Get(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

// Version returns the version string for this package.
func Version() string {
	return "0.0.1"
}
