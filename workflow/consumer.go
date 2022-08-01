package workflow

import (
	"errors"
	"fmt"
	"time"

	"github.com/weebank/dio/form"
)

// Workflow consumer (holds all instances)
type WorkflowConsumer struct {
	service   *WorkflowService
	instances map[string]WorkflowInstance
}

// Workflow instance (progress of each user in a workflow)
type WorkflowInstance struct {
	consumer        *WorkflowConsumer
	lastInteraction time.Time
	workflow        string
	step            string
	responsesMap    map[string]form.ResponseCollection
}

// Last interaction getter
func (wI WorkflowInstance) LastInteraction() time.Time {
	return wI.lastInteraction
}

// Update last interaction
func (wI *WorkflowInstance) Refresh() {
	wI.lastInteraction = time.Now()
}

// Get user responses for current step
func (wI WorkflowInstance) Responses() form.ResponseCollection {
	if responses, ok := wI.responsesMap[wI.step]; ok {
		return nil
	} else {
		publicResponses := make(form.ResponseCollection)
		for name, response := range responses {
			if !wI.consumer.service.workflows[wI.workflow].form.Steps[wI.step][name].Sensitive {
				publicResponses[name] = response
			}
		}
		return publicResponses
	}
}

// Check if workflow has expired
func (wI WorkflowInstance) HasExpired() bool {
	return time.Since(wI.LastInteraction()) > wI.consumer.service.lifetime
}

// Workflow service constructor
func (wS *WorkflowService) NewConsumer() WorkflowConsumer {
	return WorkflowConsumer{
		service:   wS,
		instances: make(map[string]WorkflowInstance),
	}
}

// Start a workflow
func (wC *WorkflowConsumer) Start(key string, workflow string) (string, error) {
	// Check workflow
	if _, ok := wC.service.workflows[workflow]; !ok {
		return "", errors.New("workflow does not exist")
	}

	// Check if instance has not been created yet
	if _, ok := wC.instances[key]; !ok {
		// Create and add instance
		wC.instances[key] = WorkflowInstance{
			consumer:        wC,
			step:            wC.service.workflows[workflow].initialStep,
			lastInteraction: time.Now(),
			workflow:        workflow,
			responsesMap:    make(map[string]form.ResponseCollection),
		}
	}

	return wC.instances[key].step, nil
}

// Peek form structure of the current step
func (wC WorkflowConsumer) Peek(key string) (form map[string]form.Field, rewindable bool, step string, err error) {
	// Check if key exists
	instance, ok := wC.instances[key]
	if !ok || instance.HasExpired() {
		return nil, false, "", errors.New("workflow has expired or key does not exist")
	}

	return wC.service.workflows[instance.workflow].form.Steps[instance.step],
		wC.service.workflows[instance.workflow].steps[instance.step].onRewind != nil,
		instance.step,
		nil
}

// Get information related to the workflow instance represented by the given key
func (wC WorkflowConsumer) Get(key string) (responses form.ResponseCollection, err error) {
	// Check if key exists
	instance, ok := wC.instances[key]
	if !ok || instance.HasExpired() {
		return nil, errors.New("workflow has expired or key does not exist")
	}

	// Update timestamp
	instance.Refresh()
	wC.instances[key] = instance

	// Return responses
	if responses := instance.Responses(); responses == nil {
		return make(form.ResponseCollection), nil
	} else {
		return responses, nil
	}
}

// Send responses to workflow
func (wC WorkflowConsumer) Interact(key string, responses form.ResponseCollection) (finished bool, err error) {
	// Check if key exists
	instance, ok := wC.instances[key]
	if !ok || instance.HasExpired() {
		return false, errors.New("workflow has expired or key does not exist")
	}

	// Check if workflow has been finished
	if instance.step == "" {
		return true, nil
	}

	// Update timestamp
	instance.Refresh()
	wC.instances[key] = instance

	// Check step
	var step WorkflowStep
	if step, ok = wC.service.workflows[instance.workflow].steps[instance.step]; !ok {
		return false, fmt.Errorf("step %s does not exist", instance.step)
	}

	// Sanitize response according to step
	form.SanitizeResponse(wC.service.workflows[instance.workflow].form, instance.step, &responses)
	// Validate responses according to form fields
	respErr := form.ValidateResponse(wC.service.workflows[instance.workflow].form, instance.step, responses)
	if respErr != nil {
		return false, respErr
	}

	// Add responses to instance
	instance.responsesMap[instance.step] = responses

	// Trigger onInteract handler
	instance.step = step.onInteract(responses)
	wC.instances[key] = instance

	return instance.step == "", nil
}

// Send responses to workflow
func (wC WorkflowConsumer) Rewind(key string) error {
	// Check if key exists
	instance, ok := wC.instances[key]
	if !ok || instance.HasExpired() {
		return errors.New("workflow has expired or key does not exist")
	}

	// Update timestamp
	instance.Refresh()
	wC.instances[key] = instance

	// Check step
	var step WorkflowStep
	if step, ok = wC.service.workflows[instance.workflow].steps[instance.step]; !ok {
		return fmt.Errorf("step %s does not exist", instance.step)
	}

	// Check if step is rewindable
	if step.onRewind == nil {
		return errors.New("step is not rewindable")
	}

	// Trigger onRewind handler
	instance.step = step.onRewind()
	wC.instances[key] = instance

	return nil
}
