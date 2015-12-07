package confluence

import (
	"net/http"
)

// ContentService handles communication with the content related methods of the
// Confluence API.
type ContentService struct {
	client *Client
}

// The Space the content lives within
type Space struct {
	// The string rep of the space. When a new space is created, e.g. 'Product
	// Development', it is assigned a unique key (e.g. 'PD').
	Key string `json:"key,omitempty"`
}

// The Body content stored in Value with format defined in Representation.
type Storage struct {
	Value          string `json:"value,omitempty"`
	Representation string `json:"representation,omitempty"`
}

// The body of a Content represented by a Storage.
type Body struct {
	Storage Storage `json:"storage,omitempty"`
}

// A piece of content. Either a page or a blog post.
type Content struct {
	// the content type to return. Default value: page. Valid values: page,
	// blogpost. Default is page.
	Type string `json:"type,omitempty"`

	// the title of the page to find. Required for page type
	Title string `json:"title,omitempty"`

	Space Space `json:"space,omitempty"`
	Body  Body  `json:"body,omitempty"`
}

// Creates a new piece of Content.
func (s *ContentService) Create(content *Content) (*Content, *http.Response, error) {
	req, err := s.client.NewRequest("POST", "wiki/rest/api/content/", content)
	if err != nil {
		return nil, nil, err
	}

	returnedContent := new(Content)
	resp, err := s.client.Do(req, returnedContent)
	if err != nil {
		return nil, resp, err
	}

	return returnedContent, resp, err
}
