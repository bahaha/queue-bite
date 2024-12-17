package form

import (
	"context"
	"queue-bite/pkg/form"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormItem(t *testing.T) {
	t.Parallel()

	t.Run("render basic form item", func(t *testing.T) {
		comp := FormItem(NewFormItemProps())
		html := renderToString(t, comp)

		assert.Contains(t, html, `"space-y-2"`)
		assert.Contains(t, html, `data-role="form-item"`)
	})

	t.Run("renders with error message when invalid", func(t *testing.T) {
		ctx := &form.FormItemContext{
			ID:           "test-form__item",
			Invalid:      true,
			ErrorMessage: "This field is required",
		}

		comp := FormItem(NewFormItemProps().WithFormItem(ctx))
		html := renderToString(t, comp)

		assert.Contains(t, html, `This field is required`)
	})

	t.Run("does not render error when valid", func(t *testing.T) {
		ctx := &form.FormItemContext{
			ID:           "test-form__item",
			Invalid:      false,
			ErrorMessage: "Never render me",
		}

		comp := FormItem(NewFormItemProps().WithFormItem(ctx))
		html := renderToString(t, comp)

		assert.NotContains(t, html, "Never render me")
	})
}

func renderToString(t *testing.T, comp templ.Component) string {
	var sb strings.Builder
	err := comp.Render(context.Background(), &sb)
	require.NoError(t, err)
	return sb.String()
}
