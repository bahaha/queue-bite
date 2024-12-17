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
		Variant: Default,
		Size:    Medium,
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

const (
	Default     Variant = "default"
	Secondary   Variant = "secondary"
	Destructive Variant = "destructive"
	Outline     Variant = "outline"
	Ghost       Variant = "ghost"
	Link        Variant = "link"
)

const (
	Small  Size = "sm"
	Medium Size = "md"
	Large  Size = "lg"
)

type ButtonVariants struct {
	Base     string
	Variants map[Variant]string
	Sizes    map[Size]string
}

var defaultButtonStyles = ButtonVariants{
	Base: "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
	Variants: map[Variant]string{
		Default:     "bg-primary text-primary-foreground hover:bg-primary/90",
		Secondary:   "bg-secondary text-secondary-foreground hover:bg-secondary/80",
		Destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90",
		Outline:     "border border-input bg-background hover:bg-accent hover:text-accent-foreground",
		Ghost:       "hover:bg-accent hover:text-accent-foreground",
		Link:        "text-primary underline-offset-4 hover:underline",
	},

	Sizes: map[Size]string{
		Small:  "h-9 px-3 text-sm",
		Medium: "h-10 px-4 py-2",
		Large:  "h-11 px-8 text-lg",
	},
}

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
