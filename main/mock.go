package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/weebank/based/form"
	"github.com/weebank/based/workflow"
)

func main() {
	// Create service
	wS := workflow.NewService("/forms")

	// Sign Up
	signUp := wS.NewWorkflow("sign-up")
	signUp.AddStep("welcome",
		workflow.WorkflowStep{
			Validate: func(responses form.ResponseCollection) error {
				return nil
			},
		},
	)

	// Create consumer
	wC := wS.NewConsumer()

	r := gin.Default()

	// Start workflow
	r.POST("/wf/:name",
		func(ctx *gin.Context) {
			if ticket, err := wC.Start(ctx.Param("name")); err == nil {
				ctx.JSON(http.StatusCreated, gin.H{"ticket": ticket})
			} else {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
		},
	)

	// Request object
	type ticketRequest struct {
		Ticket string `json:"ticket"`
	}

	// Get workflow
	r.GET("/wf",
		func(ctx *gin.Context) {
			req := new(ticketRequest)
			err := ctx.Bind(req)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		},
	)

	r.Run()
}
