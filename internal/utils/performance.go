package utils

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// StepTiming holds timing information for a single step
type StepTiming struct {
	Name      string
	StartTime time.Time
	Duration  time.Duration
	SubSteps  []*StepTiming
}

// StepAggregate holds aggregate timing information for a step
type StepAggregate struct {
	Count    int
	Total    time.Duration
	Average  time.Duration
	Min      time.Duration
	Max      time.Duration
	StepName string
}

// PerformanceTracker tracks execution times of different steps
type PerformanceTracker struct {
	currentStep *StepTiming
	steps       []*StepTiming
	aggregates  map[string]*StepAggregate
	mu          sync.Mutex
}

func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{
		steps:      make([]*StepTiming, 0),
		aggregates: make(map[string]*StepAggregate),
	}
}

// StartStep begins timing a new step
func (pt *PerformanceTracker) StartStep(name string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	step := &StepTiming{
		Name:      name,
		StartTime: time.Now(),
	}

	if pt.currentStep != nil {
		pt.currentStep.SubSteps = append(pt.currentStep.SubSteps, step)
	} else {
		pt.steps = append(pt.steps, step)
	}
	pt.currentStep = step
}

// EndStep completes timing for the current step
func (pt *PerformanceTracker) EndStep() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if pt.currentStep != nil {
		pt.currentStep.Duration = time.Since(pt.currentStep.StartTime)
		pt.updateAggregates(pt.currentStep)

		// Move back to parent step if exists
		found := false
		for _, step := range pt.steps {
			if found = pt.findParentStep(step, pt.currentStep); found {
				break
			}
		}
		if !found {
			pt.currentStep = nil
		}
	}
}

// findParentStep recursively finds the parent of a step
func (pt *PerformanceTracker) findParentStep(current *StepTiming, target *StepTiming) bool {
	for _, subStep := range current.SubSteps {
		if subStep == target {
			pt.currentStep = current
			return true
		}
		if pt.findParentStep(subStep, target) {
			return true
		}
	}
	return false
}

// GenerateReport creates a formatted performance report
func (pt *PerformanceTracker) GenerateReport() string {
	var sb strings.Builder
	sb.WriteString("\n=== Performance Report ===\n")

	for _, step := range pt.steps {
		pt.writeStepReport(&sb, step, 0)
	}

	return sb.String()
}

func (pt *PerformanceTracker) writeStepReport(sb *strings.Builder, step *StepTiming, level int) {
	indent := strings.Repeat("  ", level)
	sb.WriteString(fmt.Sprintf("%s%s: %v\n", indent, step.Name, step.Duration.Round(time.Millisecond)))

	for _, subStep := range step.SubSteps {
		pt.writeStepReport(sb, subStep, level+1)
	}
}

// updateAggregates updates aggregate timing information for a step
func (pt *PerformanceTracker) updateAggregates(step *StepTiming) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	agg, exists := pt.aggregates[step.Name]
	if !exists {
		agg = &StepAggregate{
			StepName: step.Name,
			Min:      step.Duration,
			Max:      step.Duration,
		}
		pt.aggregates[step.Name] = agg
	}

	agg.Count++
	agg.Total += step.Duration
	agg.Average = agg.Total / time.Duration(agg.Count)

	if step.Duration < agg.Min {
		agg.Min = step.Duration
	}
	if step.Duration > agg.Max {
		agg.Max = step.Duration
	}

	// Recursively update aggregates for substeps
	for _, subStep := range step.SubSteps {
		pt.updateAggregates(subStep)
	}
}

// GenerateAggregateReport generates an aggregate performance report
func (pt *PerformanceTracker) GenerateAggregateReport() string {
	var sb strings.Builder
	sb.WriteString("\n=== Aggregate Performance Report ===\n")

	// Sort steps by total time
	var steps []*StepAggregate
	for _, agg := range pt.aggregates {
		steps = append(steps, agg)
	}
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Total > steps[j].Total
	})

	// Write sorted aggregates
	for _, agg := range steps {
		sb.WriteString(fmt.Sprintf(
			"Step: %s\n"+
				"  Count:   %d\n"+
				"  Total:   %v\n"+
				"  Average: %v\n"+
				"  Min:     %v\n"+
				"  Max:     %v\n",
			agg.StepName,
			agg.Count,
			agg.Total.Round(time.Millisecond),
			agg.Average.Round(time.Millisecond),
			agg.Min.Round(time.Millisecond),
			agg.Max.Round(time.Millisecond),
		))
	}

	return sb.String()
}
