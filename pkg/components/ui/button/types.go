package button

type Variant string
type Size string

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

type ButtonProps struct {
	Class     string
	Variant   Variant
	Size      Size
	Disabled  bool
	AriaLabel string
}

func New() *ButtonProps {
	return &ButtonProps{
		Variant: Default,
		Size:    Medium,
	}
}

func (p *ButtonProps) WithVariant(v Variant) *ButtonProps {
	p.Variant = v
	return p
}

func (p *ButtonProps) WithSize(s Size) *ButtonProps {
	p.Size = s
	return p
}

func (p *ButtonProps) WithClass(c string) *ButtonProps {
	p.Class = c
	return p
}
