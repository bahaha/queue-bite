package form

import (
	"context"

	f "queue-bite/pkg/form"
)

type FormItemProps struct {
	Ctx   *f.FormItemContext
	Class string
}

func NewFormItemProps() *FormItemProps {
	return &FormItemProps{}
}

func (p *FormItemProps) WithClass(class string) *FormItemProps {
	p.Class = class
	return p
}

func (p *FormItemProps) WithFormItem(item *f.FormItemContext) *FormItemProps {
	p.Ctx = item
	return p
}

func AttachFormItemContext(parent context.Context, itemCtx *f.FormItemContext) context.Context {
	return context.WithValue(parent, itemCtx.ID, itemCtx)
}

func GetFormItemContext(ctx context.Context, ctxID string) *f.FormItemContext {
	if ctx == nil {
		return &f.FormItemContext{}
	}

	if itemCtx, ok := ctx.Value(ctxID).(*f.FormItemContext); ok {
		return itemCtx
	}

	return &f.FormItemContext{}
}
