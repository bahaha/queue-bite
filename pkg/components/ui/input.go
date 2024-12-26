package ui

import (
	"context"
	"queue-bite/pkg/components/ui/form"
	"queue-bite/pkg/utils"

	"github.com/a-h/templ"
)

type inputProps struct {
	ctx   context.Context
	ctxID string

	Class    string
	HasError bool
}

func InputProps() *inputProps {
	return &inputProps{}
}

func (p *inputProps) WithinContext(ctx context.Context, id string) *inputProps {
	p.ctx = ctx
	p.ctxID = id
	return p
}

func (p *inputProps) WithError(error bool) *inputProps {
	p.HasError = error
	return p
}

func (p *inputProps) WithClass(class string) *inputProps {
	p.Class = class
	return p
}

type InputVariants struct {
	Base     string
	HasError string
}

var defaultInputStyles = InputVariants{
	Base:     "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-base ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
	HasError: "ring-2 ring-offset-2 ring-destructive focus-visible:ring-destructive",
}

// NewInput generate Templ attributes for an input based on the props.
// It handles both standalone inputs and form-connected inputs automatically.
// When connected to a form item, it will
//   - Auto-bind form item ID, name, and value
//   - Sync error states with the form validation of form model
//
// For standalone usage:
//
//	 <input type="text"
//			placeholder="Enter text"
//			{ui.NewInput(ui.InputProps().WithError(hasError))...}
//	 />
//
// For form-connected usage:
//
//	 @form.FormItem(form.NewFormItemProps().WithFormItem(formData.Name)) {
//		   <input type="text"
//			      placeholder="Enter text"
//	              {ui.NewInput(ui.InputProps().WithinContext(ctx, formData.Name.ID))...}
//		    />
//	 }
func NewInput(props *inputProps) templ.Attributes {
	attrs := templ.Attributes{}
	hasError := props.HasError

	if props.ctx != nil && props.ctxID != "" {
		formItem := form.GetFormItemContext(props.ctx, props.ctxID)
		attrs["id"] = formItem.ID
		attrs["name"] = formItem.Name
		attrs["value"] = formItem.Value
		hasError = formItem.Invalid
	}

	attrs["class"] = utils.Cn(
		defaultInputStyles.Base,
		utils.AppendClass(defaultInputStyles.HasError, hasError),
		props.Class,
	)

	return attrs
}
