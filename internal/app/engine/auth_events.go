package engine

import (
	"net/http"
)

//go:generate stringer -output stringers.go -type "AuthEvent"

// AuthEvent type is for describing events
type AuthEvent int

// AuthEvent kinds
const (
	EventRegister AuthEvent = iota
	EventAuth
	// EventAuthHijack is used to steal the authentication process after a
	// successful auth but before any session variable has been put in.
	// Most useful for defining an additional step for authentication
	// (like 2fa). It needs to be separate to EventAuth because other modules
	// do checks that would also interrupt event handlers with an authentication
	// failure so there's an ordering problem.
	EventAuthHijack
	EventOAuth2
	EventAuthFail
	EventOAuth2Fail
	EventRecoverStart
	EventRecoverEnd
	EventGetUser
	EventGetUserSession
	EventPasswordReset
	EventLogout
)

// AuthEventHandler reacts to events that are fired by Engine controllers.
// These controllers will normally process a request by themselves, but if
// there is special consideration for example a successful login, but the
// user is locked, the lock module's controller may seize control over the
// request.
//
// Very much a controller level middleware.
type AuthEventHandler func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error)

// AuthEvents is a collection of AuthEvents that fire before and after certain methods.
type AuthEvents struct {
	before map[AuthEvent][]AuthEventHandler
	after  map[AuthEvent][]AuthEventHandler
}

// NewAuthEvents creates a new set of before and after AuthEvents.
func NewAuthEvents() *AuthEvents {
	return &AuthEvents{
		before: make(map[AuthEvent][]AuthEventHandler),
		after:  make(map[AuthEvent][]AuthEventHandler),
	}
}

// Before event, call f.
func (c *AuthEvents) Before(e AuthEvent, f AuthEventHandler) {
	events := c.before[e]
	events = append(events, f)
	c.before[e] = events
}

// After event, call f.
func (c *AuthEvents) After(e AuthEvent, f AuthEventHandler) {
	events := c.after[e]
	events = append(events, f)
	c.after[e] = events
}

// FireBefore executes the handlers that were registered to fire before
// the event passed in.
//
// If it encounters an error it will stop immediately without calling
// other handlers.
//
// If a handler handles the request, it will pass this information both
// to handlers further down the chain (to let them know that w has been used)
// as well as set w to nil as a precaution.
func (c *AuthEvents) FireBefore(e AuthEvent, w http.ResponseWriter, r *http.Request) (bool, error) {
	return c.call(c.before[e], w, r)
}

// FireAfter event to all the AuthEvents with a context. The error can safely be
// ignored as it is logged.
func (c *AuthEvents) FireAfter(e AuthEvent, w http.ResponseWriter, r *http.Request) (bool, error) {
	return c.call(c.after[e], w, r)
}

func (c *AuthEvents) call(evs []AuthEventHandler, w http.ResponseWriter, r *http.Request) (bool, error) {
	handled := false

	for _, fn := range evs {
		interrupt, err := fn(w, r, handled)
		if err != nil {
			return false, err
		}
		if interrupt {
			handled = true
		}
	}

	return handled, nil
}