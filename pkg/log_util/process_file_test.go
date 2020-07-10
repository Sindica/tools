package log_util

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_GetTimeFromLog(t *testing.T) {
	input1 := "I0709 01:24:21.904119       1 scheduling_queue.go:817] About to try and schedule pod system/4m3obq-testns/saturation-deployment-0-5c568bc7fc-wgt2l"
	resultTime, err := GetTimeFromLog(input1)
	assert.Nil(t, err)
	assert.Equal(t, "01:24:21.904119", resultTime)

	input2 := "I0709 01:24:22.406258       1 scheduler.go:594] pod system/4m3obq-testns/saturation-deployment-0-5c568bc7fc-dwr6h is bound successfully on node hollow-node-n8jw4, 230 nodes evaluated, 230 nodes were found feasible"
	resultTime, err = GetTimeFromLog(input2)
	assert.Nil(t, err)
	assert.Equal(t, "01:24:22.406258", resultTime)

	input3 := ""
	resultTime, err = GetTimeFromLog(input3)
	assert.NotNil(t, err)
}

func Test_GetTimeDiff(t *testing.T) {
	timeDiff1, err := GetTimeDiff("01:24:21.904119", "01:25:21.904119")
	assert.Nil(t, err)
	assert.Equal(t, time.Duration(1*time.Minute), timeDiff1)

	timeDiff1, err = GetTimeDiff("01:24:21.904119", "01:24:22.904119")
	assert.Nil(t, err)
	assert.Equal(t, time.Duration(1*time.Second), timeDiff1)

	timeDiff1, err = GetTimeDiff("01:24:21.904119", "01:24:21.904120")
	assert.Nil(t, err)
	assert.Equal(t, time.Duration(1*time.Nanosecond), timeDiff1)

	timeDiff1, err = GetTimeDiff("23:59:21.904119", "00:01:21.904119")
	assert.Nil(t, err)
	assert.Equal(t, time.Duration(2*time.Minute), timeDiff1)
}

func Test_parseTime(t *testing.T) {
	h, m, s, ns, err := parseTime("01:24:21.904119")
	assert.Nil(t, err)
	assert.Equal(t, 1, h)
	assert.Equal(t, 24, m)
	assert.Equal(t, 21, s)
	assert.Equal(t, 904119, ns)
}
