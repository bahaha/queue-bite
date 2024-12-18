package waitlist

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-playground/form/v4"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"

	"queue-bite/internal/config"
	log "queue-bite/internal/config/logger"
	view "queue-bite/internal/features/waitlist/views"
	f "queue-bite/pkg/form"
	"queue-bite/pkg/session"
	"queue-bite/pkg/utils"
)

type waitlistHandler struct{}

func newWaitlistHandler() *waitlistHandler {
	return &waitlistHandler{}
}

type QueuedParty struct {
	ID   string
	Name string
}

// Handle form submission for joining a restaurant's waitlist.
// The handler performs the following steps:
//  1. Receives form data for party registration
//  2. Validates required party information
//  3. On validation failure, re-displays form with error message in user's language
//  4. On success, returns confirmation with party's position in line
//
// Request Payload (Form Data):
//   - PartyName: requried
//
// Flow:
//
//	Browser -> Submit Form -> Validate ─┬─ Invalid ─> Show Errors
//	                                    └─ Valid ───> Show Position
func (h *waitlistHandler) JoinWaitlist(
	logger log.Logger,
	validate *validator.Validate,
	uni *ut.UniversalTranslator,
	cookieManager *session.CookieManager,
	cookieCfgs *config.QueueBiteCookies,
) http.HandlerFunc {
	formDecoder := form.NewDecoder()
	type JoinWaitlistRequest struct {
		PartyName string `validate:"required"`
	}

	var setQueuedPartyCookie func(w http.ResponseWriter, PartyName string)
	setQueuedPartyCookie = func(w http.ResponseWriter, partyName string) {
		cookieManager.SetCookie(w, cookieCfgs.QueuedPartyCookie, &QueuedParty{
			ID:   utils.GenerateUID(),
			Name: partyName,
		})
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

		logger.LogDebug(FEAT_WAITLIST, "valid join waitlist request", "form", joinWaitlist)
		// return a success partial page with the order of the party
		setQueuedPartyCookie(w, joinWaitlist.PartyName)
	}
}
