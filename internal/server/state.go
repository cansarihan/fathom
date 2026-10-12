package server

import (
	"time"

	"github.com/cansarihan/fathom/internal/campaign"
	"github.com/cansarihan/fathom/internal/crash"
	"github.com/cansarihan/fathom/internal/discover"
	"github.com/cansarihan/fathom/internal/progress"
)

const maxRateSamples = 240

func (state *campaignState) snapshot() Campaign {
	state.mu.Lock()
	defer state.mu.Unlock()
	record := state.record
	record.Targets = append([]TargetRun(nil), state.record.Targets...)
	record.Crashes = append([]crash.Crash(nil), state.record.Crashes...)
	record.Rate = append([]float64(nil), state.record.Rate...)
	return record
}

func (state *campaignState) initTargets(targets []discover.Target) {
	state.mu.Lock()
	defer state.mu.Unlock()
	for _, t := range targets {
		state.record.Targets = append(state.record.Targets, TargetRun{Target: t.Name, Status: "pending"})
	}
}

func (state *campaignState) setStatus(target, status string) {
	state.mu.Lock()
	defer state.mu.Unlock()
	if t := state.targetLocked(target); t != nil {
		t.Status = status
	}
}

func (state *campaignState) apply(target string, u campaign.Update) {
	state.mu.Lock()
	defer state.mu.Unlock()
	t := state.targetLocked(target)
	if t == nil {
		return
	}
	switch u.Event.Kind {
	case progress.KindProgress:
		t.Execs = u.Event.Execs
		t.ExecsPerSec = u.Event.ExecsPerSec
		t.Corpus = u.Event.CorpusTotal
		t.Interesting = u.Event.Interesting
		if len(state.record.Rate) < maxRateSamples {
			state.record.Rate = append(state.record.Rate, float64(u.Event.ExecsPerSec))
		} else {
			state.record.Rate = append(state.record.Rate[1:], float64(u.Event.ExecsPerSec))
		}
	case progress.KindCrash:
		t.Crashed = true
		t.Status = "crash"
	}
}

func (state *campaignState) setResult(target string, result campaign.Result) {
	state.mu.Lock()
	defer state.mu.Unlock()
	t := state.targetLocked(target)
	if t == nil {
		return
	}
	t.Execs = result.Execs
	t.ExecsPerSec = result.ExecsPerSec
	t.Corpus = result.Corpus
	t.Interesting = result.Interesting
	t.Crashed = result.Crashed
	if result.Crashed {
		t.Status = "crash"
	} else {
		t.Status = "done"
	}
	state.record.TotalExecs += result.Execs
}

func (state *campaignState) finish(crashes []crash.Crash) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.record.Crashes = crashes
	state.record.Status = "done"
	state.record.Finished = time.Now()
}

func (state *campaignState) fail(message string) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.record.Status = "error"
	state.record.Err = message
	state.record.Finished = time.Now()
}

func (state *campaignState) targetLocked(name string) *TargetRun {
	for i := range state.record.Targets {
		if state.record.Targets[i].Target == name {
			return &state.record.Targets[i]
		}
	}
	return nil
}
