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
	workflowService *WorkflowService
	workflow        string
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
	Validate func(responses form.ResponseCollection) error
}

// Add step to build workflow
func (w WorkflowBuilder) AddStep(name string, step WorkflowStep) error {
	workflow, ok := w.workflowService.Workflow(w.workflow)
	if !ok {
		return errors.New("workflow does not exist")
	}

	if workflow.initialStep == "" {
		workflow.initialStep = name
	}
	workflow.steps[name] = step
	w.workflowService.workflows[w.workflow] = workflow

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

// Return a workflow and if it exists
func (wS WorkflowService) Workflow(workflow string) (Workflow, bool) {
	wf, ok := wS.workflows[workflow]

	return wf, ok
}
