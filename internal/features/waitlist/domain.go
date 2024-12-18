package waitlist

type Party struct {
	ID   string
	Name string
	Size int
}

func NewPartyToJoin(name string, size int) *Party {
	return &Party{
		Name: name,
		Size: size,
	}
}
