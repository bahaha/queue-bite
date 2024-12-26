package view

import (
	"github.com/jinzhu/copier"

	d "queue-bite/internal/domain"
	"queue-bite/internal/features/waitlist/domain"
)

func ToVitrineProps(
	queuedParty *domain.QueuedParty,
	status *domain.QueueStatus,
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

func NewQueuedPartyProps(party *domain.QueuedParty) *QueuedPartyProps {
	props := &QueuedPartyProps{}
	copier.Copy(props, party)
	props.ReadyForSeating = props.Status == d.PartyStatusReady
	return props
}
