package fpbun

type NameConverter interface {
	Convert(name string) string
}

type NameConverterFunc func(string) string

func (nc NameConverterFunc) Convert(name string) string {
	return nc(name)
}

func PassthroughNameConverter(name string) string {
	return name
}
