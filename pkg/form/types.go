package form

type FormItemContext struct {
	ID           string
	Name         string
	Value        interface{}
	Invalid      bool
	ErrorMessage string
}
