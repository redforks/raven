package raven

import (
	"context"
	"fmt"

	raven "github.com/getsentry/raven-go"
	"github.com/redforks/appinfo"
	"github.com/redforks/errors"
	"github.com/redforks/life"
)

// Info interface provide extra info error report need. Implement this
// interface in your session object
type Info interface {
	// Return current username, if user not logged in return empty string
	Username() string // Name as spork/session.Session.Username

	// Return current custom name, return empty string if not available.
	CustomerName() string
}

// HandleError handles non-nil err value, get error cause, send to errrpt
// service by config option. err type is interface{} to allow handle return
// value of recover() function.
func HandleError(ctx context.Context, err interface{}) {
	if e, ok := err.(error); ok {
		onError(ctx, e)
		return
	}

	onOther(ctx, err)
}

func resolveDID() string {
	return appinfo.InstallID()
}

func resolveCausedBy(err error) string {
	return errors.GetCausedBy(err).String()
}

func onError(ctx context.Context, err error) {
	if !needReport(errors.GetCausedBy(err)) {
		return
	}

	did := resolveDID()
	raven.CaptureError(err, map[string]string{
		"CausedBy": resolveCausedBy(err),
		"DID":      did,
	}, nil)
}

func onOther(ctx context.Context, err interface{}) {
	msg := fmt.Sprint(err)
	SendMessage(ctx, msg)
}

// SendMessage send message to sentry.io.
//
// Normally errors send to sentry, by using redforks/errors.Handle(),
// but some errors we want pay attention to, not golang errors, use
// SendMessage().
//
// Better send these messages to GA, but GA interface not done yet.
// Crash/critical messages goto sentry, track/trending messages should goto GA.
// Use GA to analyse potential problems, these problems are not hurry, fix them
// good for User Experience.
func SendMessage(ctx context.Context, msg string) {
	did := resolveDID()
	raven.CaptureMessage(msg, map[string]string{
		"CausedBy": "Other",
		"DID":      did,
	}, nil)
}

func init() {
	errors.SetHandler(HandleError)
	life.RegisterHook("raven", 0, life.OnAbort, raven.DefaultClient.Wait)
	life.Register("raven", func() {
		raven.SetRelease(appinfo.Version())
	}, nil)
}
