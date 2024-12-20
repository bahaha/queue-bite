package view

import (
	"queue-bite/internal/features/waitlist/domain"

	"github.com/jinzhu/copier"
)

func ToVitrineProps(
	queuedParty *domain.QueuedParty,
	status *domain.QueueStatus,
) *VitrinePageData {
	pageProps := &VitrinePageData{}

	if status != nil {
		pageProps.QueueStatus = status
	}

	if queuedParty != nil {
		props := &QueuedPartyProps{}
		copier.Copy(props, queuedParty)
		pageProps.QueuedPartyProps = props
	} else {
		pageProps.Form = NewJoinFormData()
	}

	return pageProps
}
