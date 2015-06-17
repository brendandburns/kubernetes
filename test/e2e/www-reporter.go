package e2e

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

type wwwReporter struct {
	sync.Mutex
	
	Config config.GinkgoConfigType
	Summary *types.SuiteSummary
	BeforeSuites map[string]*types.SetupSummary
	AfterSuites map[string]*types.SetupSummary
	RunningSpecs map[string]*types.SpecSummary
	CompletedSpecs map[string]*types.SpecSummary
}

func (w *wwwReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
	w.Lock()
	defer w.Unlock()
	w.Config = config
	w.Summary = summary
}

func (w *wwwReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {
	w.Lock()
	defer w.Unlock()
	w.BeforeSuites[setupSummary.SuiteID] = setupSummary
}

func (w *wwwReporter) SpecWillRun(specSummary *types.SpecSummary) {
	w.Lock()
	defer w.Unlock()
	testName := strings.Join(specSummary.ComponentTexts[1:], " ")
	w.RunningSpecs[testName] = specSummary
}

func (w *wwwReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	w.Lock()
	defer w.Unlock()
	testName := strings.Join(specSummary.ComponentTexts[1:], " ")
	w.CompletedSpecs[testName] = specSummary	
}

func (w *wwwReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {
	w.Lock()
	defer w.Unlock()
	w.AfterSuites[setupSummary.SuiteID] = setupSummary
}

func (w *wwwReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	
}

func (w *wwwReporter) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	w.Lock()
	defer w.Unlock()
	data, err := json.Marshal(w)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write(data)
}

func NewWWWReporter() *wwwReporter {
	w := &wwwReporter{
		BeforeSuites: map[string]*types.SetupSummary{},
		AfterSuites: map[string]*types.SetupSummary{},
		RunningSpecs: map[string]*types.SpecSummary{},
		CompletedSpecs: map[string]*types.SpecSummary{},
	}
	return w
}