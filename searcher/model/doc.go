package model

type IndexDoc struct {
	Id       uint32                 `json:"id,omitempty"`
	Text     string                 `json:"text,omitempty"`
	Document map[string]interface{} `json:"document,omitempty"`
}

type StorageIndexDoc struct {
	*IndexDoc
	Keys []string `json:"keys,omitempty"`
}

type ResponseDoc struct {
	IndexDoc
	OriginalText string   `json:"originalText,omitempty"`
	Score        int      `json:"score,omitempty"`
	Keys         []string `json:"keys,omitempty"`
}

type ResponseDocsSort []ResponseDoc

func (r ResponseDocsSort) Len() int {
	return len(r)
}

func (r ResponseDocsSort) Less(i, j int) bool {
	return r[i].Score < r[j].Score
}

func (r ResponseDocsSort) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
