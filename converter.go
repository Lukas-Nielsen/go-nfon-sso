package nfon

type Data struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type Link struct {
	Rel  string `json:"link"`
	Href string `json:"href"`
}

func DataToMap(data []Data) map[string]any {
	result := make(map[string]any)
	for _, entry := range data {
		result[entry.Name] = entry.Value
	}
	return result
}

func LinksToMap(data []Link) map[string]string {
	result := make(map[string]string)
	for _, entry := range data {
		result[entry.Rel] = entry.Href
	}
	return result
}
