package form

import "github.com/a-h/templ"

type FormItemStyleVariant struct {
	Base string
}

var formItemAttrs = templ.Attributes{
	"data-role": "form-item",
}

var defaultStyles = FormItemStyleVariant{
	Base: "space-y-2",
}
