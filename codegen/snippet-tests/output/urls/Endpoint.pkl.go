// Code generated from Pkl module `Urls`. DO NOT EDIT.
package urls

import "net/url"

type Endpoint struct {
	Url url.URL `pkl:"url"`

	Fallback *url.URL `pkl:"fallback"`

	Uri url.URL `pkl:"uri"`
}
