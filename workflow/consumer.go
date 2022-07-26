package workflow

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/weebank/based/form"
)

// Workflow consumer (holds all instances)
type WorkflowConsumer struct {
	service   *WorkflowService
	instances map[uuid.UUID]WorkflowInstance
}

// Workflow instance (progress of each user in a workflow)
type WorkflowInstance struct {
	consumer        *WorkflowConsumer
	lastInteraction time.Time
	workflow        string
	step            string
	responsesMap    map[string]form.ResponseCollection
}

func (wI WorkflowInstance) LastInteraction() time.Time {
	return wI.lastInteraction
}

func (wI WorkflowInstance) Responses(step string) form.ResponseCollection {
	return wI.responsesMap[step]
}

func (wI WorkflowInstance) HasExpired() bool {
	return time.Since(wI.lastInteraction) > wI.consumer.service.ticketLifetime
}

// Workflow service constructor
func (wS *WorkflowService) NewConsumer() WorkflowConsumer {
	return WorkflowConsumer{
		service:   wS,
		instances: make(map[uuid.UUID]WorkflowInstance),
	}
}

// Start a workflow
func (wC *WorkflowConsumer) Start(workflow string) (string, error) {
	if _, ok := wC.service.workflows[workflow]; !ok {
		return uuid.Nil.String(), errors.New("workflow does not exist")
	}

	id := uuid.New()
	wC.instances[id] = WorkflowInstance{
		consumer:        wC,
		step:            wC.service.workflows[workflow].initialStep,
		lastInteraction: time.Now(),
		workflow:        workflow,
		responsesMap:    make(map[string]form.ResponseCollection),
	}

	return id.String(), nil
}

// Peek form structure of the current step
func (wC WorkflowConsumer) Peek(ticket string) (map[string]form.Rule, error) {
	id, err := uuid.Parse(ticket)
	if err != nil {
		return nil, errors.New("ticket is not a valid uuid")
	}

	instance, ok := wC.instances[id]
	if !ok || instance.HasExpired() {
		return nil, errors.New("ticket has expired or does not exist")
	}

	return wC.service.workflows[instance.workflow].form.Steps[instance.step], nil
}

// Get information related to the workflow instance represented by the given ticket
func (wC WorkflowConsumer) Get(ticket string) (form.ResponseCollection, error) {
	id, err := uuid.Parse(ticket)
	if err != nil {
		return nil, errors.New("ticket is not a valid uuid")
	}

	instance, ok := wC.instances[id]
	if !ok || instance.HasExpired() {
		return nil, errors.New("ticket has expired or does not exist")
	}

	instance.lastInteraction = time.Now()
	wC.instances[id] = instance

	if res, ok := instance.responsesMap[instance.step]; !ok {
		return nil, fmt.Errorf("instance has no responses for the current step (%s)", instance.step)
	} else {
		return res, nil
	}
}

// Send responses to workflow
func (wC WorkflowConsumer) Interact(ticket string, responses form.ResponseCollection) (finished bool, err error) {
	id, err := uuid.Parse(ticket)
	if err != nil {
		return false, errors.New("ticket is not a valid uuid")
	}

	instance, ok := wC.instances[id]
	if !ok || instance.HasExpired() {
		return false, errors.New("ticket has expired or does not exist")
	}

	instance.lastInteraction = time.Now()
	instance.responsesMap[instance.step] = responses

	respErr := form.ValidateResponse(wC.service.workflows[instance.workflow].form, instance.step, responses)
	if respErr != nil {
		return false, respErr
	}

	var step WorkflowStep
	if step, ok = wC.service.workflows[instance.workflow].steps[instance.step]; !ok {
		return false, fmt.Errorf("step %s does not exist", instance.step)
	}

	instance.step = step.onInteract(responses)
	if instance.step == "" {
		delete(wC.instances, id)
		return true, nil
	} else {
		wC.instances[id] = instance
		return false, nil
	}
}
