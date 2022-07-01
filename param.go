package golam

type PathParam struct {
	Key   string
	Value string
}

type PathParams map[string]PathParam

// TODO 바인딩 추가 필요
