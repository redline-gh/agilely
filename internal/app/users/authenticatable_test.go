package users

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ibraheemdev/agilely/internal/app/engine"
	"github.com/ibraheemdev/agilely/test"
)

func TestEngineInit(t *testing.T) {
	t.Parallel()

	e := engine.New()

	router := &test.Router{}
	renderer := &test.Renderer{}
	errHandler := &test.ErrorHandler{}
	e.Config.Core.Router = router
	e.Config.Core.ViewRenderer = renderer
	e.Config.Core.ErrorHandler = errHandler

	u := &Users{}
	if err := u.Init(e); err != nil {
		t.Fatal(err)
	}

	if err := renderer.HasLoadedViews(PageLogin); err != nil {
		t.Error(err)
	}

	if err := router.HasGets("/login"); err != nil {
		t.Error(err)
	}
	if err := router.HasPosts("/login"); err != nil {
		t.Error(err)
	}
}

func TestAuthGet(t *testing.T) {
	t.Parallel()

	ab := engine.New()
	responder := &test.Responder{}
	ab.Config.Core.Responder = responder

	a := &Users{ab}

	r := test.Request("GET")
	r.URL.RawQuery = "redir=/redirectpage"
	if err := a.LoginGet(nil, r); err != nil {
		t.Error(err)
	}

	if responder.Page != PageLogin {
		t.Error("wanted login page, got:", responder.Page)
	}

	if responder.Status != http.StatusOK {
		t.Error("wanted ok status, got:", responder.Status)
	}

	if got := responder.Data[engine.FormValueRedirect]; got != "/redirectpage" {
		t.Error("redirect page was wrong:", got)
	}
}

type testHarness struct {
	users *Users
	ab    *engine.Engine

	bodyReader *test.BodyReader
	responder  *test.Responder
	redirector *test.Redirector
	session    *test.ClientStateRW
	storer     *test.ServerStorer
}

func testSetup() *testHarness {
	harness := &testHarness{}

	harness.ab = engine.New()
	harness.bodyReader = &test.BodyReader{}
	harness.redirector = &test.Redirector{}
	harness.responder = &test.Responder{}
	harness.session = test.NewClientRW()
	harness.storer = test.NewServerStorer()

	harness.ab.Config.Core.BodyReader = harness.bodyReader
	harness.ab.Config.Core.Logger = test.Logger{}
	harness.ab.Config.Core.Responder = harness.responder
	harness.ab.Config.Core.Redirector = harness.redirector
	harness.ab.Config.Storage.SessionState = harness.session
	harness.ab.Config.Storage.Server = harness.storer

	harness.users = &Users{harness.ab}

	return harness
}

func TestAuthPostSuccess(t *testing.T) {
	t.Parallel()

	setupMore := func(h *testHarness) *testHarness {
		h.bodyReader.Return = test.Values{
			PID:      "test@test.com",
			Password: "hello world",
		}
		h.storer.Users["test@test.com"] = &test.User{
			Email:    "test@test.com",
			Password: "$2a$10$IlfnqVyDZ6c1L.kaA/q3bu1nkAC6KukNUsizvlzay1pZPXnX2C9Ji", // hello world
		}
		h.session.ClientValues[engine.SessionHalfAuthKey] = "true"

		return h
	}

	t.Run("normal", func(t *testing.T) {
		t.Parallel()
		h := setupMore(testSetup())

		var beforeCalled, afterCalled bool
		var beforeHasValues, afterHasValues bool
		h.ab.Events.Before(engine.EventAuth, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
			beforeCalled = true
			beforeHasValues = r.Context().Value(engine.CTXKeyValues) != nil
			return false, nil
		})
		h.ab.Events.After(engine.EventAuth, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
			afterCalled = true
			afterHasValues = r.Context().Value(engine.CTXKeyValues) != nil
			return false, nil
		})

		r := test.Request("POST")
		resp := httptest.NewRecorder()
		w := h.ab.NewResponse(resp)

		if err := h.users.LoginPost(w, r); err != nil {
			t.Error(err)
		}

		if resp.Code != http.StatusTemporaryRedirect {
			t.Error("code was wrong:", resp.Code)
		}
		if h.redirector.Options.RedirectPath != "/" {
			t.Error("redirect path was wrong:", h.redirector.Options.RedirectPath)
		}

		if _, ok := h.session.ClientValues[engine.SessionHalfAuthKey]; ok {
			t.Error("half auth should have been deleted")
		}
		if pid := h.session.ClientValues[engine.SessionKey]; pid != "test@test.com" {
			t.Error("pid was wrong:", pid)
		}

		if !beforeCalled {
			t.Error("before should have been called")
		}
		if !afterCalled {
			t.Error("after should have been called")
		}
		if !beforeHasValues {
			t.Error("before callback should have access to values")
		}
		if !afterHasValues {
			t.Error("after callback should have access to values")
		}
	})

	t.Run("handledBefore", func(t *testing.T) {
		t.Parallel()
		h := setupMore(testSetup())

		var beforeCalled bool
		h.ab.Events.Before(engine.EventAuth, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
			w.WriteHeader(http.StatusTeapot)
			beforeCalled = true
			return true, nil
		})

		r := test.Request("POST")
		resp := httptest.NewRecorder()
		w := h.ab.NewResponse(resp)

		if err := h.users.LoginPost(w, r); err != nil {
			t.Error(err)
		}

		if h.responder.Status != 0 {
			t.Error("a status should never have been sent back")
		}
		if _, ok := h.session.ClientValues[engine.SessionKey]; ok {
			t.Error("session key should not have been set")
		}

		if !beforeCalled {
			t.Error("before should have been called")
		}
		if resp.Code != http.StatusTeapot {
			t.Error("should have left the response alone once teapot was sent")
		}
	})

	t.Run("handledAfter", func(t *testing.T) {
		t.Parallel()
		h := setupMore(testSetup())

		var afterCalled bool
		h.ab.Events.After(engine.EventAuth, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
			w.WriteHeader(http.StatusTeapot)
			afterCalled = true
			return true, nil
		})

		r := test.Request("POST")
		resp := httptest.NewRecorder()
		w := h.ab.NewResponse(resp)

		if err := h.users.LoginPost(w, r); err != nil {
			t.Error(err)
		}

		if h.responder.Status != 0 {
			t.Error("a status should never have been sent back")
		}
		if _, ok := h.session.ClientValues[engine.SessionKey]; !ok {
			t.Error("session key should have been set")
		}

		if !afterCalled {
			t.Error("after should have been called")
		}
		if resp.Code != http.StatusTeapot {
			t.Error("should have left the response alone once teapot was sent")
		}
	})
}

func TestAuthPostBadPassword(t *testing.T) {
	t.Parallel()

	setupMore := func(h *testHarness) *testHarness {
		h.bodyReader.Return = test.Values{
			PID:      "test@test.com",
			Password: "world hello",
		}
		h.storer.Users["test@test.com"] = &test.User{
			Email:    "test@test.com",
			Password: "$2a$10$IlfnqVyDZ6c1L.kaA/q3bu1nkAC6KukNUsizvlzay1pZPXnX2C9Ji", // hello world
		}

		return h
	}

	t.Run("normal", func(t *testing.T) {
		t.Parallel()
		h := setupMore(testSetup())

		r := test.Request("POST")
		resp := httptest.NewRecorder()
		w := h.ab.NewResponse(resp)

		var afterCalled bool
		h.ab.Events.After(engine.EventAuthFail, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
			afterCalled = true
			return false, nil
		})

		if err := h.users.LoginPost(w, r); err != nil {
			t.Error(err)
		}

		if resp.Code != 200 {
			t.Error("wanted a 200:", resp.Code)
		}

		if h.responder.Data[engine.DataErr] != "Invalid Credentials" {
			t.Error("wrong error:", h.responder.Data)
		}

		if _, ok := h.session.ClientValues[engine.SessionKey]; ok {
			t.Error("user should not be logged in")
		}

		if !afterCalled {
			t.Error("after should have been called")
		}
	})

	t.Run("handledAfter", func(t *testing.T) {
		t.Parallel()
		h := setupMore(testSetup())

		r := test.Request("POST")
		resp := httptest.NewRecorder()
		w := h.ab.NewResponse(resp)

		var afterCalled bool
		h.ab.Events.After(engine.EventAuthFail, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
			w.WriteHeader(http.StatusTeapot)
			afterCalled = true
			return true, nil
		})

		if err := h.users.LoginPost(w, r); err != nil {
			t.Error(err)
		}

		if h.responder.Status != 0 {
			t.Error("responder should not have been called to give a status")
		}
		if _, ok := h.session.ClientValues[engine.SessionKey]; ok {
			t.Error("user should not be logged in")
		}

		if !afterCalled {
			t.Error("after should have been called")
		}
		if resp.Code != http.StatusTeapot {
			t.Error("should have left the response alone once teapot was sent")
		}
	})
}

func TestAuthPostUserNotFound(t *testing.T) {
	t.Parallel()

	harness := testSetup()
	harness.bodyReader.Return = test.Values{
		PID:      "test@test.com",
		Password: "world hello",
	}

	r := test.Request("POST")
	resp := httptest.NewRecorder()
	w := harness.ab.NewResponse(resp)

	// This event is really the only thing that separates "user not found"
	// from "bad password"
	var afterCalled bool
	harness.ab.Events.After(engine.EventAuthFail, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
		afterCalled = true
		return false, nil
	})

	if err := harness.users.LoginPost(w, r); err != nil {
		t.Error(err)
	}

	if resp.Code != 200 {
		t.Error("wanted a 200:", resp.Code)
	}

	if harness.responder.Data[engine.DataErr] != "Invalid Credentials" {
		t.Error("wrong error:", harness.responder.Data)
	}

	if _, ok := harness.session.ClientValues[engine.SessionKey]; ok {
		t.Error("user should not be logged in")
	}

	if afterCalled {
		t.Error("after should not have been called")
	}
}