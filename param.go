package golam

type PathParam struct {
	Key   string
	Value string
}

type PathParams map[string]PathParam

func (p PathParams) Get(key string) string {
	param, ok := p[key]
	if !ok {
		return ""
	}

	return param.Value
}

func (p PathParams) Set(key string, param PathParam) {
	p[key] = param
}

// TODO 바인딩 추가 필요
