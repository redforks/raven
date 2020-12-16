package raven

import (
	"log"
	"strings"
	"sync"

	"github.com/redforks/config"
	"github.com/redforks/errors"
	"github.com/redforks/testing/reset"
)

const tag = "raven"

var (
	lCauses = sync.Mutex{}
	causes  []errors.CausedBy
)

type option struct {
	ReportCause []string // what error causes will report to errrpt service.
}

func (o *option) parseReportCause() ([]errors.CausedBy, error) {
	causes := make([]errors.CausedBy, 0, len(o.ReportCause))
	var badCauses []string
	for _, s := range o.ReportCause {
		cause := parseCausedBy(s)
		if cause == errors.NoError {
			badCauses = append(badCauses, s)
			continue
		}
		causes = append(causes, cause)
	}

	if badCauses != nil {
		return nil, errors.Runtimef("[%s] Unknown ReportCause: %s", tag, strings.Join(badCauses, ", "))
	}
	return causes, nil
}

func (o *option) Init() error {
	cus, err := o.parseReportCause()
	if err != nil {
		return err
	}

	lCauses.Lock()
	causes = cus
	lCauses.Unlock()
	log.Printf("[%s] Set ReportCause to: %s", tag, strings.Join(o.ReportCause, ", "))
	return nil
}

func (o *option) Apply() {
	cus, err := o.parseReportCause()
	if err != nil {
		log.Printf("[%s] Error applying ReportCause: %s", tag, strings.Join(o.ReportCause, ", "))
		return
	}

	lCauses.Lock()
	causes = cus
	log.Printf("[%s] Apply ReportCause to: %s", tag, strings.Join(o.ReportCause, ", "))
	lCauses.Unlock()
}

// NewDefaultOption returns default errrpt option
func newDefaultOption() config.Option {
	return &option{
		ReportCause: []string{"Bug", "Runtime"},
	}
}

func parseCausedBy(s string) errors.CausedBy {
	switch s {
	case "Bug":
		return errors.ByBug
	case "Runtime":
		return errors.ByRuntime
	case "External":
		return errors.ByExternal
	case "Input":
		return errors.ByInput
	default:
		return errors.NoError
	}
}

// NeedReport returns true if the error reason cause need report to errrpt server.
func needReport(cause errors.CausedBy) bool {
	r := false
	lCauses.Lock()
	for _, c := range causes {
		if c == cause {
			r = true
			break
		}
	}
	lCauses.Unlock()
	return r
}

func init() {
	config.Register(tag, newDefaultOption)

	reset.Register(nil, func() {
		causes = nil
	})
}
