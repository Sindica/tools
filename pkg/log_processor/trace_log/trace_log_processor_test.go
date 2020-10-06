package trace_log

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ParseStep(t *testing.T) {
	// I1003 02:20:09.944334       1 trace.go:81] Trace[1282699261]: "getClientAndClusterIdFromKey: key=/registry/leases/kube-node-lease/hollow-node-jghlp" (started: 2020-10-03 02:20:09.944256833 +0000 UTC m=+8097.689220689) (total time: 56.192µs):
	step := ParseStep("I1003 02:20:09.944334       1 trace.go:81] Trace[1282699261]: \"getClientAndClusterIdFromKey: key=/registry/leases/kube-node-lease/hollow-node-jghlp\" (started: 2020-10-03 02:20:09.944256833 +0000 UTC m=+8097.689220689) (total time: 56.192µs):")
	assert.True(t, step.isStart)
	assert.False(t, step.isEnd)
	assert.Equal(t, "1282699261", step.traceId)
	assert.Equal(t, "2020-10-03 02:20:09.944256833", step.startTime)
	assert.Equal(t, "56.192µs", step.totalDuration)
	assert.Equal(t, "\"getClientAndClusterIdFromKey: key=/registry/leases/kube-node-lease/hollow-node-jghlp\"", step.stepMessage)

	// kube-apiserver.log-20201002-1601632508.gz:I1002 09:51:44.416187       1 trace.go:81] Trace[452806332]: "*****ETCD3 GetToList: key=/configmaps/test-7tbsnb-18/big-deployment-0, resourceVersion=, limit: 500, continue: " (started: 2020-10-02 09:51:41.129206438 +0000 UTC m=+45.462536308) (total time: 3.28696424s):
	step = ParseStep("kube-apiserver.log-20201002-1601632508.gz:I1002 09:51:44.416187       1 trace.go:81] Trace[452806332]: \"*****ETCD3 GetToList: key=/configmaps/test-7tbsnb-18/big-deployment-0, resourceVersion=, limit: 500, continue: \" (started: 2020-10-02 09:51:41.129206438 +0000 UTC m=+45.462536308) (total time: 3.28696424s):")
	assert.True(t, step.isStart)
	assert.False(t, step.isEnd)
	assert.Equal(t, "452806332", step.traceId)
	assert.Equal(t, "2020-10-02 09:51:41.129206438", step.startTime)
	assert.Equal(t, "3.28696424s", step.totalDuration)
	assert.Equal(t, "\"*****ETCD3 GetToList: key=/configmaps/test-7tbsnb-18/big-deployment-0, resourceVersion=, limit: 500, continue: \"", step.stepMessage)

	// Trace[1282699261]: [56.192µs] [1.714µs] END
	step = ParseStep("Trace[1282699261]: [56.192µs] [1.714µs] END")
	assert.False(t, step.isStart)
	assert.True(t, step.isEnd)
	assert.Equal(t, "1282699261", step.traceId)
	assert.Equal(t, "END", step.stepMessage)
	assert.Equal(t, "56.192µs", step.totalDuration)
	assert.Equal(t, "1.714µs", step.stepDuration)

	// kube-apiserver.log-20201002-1601632508.gz:Trace[1826955112]: [546.880485ms] [546.880485ms] END
	step = ParseStep("kube-apiserver.log-20201002-1601632508.gz:Trace[1826955112]: [546.880485ms] [546.880485ms] END")
	assert.False(t, step.isStart)
	assert.True(t, step.isEnd)
	assert.Equal(t, "1826955112", step.traceId)
	assert.Equal(t, "END", step.stepMessage)
	assert.Equal(t, "546.880485ms", step.totalDuration)
	assert.Equal(t, "546.880485ms", step.stepDuration)

	// Trace[1282699261]: [54.478µs] [54.478µs] Returning from getClientAndClusterIdFromKey
	step = ParseStep("Trace[1282699261]: [54.478µs] [54.478µs] Returning from getClientAndClusterIdFromKey")
	assert.False(t, step.isStart)
	assert.False(t, step.isEnd)
	assert.Equal(t, "1282699261", step.traceId)
	assert.Equal(t, "Returning from getClientAndClusterIdFromKey", step.stepMessage)
	assert.Equal(t, "", step.totalDuration)
	assert.Equal(t, "54.478µs", step.stepDuration)
}

func Test_getTraceId(t *testing.T) {
	traceId, err := getTraceId("kube-apiserver.log-20201002-1601632508.gz:Trace[1826955112]:")
	assert.Nil(t, err)
	assert.Equal(t, "1826955112", traceId)

	traceId, err = getTraceId("Trace[1282699261]:")
	assert.Nil(t, err)
	assert.Equal(t, "1282699261", traceId)
}