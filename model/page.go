package model

type Page struct {
	HtmlVersion   string         `json:"htmlVersion"`
	Title         string         `json:"title"`
	HeadingsCount map[string]int `json:"headingsCountByLevel"`
	Links         []Link         `json:"links"`
	LinksCount    map[string]int `json:"linksCount"`
	HasLoginForm  bool           `json:"hasLoginForm"`
}

type Link struct {
	Url          string `json:"url"`
	StatusCode   int    `json:"statusCode"`
	IsInternal   bool   `json:"isInternal"`
	IsAccessible bool   `json:"isAccessible"`
}
