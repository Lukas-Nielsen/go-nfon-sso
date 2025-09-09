package nfon

type Data struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type Links struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type response struct {
	Href   string  `json:"href"`
	Offset int     `json:"offset"`
	Total  int     `json:"total"`
	Size   int     `json:"size"`
	Links  []Links `json:"links"`
	Data   []Data  `json:"data"`
	Items  []struct {
		Href  string  `json:"href"`
		Links []Links `json:"links"`
		Data  []Data  `json:"data"`
	} `json:"items"`
}

type Response struct {
	Href   string
	Offset int
	Total  int
	Size   int
	Links  map[string]string
	Data   map[string]any
	Items  []Items
}

type Items struct {
	Href  string
	Links map[string]string
	Data  map[string]any
}

func DataToMap(data []Data) map[string]any {
	result := make(map[string]any, len(data))
	for _, entry := range data {
		result[entry.Name] = entry.Value
	}
	return result
}

func LinksToMap(data []Links) map[string]string {
	result := make(map[string]string, len(data))
	for _, entry := range data {
		result[entry.Rel] = entry.Href
	}
	return result
}

func (r response) parse() Response {
	resp := Response{
		Href:   r.Href,
		Offset: r.Offset,
		Size:   r.Size,
		Total:  r.Total,
		Links:  LinksToMap(r.Links),
		Data:   DataToMap(r.Data),
		Items:  make([]Items, 0, len(r.Items)),
	}
	for _, e := range r.Items {
		resp.Items = append(resp.Items, Items{
			Href:  e.Href,
			Links: LinksToMap(e.Links),
			Data:  DataToMap(e.Data),
		})
	}
	return resp
}
