package onion

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func TestOnionWith(t *testing.T) {
	result := ""
	response := httptest.NewRecorder()

	n1 := New()
	n1.Use(HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.Handler) {
		result = "one"
		next.ServeHTTP(rw, r)
	}))
	n1.Use(HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.Handler) {
		result += "two"
		next.ServeHTTP(rw, r)
	}))

	n1.ServeHTTP(response, (*http.Request)(nil))
	expect(t, 2, len(n1.Handlers()))
	expect(t, result, "onetwo")

	n2 := n1.With(HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.Handler) {
		result += "three"
		next.ServeHTTP(rw, r)
	}))

	// Verify that n1 was left intact and not modified.
	n1.ServeHTTP(response, (*http.Request)(nil))
	expect(t, 2, len(n1.Handlers()))
	expect(t, result, "onetwo")

	n2.ServeHTTP(response, (*http.Request)(nil))
	expect(t, 3, len(n2.Handlers()))
	expect(t, result, "onetwothree")
}

func TestOnionWith_doNotModifyOriginal(t *testing.T) {
	result := ""
	response := httptest.NewRecorder()

	n1 := New()
	n1.handlers = make([]Handler, 0, 10) // enforce initial capacity
	n1.Use(HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.Handler) {
		result = "one"
		next.ServeHTTP(rw, r)
	}))

	n1.ServeHTTP(response, (*http.Request)(nil))
	expect(t, 1, len(n1.Handlers()))

	n2 := n1.With(HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.Handler) {
		result += "two"
		next.ServeHTTP(rw, r)
	}))
	n3 := n1.With(HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.Handler) {
		result += "three"
		next.ServeHTTP(rw, r)
	}))

	// rebuilds middleware
	n2.UseHandlerFunc(func(rw http.ResponseWriter, r *http.Request) {})
	n3.UseHandlerFunc(func(rw http.ResponseWriter, r *http.Request) {})

	n1.ServeHTTP(response, (*http.Request)(nil))
	expect(t, 1, len(n1.Handlers()))
	expect(t, result, "one")

	n2.ServeHTTP(response, (*http.Request)(nil))
	expect(t, 3, len(n2.Handlers()))
	expect(t, result, "onetwo")

	n3.ServeHTTP(response, (*http.Request)(nil))
	expect(t, 3, len(n3.Handlers()))
	expect(t, result, "onethree")
}

func TestOnionServeHTTP(t *testing.T) {
	result := ""
	response := httptest.NewRecorder()

	n := New()
	n.Use(HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.Handler) {
		result += "foo"
		next.ServeHTTP(rw, r)
		result += "ban"
	}))
	n.Use(HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.Handler) {
		result += "bar"
		next.ServeHTTP(rw, r)
		result += "baz"
	}))
	n.Use(HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.Handler) {
		result += "bat"
		rw.WriteHeader(http.StatusBadRequest)
	}))

	n.ServeHTTP(response, (*http.Request)(nil))

	expect(t, result, "foobarbatbazban")
	expect(t, response.Code, http.StatusBadRequest)
}

// Ensures that a Onion middleware chain
// can correctly return all of its handlers.
func TestHandlers(t *testing.T) {
	response := httptest.NewRecorder()
	n := New()
	handlers := n.Handlers()
	expect(t, 0, len(handlers))

	n.Use(HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.Handler) {
		rw.WriteHeader(http.StatusOK)
	}))

	// Expects the length of handlers to be exactly 1
	// after adding exactly one handler to the middleware chain
	handlers = n.Handlers()
	expect(t, 1, len(handlers))

	// Ensures that the first handler that is in sequence behaves
	// exactly the same as the one that was registered earlier
	handlers[0].ServeHTTP(response, (*http.Request)(nil), nil)
	expect(t, response.Code, http.StatusOK)
}

func TestOnion_Use_Nil(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Errorf("Expected onion.Use(nil) to panic, but it did not")
		}
	}()

	n := New()
	n.Use(nil)
}

func voidHTTPHandlerFunc(rw http.ResponseWriter, r *http.Request) {
	// Do nothing
}

// Test for function Wrap
func TestWrap(t *testing.T) {
	response := httptest.NewRecorder()

	handler := Wrap(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(response, (*http.Request)(nil), http.HandlerFunc(voidHTTPHandlerFunc))

	expect(t, response.Code, http.StatusOK)
}

// Test for function WrapFunc
func TestWrapFunc(t *testing.T) {
	response := httptest.NewRecorder()

	// WrapFunc(f) equals Wrap(http.HandlerFunc(f)), it's simpler and usefull.
	handler := WrapFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	handler.ServeHTTP(response, (*http.Request)(nil), http.HandlerFunc(voidHTTPHandlerFunc))

	expect(t, response.Code, http.StatusOK)
}

type voidHandler struct{}

func (vh *voidHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.Handler) {
	next.ServeHTTP(rw, r)
}

func BenchmarkOnion(b *testing.B) {
	h0 := &voidHandler{}
	h1 := &voidHandler{}
	h2 := &voidHandler{}
	h3 := &voidHandler{}
	h4 := &voidHandler{}
	h5 := &voidHandler{}
	h6 := &voidHandler{}
	h7 := &voidHandler{}
	h8 := &voidHandler{}
	h9 := &voidHandler{}

	n := New(h0, h1, h2, h3, h4, h5, h6, h7, h8, h9)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n.ServeHTTP(nil, nil)
	}
}
