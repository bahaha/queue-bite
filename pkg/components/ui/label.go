package ui

import (
	"context"
	"queue-bite/pkg/components/ui/form"
	"queue-bite/pkg/utils"

	"github.com/a-h/templ"
)

type labelProps struct {
	ctx   context.Context
	ctxID string

	Class    string
	Required bool
}

func LabelProps() *labelProps {
	return &labelProps{}
}

func (p *labelProps) WithClass(class string) *labelProps {
	p.Class = class
	return p
}

func (p *labelProps) WithRequired(required bool) *labelProps {
	p.Required = required
	return p
}

func (p *labelProps) WithinContext(ctx context.Context, id string) *labelProps {
	p.ctx = ctx
	p.ctxID = id
	return p
}

type styleVariants struct {
	Base     string
	Required string
}

var defaultLabelStyles = styleVariants{
	Base:     "text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70",
	Required: "after:content-['*'] after:ml-0.5 after:text-destructive",
}

func NewLabel(props *labelProps) templ.Attributes {
	attrs := templ.Attributes{
		"class": utils.Cn(
			defaultLabelStyles.Base,
			utils.AppendClass(defaultLabelStyles.Required, props.Required),
			props.Class,
		),
	}

	if props.ctx != nil && props.ctxID != "" {
		formItem := form.GetFormItemContext(props.ctx, props.ctxID)
		attrs["for"] = formItem.ID
	}

	return attrs
}
