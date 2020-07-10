package log_processor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getPodFullNameFromBoundLog(t *testing.T) {
	inputLine := "I0709 01:24:22.406258       1 scheduler.go:594] pod system/4m3obq-testns/saturation-deployment-0-5c568bc7fc-dwr6h is bound successfully on node hollow-node-n8jw4, 230 nodes evaluated, 230 nodes were found feasible"
	podName := getPodFullNameFromBoundLog(inputLine)
	assert.Equal(t, "system/4m3obq-testns/saturation-deployment-0-5c568bc7fc-dwr6h", podName)
}

func Test_getPodFullNameFromTryScheduleLog(t *testing.T) {
	inputLine := "I0709 01:24:21.904119       1 scheduling_queue.go:817] About to try and schedule pod system/4m3obq-testns/saturation-deployment-0-5c568bc7fc-wgt2l"
	podName := getPodFullNameFromTryScheduleLog(inputLine)
	assert.Equal(t, "system/4m3obq-testns/saturation-deployment-0-5c568bc7fc-wgt2l", podName)
}
