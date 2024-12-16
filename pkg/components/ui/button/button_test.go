package button

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestButtonProps(t *testing.T) {
	t.Parallel()

	t.Run("default props", func(t *testing.T) {
		t.Parallel()

		props := New()
		assert.Equal(t, Default, props.Variant)
		assert.Equal(t, Medium, props.Size)
	})

	t.Run("custom props", func(t *testing.T) {
		t.Parallel()

		props := New().
			WithVariant(Destructive).
			WithSize(Large).
			WithClass("w-full")

		assert.Equal(t, Destructive, props.Variant)
		assert.Equal(t, Large, props.Size)
		assert.Equal(t, "w-full", props.Class)
	})
}
