package button

type StyleConfig struct {
	Base     string
	Variants map[Variant]string
	Sizes    map[Size]string
}

var defaultStyles = StyleConfig{
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
