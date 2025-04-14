package nfon

type Data struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type Links struct {
	Rel  string `json:"link"`
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
	result := make(map[string]any)
	for _, entry := range data {
		result[entry.Name] = entry.Value
	}
	return result
}

func LinksToMap(data []Links) map[string]string {
	result := make(map[string]string)
	for _, entry := range data {
		result[entry.Rel] = entry.Href
	}
	return result
}

func (r response) parse() Response {
	var c Response
	c.Href = r.Href
	c.Offset = r.Offset
	c.Size = r.Size
	c.Total = r.Total

	c.Links = LinksToMap(r.Links)
	c.Data = DataToMap(r.Data)

	var t []Items
	for _, e := range r.Items {
		e := e
		t = append(t, Items{
			Href:  e.Href,
			Links: LinksToMap(e.Links),
			Data:  DataToMap(e.Data),
		})
	}
	c.Items = t

	return c
}
