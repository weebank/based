package workflow

import (
	"errors"
	"time"

	"github.com/weebank/based/form"
)

// Workflow
type Workflow struct {
	initialStep string
	steps       map[string]WorkflowStep
}

// Workflow service (holds all workflows and rules)
type WorkflowService struct {
	baseDir   string
	workflows map[string]Workflow

	ticketLifetime time.Duration
}

// Workflow service constructor
func NewService(baseDir string) WorkflowService {
	return WorkflowService{
		baseDir:   baseDir,
		workflows: make(map[string]Workflow),
	}
}

// Workflow builder (utilitary struct used to build workflows with steps)
type WorkflowBuilder struct {
	service  *WorkflowService
	workflow string
}

// Create new workflow and add it to the service
func (wS *WorkflowService) NewWorkflow(name string) WorkflowBuilder {
	wS.workflows[name] = Workflow{
		steps: make(map[string]WorkflowStep),
	}

	return WorkflowBuilder{wS, name}
}

// Workflow step
type WorkflowStep struct {
	form     *form.Form
	validate func(responses form.ResponseCollection, isFormValid bool) string
}

// Add step to build workflow
func (w WorkflowBuilder) AddStep(name string, hasForm bool, validate func(responses form.ResponseCollection, isFormValid bool) string) error {
	// Check workflow
	workflow, ok := w.service.Workflow(w.workflow)
	if !ok {
		return errors.New("workflow does not exist")
	}

	// Add initial step (if needed)
	if workflow.initialStep == "" {
		workflow.initialStep = name
	}

	// Build step
	step := WorkflowStep{}

	// Compile form (if needed)
	if hasForm {
		var errs form.FormErrors
		if step.form, errs = form.CompileForm(w.service.baseDir); len(errs) > 0 {
			return errs
		}
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
