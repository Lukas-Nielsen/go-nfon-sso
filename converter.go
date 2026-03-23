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

func (r *Response) MarshalJSON() ([]byte, error) {
	aux := struct {
		Href   string `json:"href,omitempty"`
		Offset int    `json:"offset,omitempty"`
		Total  int    `json:"total,omitempty"`
		Size   int    `json:"size,omitempty"`
		Links  []Link `json:"links,omitempty"`
		Data   []Data `json:"data,omitempty"`
		Items  []struct {
			Href  string `json:"href,omitempty"`
			Links []Link `json:"links,omitempty"`
			Data  []Data `json:"data,omitempty"`
		} `json:"items,omitempty"`
	}{
		Href:   r.Href,
		Offset: r.Offset,
		Total:  r.Total,
		Size:   r.Size,
	}

	for rel, href := range r.Links {
		aux.Links = append(aux.Links, Link{Rel: rel, Href: href})
	}

	for name, value := range r.Data {
		aux.Data = append(aux.Data, Data{Name: name, Value: value})
	}

	for _, it := range r.Items {
		item := struct {
			Href  string `json:"href,omitempty"`
			Links []Link `json:"links,omitempty"`
			Data  []Data `json:"data,omitempty"`
		}{
			Href: it.Href,
		}
		for rel, href := range it.Links {
			item.Links = append(item.Links, Link{Rel: rel, Href: href})
		}
		for name, value := range it.Data {
			item.Data = append(item.Data, Data{Name: name, Value: value})
		}
		aux.Items = append(aux.Items, item)
	}

	return json.Marshal(&aux)
}
