package nfon

import (
	"encoding/json"
)

type Data struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type Response struct {
	Href   string            `json:"href"`
	Offset int               `json:"offset"`
	Total  int               `json:"total"`
	Size   int               `json:"size"`
	Links  map[string]string `json:"links"`
	Data   map[string]any    `json:"data"`
	Items  []Item            `json:"items"`
}

type Item struct {
	Href  string            `json:"href"`
	Links map[string]string `json:"links"`
	Data  map[string]any    `json:"data"`
}

func (r *Response) UnmarshalJSON(data []byte) error {
	var aux struct {
		Href   string `json:"href"`
		Offset int    `json:"offset"`
		Total  int    `json:"total"`
		Size   int    `json:"size"`
		Links  []Link `json:"links"`
		Data   []Data `json:"data"`
		Items  []struct {
			Href  string `json:"href"`
			Links []Link `json:"links"`
			Data  []Data `json:"data"`
		} `json:"items"`
	}
	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	r.Href = aux.Href
	r.Offset = aux.Offset
	r.Total = aux.Total
	r.Size = aux.Size
	r.Links = make(map[string]string)
	for _, link := range aux.Links {
		r.Links[link.Rel] = link.Href
	}

	r.Data = make(map[string]any)
	for _, d := range aux.Data {
		r.Data[d.Name] = d.Value
	}

	r.Items = make([]Item, 0, len(aux.Items))
	for _, item := range aux.Items {
		i := Item{
			Href:  item.Href,
			Links: make(map[string]string),
			Data:  make(map[string]any),
		}
		for _, link := range item.Links {
			i.Links[link.Rel] = link.Href
		}
		for _, d := range item.Data {
			i.Data[d.Name] = d.Value
		}
		r.Items = append(r.Items, i)
	}
	return nil
}
