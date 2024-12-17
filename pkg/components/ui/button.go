package ui

import (
	"queue-bite/pkg/utils"

	"github.com/a-h/templ"
)

type Variant string
type Size string

type buttonProps struct {
	Class   string
	Variant Variant
	Size    Size
}

func ButtonProps() *buttonProps {
	return &buttonProps{
		Variant: Button.Variants.Default,
		Size:    Button.Sizes.Medium,
	}
}

func (p *buttonProps) WithClass(class string) *buttonProps {
	p.Class = class
	return p
}

func (p *buttonProps) WithVariant(variant Variant) *buttonProps {
	p.Variant = variant
	return p
}

func (p *buttonProps) WithSize(size Size) *buttonProps {
	p.Size = size
	return p
}

type buttonVariants struct {
	Default     Variant
	Secondary   Variant
	Destructive Variant
	Outline     Variant
	Ghost       Variant
	Link        Variant
}

type buttonSizes struct {
	Small  Size
	Medium Size
	Large  Size
}

var Button = struct {
	Variants buttonVariants
	Sizes    buttonSizes
}{
	Variants: buttonVariants{
		Default:     "default",
		Secondary:   "secondary",
		Destructive: "destructive",
		Outline:     "outline",
		Ghost:       "ghost",
		Link:        "link",
	},
	Sizes: buttonSizes{
		Small:  "sm",
		Medium: "md",
		Large:  "lg",
	},
}

type ButtonVariants struct {
	Base     string
	Variants map[Variant]string
	Sizes    map[Size]string
}

var defaultButtonStyles = ButtonVariants{
	Base: "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
	Variants: map[Variant]string{
		Button.Variants.Default:     "bg-primary text-primary-foreground hover:bg-primary/90",
		Button.Variants.Secondary:   "bg-secondary text-secondary-foreground hover:bg-secondary/80",
		Button.Variants.Destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90",
		Button.Variants.Outline:     "border border-input bg-background hover:bg-accent hover:text-accent-foreground",
		Button.Variants.Ghost:       "hover:bg-accent hover:text-accent-foreground",
		Button.Variants.Link:        "text-primary underline-offset-4 hover:underline",
	},

	Sizes: map[Size]string{
		Button.Sizes.Small:  "h-9 px-3 text-sm",
		Button.Sizes.Medium: "h-10 px-4 py-2",
		Button.Sizes.Large:  "h-11 px-8 text-lg",
	},
}

// NewButton generates Templ attributes especially the TailwindCSS classes for styling a button based on the props.
// The button attributes that are not listed in the ButtonProps could just write on the button element for simplicity.
// If you need custimize class for the button, use WithClass instead, or it will ignore the button variant styles.
//
// Example:
//
//	<button
//	    type="submit"
//	    aria-label="Delete item"
//	    {ui.NewButton(ui.ButtonProps().
//	        WithVariant(ui.Button.Variants.Outline).
//	        WithSize(ui.Button.Sizes.Large)...}
//	>
//	    Delete
//	</button>
func NewButton(props *buttonProps) templ.Attributes {
	attrs := templ.Attributes{}

	attrs["class"] = utils.Cn(
		defaultButtonStyles.Base,
		defaultButtonStyles.Variants[props.Variant],
		defaultButtonStyles.Sizes[props.Size],
		props.Class,
	)

	return attrs
}
