package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/weebank/dio/form"
)

// Workflow
type Workflow struct {
	initialStep string
	steps       map[string]WorkflowStep
	form        *form.Form
}

// Workflow service (holds all workflows and rules)
type WorkflowService struct {
	baseDir   string
	workflows map[string]Workflow

	ticketLifetime time.Duration
}

// Workflow service constructor
func NewService(baseDir string) WorkflowService {
	currDir, _ := os.Getwd()
	return WorkflowService{
		baseDir:   filepath.Join(currDir, baseDir),
		workflows: make(map[string]Workflow),

		ticketLifetime: time.Hour * 12,
	}
}

// Workflow builder (utilitary struct used to build workflows with steps)
type WorkflowBuilder struct {
	service  *WorkflowService
	workflow string
}

// Create new workflow and add it to the service
func (wS *WorkflowService) NewWorkflow(name string) (*WorkflowBuilder, form.FormErrors) {
	// Compile form
	form, err := form.CompileForm(filepath.Join(wS.baseDir, name+".yaml"))
	if err != nil {
		return nil, err
	}

	// Create workflow
	wS.workflows[name] = Workflow{
		steps: make(map[string]WorkflowStep),
		form:  form,
	}

	return &WorkflowBuilder{wS, name}, nil
}

// Workflow step
type WorkflowStep struct {
	onInteract func(responses form.ResponseCollection) (next string)
	onRewind   func() (prev string)
}

// Add step to build workflow
func (w WorkflowBuilder) AddStep(
	name string,
	onInteract func(responses form.ResponseCollection) (next string),
	onRewind func() (prev string),
) error {
	// Check workflow
	workflow, ok := w.service.workflows[w.workflow]
	if !ok {
		return errors.New("workflow does not exist")
	}

	// Check onInteract
	if onInteract == nil {
		return errors.New("onInteract is nil")
	}

	// Add initial step (if needed)
	if workflow.initialStep == "" {
		workflow.initialStep = name
	}

	// Build step
	step := WorkflowStep{
		onInteract: onInteract,
		onRewind:   onRewind,
	}

	// Add step
	workflow.steps[name] = step
	w.service.workflows[w.workflow] = workflow

	return nil
}

// Return all available routes to start workflows
func (wS WorkflowService) Routes() []string {
	routes := make([]string, len(wS.workflows))
	i := 0
	for k := range wS.workflows {
		routes[i] = k
		i++
	}

	return routes
}

// Return a workflow if it exists
func (wS WorkflowService) Workflow(workflow string) (Workflow, bool) {
	wf, ok := wS.workflows[workflow]

	return wf, ok
}
