package etcd_log

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getValue(t *testing.T) {
	// "key:\"/registry/configmaps\"
	input := "\"key:\\\"/registry/configmaps\\\""
	output, err := getKey(input)
	assert.Nil(t, err)
	assert.Equal(t, "/registry/configmaps", output)

	// "key:\"/registry/clusterroles/system:aggregate-to-view\"
	input = "\"key:\\\"/registry/clusterroles/system:aggregate-to-view\\\""
	output, err = getKey(input)
	assert.Nil(t, err)
	assert.Equal(t, "/registry/clusterroles/system:aggregate-to-view", output)

	// range_end:\"/registry/configmapt\"
	input = "range_end:\\\"/registry/configmapt\\\""
	output, err = getRangeEnd(input)
	assert.Nil(t, err)
	assert.Equal(t, "/registry/configmapt", output)

	// "range_response_count:0
	input = "\"range_response_count:0"
	output, err = getRangeResponseCount(input)
	assert.Nil(t, err)
	assert.Equal(t, "0", output)

	// size:4
	input = "size:4"
	output, err = getSize(input)
	assert.Nil(t, err)
	assert.Equal(t, "4", output)
}

func Test_getDurationInNano(t *testing.T) {
	// (126.195µs)
	input := "(126.195µs)"
	output, err := getDurationInNano(input)
	assert.Nil(t, err)
	assert.Equal(t, "126195", output)

	// (1.017403ms)
	input = "(1.017403ms)"
	output, err = getDurationInNano(input)
	assert.Nil(t, err)
	assert.Equal(t, "1017403", output)
}
