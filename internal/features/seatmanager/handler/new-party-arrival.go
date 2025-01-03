package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-playground/form/v4"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"

	log "queue-bite/internal/config/logger"
	d "queue-bite/internal/domain"
	hd "queue-bite/internal/features/hostdesk/service"
	"queue-bite/internal/features/seatmanager/domain"
	"queue-bite/internal/features/seatmanager/handler/view"
	"queue-bite/internal/features/seatmanager/service"
	w "queue-bite/internal/features/waitlist/domain"
	fm "queue-bite/pkg/form"
	"queue-bite/pkg/session"
	"queue-bite/pkg/utils"
)

var SEAT_MANAGER_ARRIVAL = "seatmanager/arrival"

func (*seatManagerHandler) HandleNewPartyArrival(
	logger log.Logger,
	validate *validator.Validate,
	uni *ut.UniversalTranslator,
	cookieManager *session.CookieManager,
	cookieQueudParty *session.CookieConfig,
	seatManager service.SeatManager,
	hostdesk hd.HostDesk,
) http.HandlerFunc {
	formDecoder := form.NewDecoder()
	totalCapacity, _ := hostdesk.GetTotalCapacity(context.Background())

	type NewPartyArrivalRequest struct {
		PartyName string `validate:"required"`
		PartySize int    `validate:"required,min=1"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var payload NewPartyArrivalRequest
		if err := validateNewPartyArrivalRequest(r, formDecoder, validate, &payload); err != nil {
			handleNewPartyArrivalValidationError(logger, uni, w, r, err, payload, totalCapacity)
			return
		}

		if payload.PartySize > totalCapacity {
			formData := view.NewJoinFormData(totalCapacity)
			fm.CopyFormValueFromPayload(formData, payload)
			formData.PartySize.Invalid = true
			formData.PartySize.ErrorMessage = fmt.Sprintf("Sorry, we could only take reservation under people %d right now.", totalCapacity)
			templ.Handler(view.JoinForm(formData)).ServeHTTP(w, r)
			return
		}

		party := d.NewParty(d.PartyID(utils.GenerateID()), payload.PartyName, payload.PartySize)
		queuedParty, err := seatManager.ProcessNewParty(r.Context(), party)
		if err != nil {
			handleErrorOnNewPartyArrival(logger, w, r, payload, totalCapacity, err)
			return
		}

		setQueuedPartyCookie(w, cookieManager, cookieQueudParty, queuedParty)
		renderQueuedParty(w, r, queuedParty)
	}
}

func validateNewPartyArrivalRequest(r *http.Request, formDecoder *form.Decoder, validate *validator.Validate, payload interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	if err := formDecoder.Decode(payload, r.Form); err != nil {
		return err
	}

	return validate.Struct(payload)
}

func handleNewPartyArrivalValidationError(
	logger log.Logger,
	uni *ut.UniversalTranslator,
	w http.ResponseWriter,
	r *http.Request,
	err error,
	payload interface{},
	totalCapacity int,
) {
	logger.LogErr(SEAT_MANAGER_ARRIVAL, err, "join waitlist validation failed")
	trans, _ := uni.FindTranslator(utils.CollectAcceptLanguages(r)...)
	formData := view.NewJoinFormData(totalCapacity)
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		fm.CopyFormValueFromPayload(formData, payload)
		fm.CollectErrorsToForm(trans, formData, validationErrs)
	}
	templ.Handler(view.JoinForm(formData)).ServeHTTP(w, r)
}

func handleErrorOnNewPartyArrival(
	logger log.Logger,
	resp http.ResponseWriter,
	req *http.Request,
	payload interface{},
	totalCapacity int,
	err error,
) {
	logger.LogErr(SEAT_MANAGER_ARRIVAL, err, "handle new party arrival failed")
	switch err {
	case domain.ErrPreserveSeats:
	case domain.ErrJoinWaitlist:
		formData := view.NewJoinFormData(totalCapacity)
		fm.CopyFormValueFromPayload(formData, payload)
		formData.ErrorMessage = "Failed to preserve seats or join waitlist, please try again later."
		templ.Handler(view.JoinForm(formData)).ServeHTTP(resp, req)
		return
	case w.ErrPartyAlreadyQueued:
		http.Error(resp, "This party is already in queue", http.StatusBadRequest)
		return
	}

	http.Error(resp, "Failed to join waitlist", http.StatusInternalServerError)
}

func setQueuedPartyCookie(w http.ResponseWriter, cookieManager *session.CookieManager, cookieQueuedParty *session.CookieConfig, party *w.QueuedParty) {
	session := &domain.PartySession{ID: party.ID, Name: party.Name, Size: party.Size}
	cookieManager.SetCookie(w, cookieQueuedParty, session)
}

func renderQueuedParty(w http.ResponseWriter, r *http.Request, party *w.QueuedParty) {
	templ.Handler(view.QueuedParty(view.NewQueuedPartyProps(party))).ServeHTTP(w, r)
}
