package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStyledButton(t *testing.T) {
	t.Parallel()

	t.Run("default button variants", func(t *testing.T) {
		attrs := NewButton(ButtonProps())
		class, ok := attrs["class"].(string)
		assert.True(t, ok, "should have `class` attribute to style the button")

		expected := "bg-primary"
		assert.Contains(t, class, expected, "default button should contain class: %s", expected)
	})

	t.Run("specific button variant", func(t *testing.T) {
		attrs := NewButton(ButtonProps().
			WithVariant(Button.Variants.Destructive).
			WithSize(Button.Sizes.Large))
		class, ok := attrs["class"].(string)
		assert.True(t, ok, "should have `class` attribute to style the button")

		expected := "bg-destructive"
		assert.Contains(t, class, expected, "destructive button should contain class: %s", expected)

		expected = "text-lg"
		assert.Contains(t, class, expected, "large button should contain class: %s", expected)
	})

	t.Run("custom class", func(t *testing.T) {
		attrs := NewButton(ButtonProps().
			WithClass("bg-sky-100"))
		class, ok := attrs["class"].(string)
		assert.True(t, ok, "should have `class` attribute to style the button")

		expected := "bg-sky-100"
		assert.Contains(t, class, expected, "custom button should contain class: %s", expected)
	})
}
