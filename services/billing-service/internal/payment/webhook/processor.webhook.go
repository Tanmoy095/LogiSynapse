package webhook

import "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/payment"

type Processor interface {
	Provider() string
	VerifyAndParse( //it verifies the webhook signature and parses the event
		payload []byte, //raw webhook payload . it comes from HTTP request body
		headers map[string]string, //HTTP headers containing signature info
	) (*payment.NormalizedEvent, error)
}
