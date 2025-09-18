package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// index a post document to the configured ES index
func indexPostToES(p *Post) error {
	body := map[string]any{
		"id":      p.ID,
		"title":   p.Title,
		"content": p.Content,
		"tags":    p.Tags,
	}
	b, _ := json.Marshal(body)
	res, err := ESClient.Index(esIndex, bytes.NewReader(b), ESClient.Index.WithDocumentID(fmt.Sprint(p.ID)))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("es index error: %s", res.String())
	}
	return nil
}
