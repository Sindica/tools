package controller_log

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getMatchCase(t *testing.T) {
	/*
	I0409 22:32:35.827427       1 eventhandlers.go:164] Getting pod saturation-deployment-0-c47675f5-xf258 from API server
	I0409 22:32:35.827433       1 scheduling_queue.go:210] adding pod arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258 to the scheduling queue.
	I0409 22:32:36.209513       1 scheduling_queue.go:819] About to try and schedule pod arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258
	I0409 22:32:36.209529       1 scheduler.go:458] Attempting to schedule pod: arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258
	I0409 22:32:36.212748       1 generic_scheduler.go:211] DEBUG: Compute predicates pod: arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258
	I0409 22:32:36.228494       1 generic_scheduler.go:231] DEBUG: Prioritizing pod: arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258
	I0409 22:32:36.245901       1 generic_scheduler.go:255] DEBUG: Selecting host pod: arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258
	I0409 22:32:36.246120       1 scheduler.go:417] Attempting to bind pod: arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258
	I0409 22:32:36.248275       1 scheduler.go:596] pod arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258 is bound successfully on node hollow-node-1-btv5d, 500 nodes evaluated, 500 nodes were found feasible
	*/

	inputLine := "I0409 22:32:35.827427       1 eventhandlers.go:164] Getting pod saturation-deployment-0-c47675f5-xf258 from API server"
	isMatch, caseId, podName := getMatchCase(inputLine)
	assert.True(t, isMatch)
	assert.Equal(t, 0, caseId)
	assert.Equal(t, "saturation-deployment-0-c47675f5-xf258", podName)
}
