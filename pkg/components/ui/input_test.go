package ui

import (
	"context"
	"queue-bite/pkg/form"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStyledInput(t *testing.T) {
	t.Parallel()

	t.Run("standalone input with default props", func(t *testing.T) {
		t.Parallel()
		attrs := NewInput(InputProps())
		class, ok := attrs["class"].(string)
		assert.True(t, ok, "should have `class` attribute to style the input")

		expected := "border-input"
		assert.Contains(t, class, expected, "default input should contain class: %s", expected)
	})

	t.Run("input with error state", func(t *testing.T) {
		t.Parallel()
		attrs := NewInput(InputProps().WithError(true))
		class, ok := attrs["class"].(string)
		assert.True(t, ok, "should have `class` attribute to style the input")

		expected := "ring-destructive"
		assert.Contains(t, class, expected, "input should contain error style class: %s", expected)
	})

	t.Run("custom class", func(t *testing.T) {
		t.Parallel()
		attrs := NewInput(InputProps().WithClass("w-full"))
		class, ok := attrs["class"].(string)
		assert.True(t, ok, "should have `class` attribute to style the input")

		expected := "w-full"
		assert.Contains(t, class, expected, "input should contain custom class: %s", expected)
	})

	t.Run("connected with a form item", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		formItem := &form.FormItemContext{
			ID:      "test-input--name",
			Name:    "Email",
			Value:   "abc@example.com",
			Invalid: true,
		}
		ctx = context.WithValue(ctx, formItem.ID, formItem)

		attrs := NewInput(InputProps().WithinContext(ctx, formItem.ID))

		id, ok := attrs["id"].(string)
		assert.True(t, ok, "should have `id` attribute to style the input")
		assert.Equal(t, "test-input--name", id, "should have correct ID")

		name, ok := attrs["name"].(string)
		assert.True(t, ok, "should have `name` attribute to style the input")
		assert.Equal(t, "Email", name, "should have correct name from the model")

		value, ok := attrs["value"].(string)
		assert.True(t, ok, "should have `name` attribute to style the input")
		assert.Equal(t, "abc@example.com", value, "should have value from the model")

		class, ok := attrs["class"].(string)
		assert.True(t, ok, "should have `class` attribute to style the input")
		expected := "ring-destructive"
		assert.Contains(t, class, expected, "input connected with an invalid model should have error class %s", expected)
	})
}
