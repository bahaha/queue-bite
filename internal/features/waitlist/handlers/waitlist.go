package waitlist

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-playground/form/v4"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"

	log "queue-bite/internal/config/logger"
	view "queue-bite/internal/features/waitlist/views"
	f "queue-bite/pkg/form"
	"queue-bite/pkg/utils"
)

type waitlistHandler struct{}

func newWaitlistHandler() *waitlistHandler {
	return &waitlistHandler{}
}

func (h *waitlistHandler) JoinWaitlist(logger log.Logger, validate *validator.Validate, uni *ut.UniversalTranslator) http.HandlerFunc {
	formDecoder := form.NewDecoder()
	type JoinWaitlistRequest struct {
		PartyName string `validate:"required"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		trans, _ := uni.FindTranslator(utils.CollectAcceptLanguages(r)...)

		var joinWaitlist JoinWaitlistRequest
		if err := formDecoder.Decode(&joinWaitlist, r.Form); err != nil {
			logger.LogErr(FEAT_WAITLIST, err, "Could not decode request form data")
			joinWaitlistForm := view.NewJoinFormData()
			templ.Handler(view.JoinForm(joinWaitlistForm)).ServeHTTP(w, r)
			return
		}

		if err := validate.Struct(joinWaitlist); err != nil {
			errs := err.(validator.ValidationErrors)

			joinWaitlistForm := view.NewJoinFormData()
			f.CopyFormValueFromPayload(joinWaitlistForm, joinWaitlist)
			f.CollectErrorsToForm(trans, joinWaitlistForm, errs)
			logger.LogErr(FEAT_WAITLIST, err, "Invalid join waitlist request")

			templ.Handler(view.JoinForm(joinWaitlistForm)).ServeHTTP(w, r)
			return
		}

		logger.LogInfo(FEAT_WAITLIST, "valid join waitlist request", "form", joinWaitlist)
	}
}
