package model

type Page struct {
	HtmlVersion            string         `json:"htmlVersion"`
	Title                  string         `json:"title"`
	HeadingsCount          map[string]int `json:"headingsCountByLevel"`
	Links                  []Link         `json:"links"`
	InternalLinksCount     int            `json:"internalLinksCount"`
	ExternalLinksCount     int            `json:"externalLinksCount"`
	InaccessibleLinksCount int            `json:"inaccessibleLinksCount"`
	HasLoginForm           bool           `json:"hasLoginForm"`
	PasswordFieldCount     map[string]int `json:"passwordFieldCount"`
}

type Link struct {
	Url          string
	StatusCode   int
	IsInternal   bool
	IsAccessible bool
}
