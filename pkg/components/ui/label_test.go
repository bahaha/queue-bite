package ui

import (
	"context"
	"queue-bite/pkg/form"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStyledLabel(t *testing.T) {
	t.Parallel()

	t.Run("standalone label", func(t *testing.T) {
		t.Parallel()
		attrs := NewLabel(LabelProps())
		class, ok := attrs["class"].(string)
		assert.True(t, ok, "should have `class` attribute to style the label")

		expected := "peer-disabled:cursor-not-allowed"
		assert.Contains(t, class, expected, "default label should contain class: %s", expected)
	})

	t.Run("label with required indicator", func(t *testing.T) {
		t.Parallel()
		attrs := NewLabel(LabelProps().WithRequired(true))
		class, ok := attrs["class"].(string)
		assert.True(t, ok, "should have `class` attribute to style the label")

		expected := "after:content-['*']"
		assert.Contains(t, class, expected, "required indicator should contain class: %s", expected)
	})

	t.Run("label with custom class", func(t *testing.T) {
		t.Parallel()
		attrs := NewLabel(LabelProps().WithClass("bg-secondary"))
		class, ok := attrs["class"].(string)
		assert.True(t, ok, "should have `class` attribute to style the label")

		expected := "bg-secondary"
		assert.Contains(t, class, expected, "custom label should contain class: %s", expected)
	})

	t.Run("connected label", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		formItem := &form.FormItemContext{
			ID:      "test-label--email",
			Name:    "Email",
			Value:   "abc@example.com",
			Invalid: true,
		}
		ctx = context.WithValue(ctx, formItem.ID, formItem)
		attrs := NewLabel(LabelProps().WithinContext(ctx, formItem.ID))

		htmlFor, ok := attrs["for"].(string)
		assert.True(t, ok, "should have `id` attribute to style the label")
		assert.Equal(t, "test-label--email", htmlFor, "should have correct for to connected ID")
	})
}
