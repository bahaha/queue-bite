package view

import (
	"github.com/jinzhu/copier"

	d "queue-bite/internal/domain"
	"queue-bite/internal/features/seatmanager/domain"
	wld "queue-bite/internal/features/waitlist/domain"
)

func ToVitrineProps(
	queuedParty *wld.QueuedParty,
	status *wld.QueueStatus,
	totalCapacity int,
) *VitrinePageData {
	pageProps := &VitrinePageData{}

	if status != nil {
		pageProps.QueueStatus = status
	}

	if queuedParty != nil {
		pageProps.QueuedPartyProps = NewQueuedPartyProps(queuedParty)
	} else {
		pageProps.Form = NewJoinFormData(totalCapacity)
	}

	return pageProps
}

func NewQueuedPartyProps(party *wld.QueuedParty) *QueuedPartyProps {
	props := &QueuedPartyProps{}
	copier.Copy(props, party)
	props.ReadyForSeating = props.Status == d.PartyStatusReady
	return props
}

func NewReadyPartyProps(partyID d.PartyID) *QueuedPartyProps {
	props := &QueuedPartyProps{QueuedParty: &wld.QueuedParty{Party: &d.Party{}}}
	props.ID = partyID
	props.ReadyForSeating = true
	return props
}

func NewYummyProps(session *domain.PartySession) *YummyProps {
	return &YummyProps{ID: session.ID, Name: session.Name, Size: session.Size}
}
