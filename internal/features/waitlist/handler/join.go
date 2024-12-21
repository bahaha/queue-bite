package handler

import (
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-playground/form/v4"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"

	log "queue-bite/internal/config/logger"
	d "queue-bite/internal/domain"
	hostdesk "queue-bite/internal/features/hostdesk/service"
	"queue-bite/internal/features/waitlist/domain"
	"queue-bite/internal/features/waitlist/handler/view"
	fm "queue-bite/pkg/form"
	"queue-bite/pkg/session"
	"queue-bite/pkg/utils"
)

func (h *WaitlistHandler) HandleJoinWaitlist(
	logger log.Logger,
	hostdesk hostdesk.HostDesk,
	validate *validator.Validate,
	uni *ut.UniversalTranslator,
	cookieManager *session.CookieManager,
	cookieQueudParty *session.CookieConfig,
) http.HandlerFunc {
	formDecoder := form.NewDecoder()

	type JoinWaitlistRequest struct {
		PartyName string `validate:"required"`
		PartySize int    `validate:"required,min=1"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var payload JoinWaitlistRequest

		if err := validateJoinRequest(r, formDecoder, validate, &payload); err != nil {
			copier.Copy(view.NewJoinFormData(), &payload)
			handleJoinValidationError(logger, uni, w, r, err, payload)
			return
		}

		party := d.NewParty(payload.PartyName, payload.PartySize)

		queuedParty, err := h.waitlist.JoinQueue(r.Context(), hostdesk, party)
		if err != nil {
			handleJoinError(logger, w, err)
			return
		}

		setQueuedPartyCookie(w, cookieManager, cookieQueudParty, queuedParty)
		renderQueuedParty(w, r, queuedParty)
	}
}

func validateJoinRequest(r *http.Request, formDecoder *form.Decoder, validate *validator.Validate, payload interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	if err := formDecoder.Decode(payload, r.Form); err != nil {
		return err
	}

	return validate.Struct(payload)
}

func handleJoinValidationError(
	logger log.Logger,
	uni *ut.UniversalTranslator,
	w http.ResponseWriter,
	r *http.Request,
	err error,
	payload interface{},
) {
	logger.LogErr("waitlist/join", err, "join waitlist validation failed")
	trans, _ := uni.FindTranslator(utils.CollectAcceptLanguages(r)...)
	formData := view.NewJoinFormData()
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		fm.CopyFormValueFromPayload(formData, payload)
		fm.CollectErrorsToForm(trans, formData, validationErrs)
	}
	templ.Handler(view.JoinForm(formData)).ServeHTTP(w, r)
}

func handleJoinError(
	logger log.Logger,
	w http.ResponseWriter,
	err error,
) {
	logger.LogErr("waitlist/join", err, "join waitlist failed")
	var qErr *domain.QueueOperationError
	if errors.As(err, &qErr) {
		switch qErr.Err {
		case domain.ErrPartyAlreadyQueued:
			http.Error(w, "This party is already in queue", http.StatusBadRequest)
			return
		}
	}

	http.Error(w, "Failed to join waitlist", http.StatusInternalServerError)
}

func setQueuedPartyCookie(w http.ResponseWriter, cookieManager *session.CookieManager, cookieQueuedParty *session.CookieConfig, party *domain.QueuedParty) {
	session := &PartySession{PartyID: party.ID}
	cookieManager.SetCookie(w, cookieQueuedParty, session)
}

func renderQueuedParty(w http.ResponseWriter, r *http.Request, party *domain.QueuedParty) {
	templ.Handler(view.QueuedParty(view.NewQueuedPartyProps(party))).ServeHTTP(w, r)
}
