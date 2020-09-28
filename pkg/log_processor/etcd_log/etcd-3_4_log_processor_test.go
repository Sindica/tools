package etcd_log

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getMethod(t *testing.T) {
	// Method:\"PUT\"
	input := "Method:\\\"PUT\\\""
	output, err := getMethod(input)
	assert.Nil(t, err)
	assert.Equal(t, "PUT\\", output)
}

func Test_getMethodAndKey(t *testing.T) {
	// success:<request_put:<key:\"/registry/ranges/serviceips\"
	input := "success:<request_put:<key:\\\"/registry/ranges/serviceips\\\""
	method, key, err := getSuccessMethodAndKey(input)
	assert.Nil(t, err)
	assert.Equal(t, "request_put", method)
	assert.Equal(t, "/registry/ranges/serviceips\\", key)

	// success:<request_put:<key:\"/registry/clusterroles/system:discovery\"
	input = "success:<request_put:<key:\\\"/registry/clusterroles/system:discovery\\\""
	method, key, err = getSuccessMethodAndKey(input)
	assert.Nil(t, err)
	assert.Equal(t, "request_put", method)
	assert.Equal(t, "/registry/clusterroles/system:discovery\\", key)

	// failure:<request_range:<key:\"/registry/ranges/serviceips\"
	input = "failure:<request_range:<key:\\\"/registry/ranges/serviceips\\\""
	method, key, err = getFailureMethodAndKey(input)
	assert.Nil(t, err)
	assert.Equal(t, "request_range", method)
	assert.Equal(t, "/registry/ranges/serviceips\\", key)

	// failure:<>>"
	input = "failure:<>>\""
	method, key, err = getFailureMethodAndKey(input)
	assert.Nil(t, err)
	assert.Equal(t, "", method)
	assert.Equal(t, "", key)
}

func Test_getAction(t *testing.T) {
	// lease_grant:<ttl:15-second id:130174b703f730e4>"
	input := "lease_grant:<ttl:15-second"
	output, err := getAction(input)
	assert.Nil(t, err)
	assert.Equal(t, "lease_grant", output)

	// compaction:<revision:1000
	input = "compaction:<revision:1000"
	output, err = getAction(input)
	assert.Nil(t, err)
	assert.Equal(t, "compaction", output)
}

func Test_getPath(t *testing.T) {
	// Path:\"/0/members/4faa637bfd19301/attributes\"
	input := "Path:\\\"/0/members/4faa637bfd19301/attributes\\\""
	output, err := getPath(input)
	assert.Nil(t, err)
	assert.Equal(t, "/0/members/4faa637bfd19301/attributes\\", output)
}

func Test_getKey(t *testing.T) {
	// "key:\"/registry/configmaps\"
	input := "\"key:\\\"/registry/configmaps\\\""
	output, err := getKey(input)
	assert.Nil(t, err)
	assert.Equal(t, "/registry/configmaps\\", output)

	// "key:\"/registry/clusterroles/system:aggregate-to-view\"
	input = "\"key:\\\"/registry/clusterroles/system:aggregate-to-view\\\""
	output, err = getKey(input)
	assert.Nil(t, err)
	assert.Equal(t, "/registry/clusterroles/system:aggregate-to-view\\", output)

	// key:\"/registry/ranges/serviceips\"
	input = "key:\\\"/registry/ranges/serviceips\\\""
	output, err = getKey(input)
	assert.Nil(t, err)
	assert.Equal(t, "/registry/ranges/serviceips\\", output)

	// txn:<compare:<key:\"compact_rev_key\"
	input = "txn:<compare:<key:\\\"compact_rev_key\\\""
	output, err = getKey(input)
	assert.Nil(t, err)
	assert.Equal(t, "compact_rev_key\\", output)
}

func Test_getValue(t *testing.T) {
	// range_end:\"/registry/configmapt\"
	input := "range_end:\\\"/registry/configmapt\\\""
	output, err := getRangeEnd(input)
	assert.Nil(t, err)
	assert.Equal(t, "/registry/configmapt\\", output)

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

	// "size:14"
	input = "size:14"
	output, err = getSize(input)
	assert.Nil(t, err)
	assert.Equal(t, "14", output)


	// mod_revision:0
	input = "mod_revision:0"
	output, err = getModRevision(input)
	assert.Nil(t, err)
	assert.Equal(t, "0", output)

	// value_size:74
	input = "value_size:74"
	output, err = getValueSize(input)
	assert.Nil(t, err)
	assert.Equal(t, "74", output)
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

func Test_NoReadOnlyRangeRequest(t *testing.T) {
	// etcd.log-20200922-1600800918.gz:2020-09-22 18:21:59.761339 I | etcdserver: request "header:<ID:10592876127946500926 > compaction:<revision:1000 > " with result "size:5" took (2.337296ms) to execute
	line := "etcd.log-20200922-1600800918.gz:2020-09-22 18:21:59.761339 I | etcdserver: request \"header:<ID:10592876127946500926 > compaction:<revision:1000 > \" with result \"size:5\" took (2.337296ms) to execute"
	hasError, _ := NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// k8s 3.4.4
	// etcd.log-20200925-1601065806.gz:2020-09-25 20:25:29.188661 I | etcdserver: request "header:<ID:10636223341819497370 username:\"client\" auth_revision:1 > txn:<compare:<target:MOD key:\"/registry/deployments/test-6vwfq1-1/latency-deployment-0\" mod_revision:156142 > success:<request_delete_range:<key:\"/registry/deployments/test-6vwfq1-1/latency-deployment-0\" > > failure:<request_range:<key:\"/registry/deployments/test-6vwfq1-1/latency-deployment-0\" > >>" with result "size:20" took (211.829µs) to execute
	line = "etcd.log-20200925-1601065806.gz:2020-09-25 20:25:29.188661 I | etcdserver: request \"header:<ID:10636223341819497370 username:\\\"client\\\" auth_revision:1 > txn:<compare:<target:MOD key:\\\"/registry/deployments/test-6vwfq1-1/latency-deployment-0\\\" mod_revision:156142 > success:<request_delete_range:<key:\\\"/registry/deployments/test-6vwfq1-1/latency-deployment-0\\\" > > failure:<request_range:<key:\\\"/registry/deployments/test-6vwfq1-1/latency-deployment-0\\\" > >>\" with result \"size:20\" took (211.829µs) to execute\n"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// k8s 3.4.4
	// etcd.log-20200925-1601065806.gz:2020-09-25 20:20:54.061610 W | etcdserver: request "header:<ID:10636223341819455499 username:\"client\" auth_revision:1 > txn:<compare:<target:MOD key:\"/registry/leases/kube-node-lease/hollow-node-dgqdv\" mod_revision:134112 > success:<request_put:<key:\"/registry/leases/kube-node-lease/hollow-node-dgqdv\" value_size:564 >> failure:<request_range:<key:\"/registry/leases/kube-node-lease/hollow-node-dgqdv\" > >>" with result "size:18" took too long (129.197656ms) to execute
	line = "etcd.log-20200925-1601065806.gz:2020-09-25 20:20:54.061610 W | etcdserver: request \"header:<ID:10636223341819455499 username:\\\"client\\\" auth_revision:1 > txn:<compare:<target:MOD key:\\\"/registry/leases/kube-node-lease/hollow-node-dgqdv\\\" mod_revision:134112 > success:<request_put:<key:\\\"/registry/leases/kube-node-lease/hollow-node-dgqdv\\\" value_size:564 >> failure:<request_range:<key:\\\"/registry/leases/kube-node-lease/hollow-node-dgqdv\\\" > >>\" with result \"size:18\" took too long (129.197656ms) to execute\n"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// k8s 3.4.4
	// etcd.log-20200925-1601065806.gz:2020-09-25 19:29:03.898829 I | etcdserver: request "header:<ID:10636223341819269789 username:\"client\" auth_revision:1 > txn:<compare:<key:\"compact_rev_key\" version:0 > success:<request_put:<key:\"compact_rev_key\" value_size:1 >> failure:<request_range:<key:\"compact_rev_key\" > >>" with result "size:16" took (166.551µs) to execute
	line = "etcd.log-20200925-1601065806.gz:2020-09-25 19:29:03.898829 I | etcdserver: request \"header:<ID:10636223341819269789 username:\\\"client\\\" auth_revision:1 > txn:<compare:<key:\\\"compact_rev_key\\\" version:0 > success:<request_put:<key:\\\"compact_rev_key\\\" value_size:1 >> failure:<request_range:<key:\\\"compact_rev_key\\\" > >>\" with result \"size:16\" took (166.551µs) to execute\n"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// k8s 3.4.4
	// etcd.log:2020-09-25 19:24:09.783370 I | etcdserver: request "header:<ID:10636223341819266001 username:\"client\" auth_revision:1 > txn:<compare:<target:MOD key:\"/registry/masterleases/10.40.0.12\" mod_revision:0 > success:<request_put:<key:\"/registry/masterleases/10.40.0.12\" value_size:65 lease:1412851304964490191 >> failure:<request_range:<key:\"/registry/masterleases/10.40.0.12\" > >>" with result "size:16" took (123.553µs) to execute
	line = "etcd.log:2020-09-25 19:24:09.783370 I | etcdserver: request \"header:<ID:10636223341819266001 username:\\\"client\\\" auth_revision:1 > txn:<compare:<target:MOD key:\\\"/registry/masterleases/10.40.0.12\\\" mod_revision:0 > success:<request_put:<key:\\\"/registry/masterleases/10.40.0.12\\\" value_size:65 lease:1412851304964490191 >> failure:<request_range:<key:\\\"/registry/masterleases/10.40.0.12\\\" > >>\" with result \"size:16\" took (123.553µs) to execute\n"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// k8s 3.4.4
	// etcd.log:2020-09-25 19:24:09.781338 I | etcdserver: request "header:<ID:10636223341819266000 username:\"client\" auth_revision:1 > lease_grant:<ttl:15-second id:139b74c6b8db03cf>" with result "size:40" took (124.69µs) to execute
	line = "etcd.log:2020-09-25 19:24:09.781338 I | etcdserver: request \"header:<ID:10636223341819266000 username:\\\"client\\\" auth_revision:1 > lease_grant:<ttl:15-second id:139b74c6b8db03cf>\" with result \"size:40\" took (124.69µs) to execute\n"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// k8s 3.4.4
	// etcd.log:2020-09-25 19:24:07.878007 I | etcdserver: request "header:<ID:10636223341819265677 username:\"client\" auth_revision:1 > txn:<compare:<target:MOD key:\"/registry/ranges/serviceips\" mod_revision:0 > success:<request_put:<key:\"/registry/ranges/serviceips\" value_size:68 >> failure:<request_range:<key:\"/registry/ranges/serviceips\" > >>" with result "size:14" took (205.359µs) to execute
	line = "etcd.log:2020-09-25 19:24:07.878007 I | etcdserver: request \"header:<ID:10636223341819265677 username:\\\"client\\\" auth_revision:1 > txn:<compare:<target:MOD key:\\\"/registry/ranges/serviceips\\\" mod_revision:0 > success:<request_put:<key:\\\"/registry/ranges/serviceips\\\" value_size:68 >> failure:<request_range:<key:\\\"/registry/ranges/serviceips\\\" > >>\" with result \"size:14\" took (205.359µs) to execute\n"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// k8s 3.4.4:
	// etcd.log:2020-09-25 19:24:07.860705 I | etcdserver: request "header:<ID:10636223341819265670 username:\"client\" auth_revision:1 > txn:<compare:<target:MOD key:\"/registry/apiregistration.k8s.io/apiservices/v1.\" mod_revision:0 > success:<request_put:<key:\"/registry/apiregistration.k8s.io/apiservices/v1.\" value_size:495 >> failure:<>>" with result "size:14" took (305.488µs) to execute
	line = "etcd.log:2020-09-25 19:24:07.860705 I | etcdserver: request \"header:<ID:10636223341819265670 username:\\\"client\\\" auth_revision:1 > txn:<compare:<target:MOD key:\\\"/registry/apiregistration.k8s.io/apiservices/v1.\\\" mod_revision:0 > success:<request_put:<key:\\\"/registry/apiregistration.k8s.io/apiservices/v1.\\\" value_size:495 >> failure:<>>\" with result \"size:14\" took (305.488µs) to execute"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// etcd.log:2020-09-22 18:12:05.129283 I | etcdserver: request "ID:10592876127946485506 Method:\"PUT\" Path:\"/0/members/4faa637bfd19301/attributes\" Val:\"{\\\"name\\\":\\\"etcd-ying3-kubemark-compare-10-kubemark-master\\\",\\\"clientURLs\\\":[\\\"http://127.0.0.1:2379\\\",\\\"https://10.40.0.11:2379\\\"]}\" " with result "" took (113.908µs) to execute
	line = "etcd.log:2020-09-22 18:12:05.129283 I | etcdserver: request \"ID:10592876127946485506 Method:\\\"PUT\\\" Path:\\\"/0/members/4faa637bfd19301/attributes\\\" Val:\\\"{\\\\\\\"name\\\\\\\":\\\\\\\"etcd-ying3-kubemark-compare-10-kubemark-master\\\\\\\",\\\\\\\"clientURLs\\\\\\\":[\\\\\\\"http://127.0.0.1:2379\\\\\\\",\\\\\\\"https://10.40.0.11:2379\\\\\\\"]}\\\" \" with result \"\" took (113.908µs) to execute"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// etcd.log:2020-09-22 18:12:47.108193 I | etcdserver: request "ID:10592876127946486575 Method:\"QGET\" " with result "" took (12.789µs) to execute
	line = "etcd.log:2020-09-22 18:12:47.108193 I | etcdserver: request \"ID:10592876127946486575 Method:\\\"QGET\\\" \" with result \"\" took (12.789µs) to execute"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// etcd.log:2020-09-22 18:12:26.540548 I | etcdserver: request "header:<ID:10592876127946486339 > lease_revoke:<id:130174b703f730e4>" with result "size:27" took (65.923µs) to execute
	line = "etcd.log:2020-09-22 18:12:26.540548 I | etcdserver: request \"header:<ID:10592876127946486339 > lease_revoke:<id:130174b703f730e4>\" with result \"size:27\" took (65.923µs) to execute"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// etcd.log:2020-09-22 18:12:11.177562 I | etcdserver: request "header:<ID:10592876127946485989 > lease_grant:<ttl:15-second id:130174b703f730e4>" with result "size:39" took (149.606µs) to execute
	line = "etcd.log:2020-09-22 18:12:11.177562 I | etcdserver: request \"header:<ID:10592876127946485989 > lease_grant:<ttl:15-second id:130174b703f730e4>\" with result \"size:39\" took (149.606µs) to execute"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// 24: etcd.log:2020-09-22 18:16:59.757018 I | etcdserver: request "header:<ID:10592876127946488796 > txn:<compare:<key:\"compact_rev_key\" version:0 > success:<request_put:<key:\"compact_rev_key\" value_size:1 >> failure:<request_range:<key:\"compact_rev_key\" > >>" with result "size:16" took (151.315µs) to execute
	line = "etcd.log:2020-09-22 18:16:59.757018 I | etcdserver: request \"header:<ID:10592876127946488796 > txn:<compare:<key:\\\"compact_rev_key\\\" version:0 > success:<request_put:<key:\\\"compact_rev_key\\\" value_size:1 >> failure:<request_range:<key:\\\"compact_rev_key\\\" > >>\" with result \"size:16\" took (151.315µs) to execute"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// 26: etcd.log:2020-09-22 18:12:11.180111 I | etcdserver: request "header:<ID:10592876127946485990 > txn:<compare:<target:MOD key:\"/registry/masterleases/10.40.0.11\" mod_revision:0 > success:<request_put:<key:\"/registry/masterleases/10.40.0.11\" value_size:74 lease:1369504091091710180 >> failure:<request_range:<key:\"/registry/masterleases/10.40.0.11\" > >>" with result "size:16" took (1.415049ms) to execute
	line = "etcd.log:2020-09-22 18:12:11.180111 I | etcdserver: request \"header:<ID:10592876127946485990 > txn:<compare:<target:MOD key:\\\"/registry/masterleases/10.40.0.11\\\" mod_revision:0 > success:<request_put:<key:\\\"/registry/masterleases/10.40.0.11\\\" value_size:74 lease:1369504091091710180 >> failure:<request_range:<key:\\\"/registry/masterleases/10.40.0.11\\\" > >>\" with result \"size:16\" took (1.415049ms) to execute"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// 25: etcd.log:2020-09-22 18:12:05.457308 I | etcdserver: request "header:<ID:10592876127946485567 > txn:<compare:<target:MOD key:\"/registry/ranges/serviceips\" mod_revision:0 > success:<request_put:<key:\"/registry/ranges/serviceips\" value_size:74 >> failure:<request_range:<key:\"/registry/ranges/serviceips\" > >>" with result "size:14" took (362.41µs) to execute
	line = "etcd.log:2020-09-22 18:12:05.457308 I | etcdserver: request \"header:<ID:10592876127946485567 > txn:<compare:<target:MOD key:\\\"/registry/ranges/serviceips\\\" mod_revision:0 > success:<request_put:<key:\\\"/registry/ranges/serviceips\\\" value_size:74 >> failure:<request_range:<key:\\\"/registry/ranges/serviceips\\\" > >>\" with result \"size:14\" took (362.41µs) to execute"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// 23: etcd.log:2020-09-22 18:12:08.563601 I | etcdserver: request "header:<ID:10592876127946485651 > txn:<compare:<target:MOD key:\"/registry/namespaces/kube-system\" mod_revision:0 > success:<request_put:<key:\"/registry/namespaces/kube-system\" value_size:146 >> failure:<>>" with result "size:14" took (106.823µs) to execute
	line = "etcd.log:2020-09-22 18:12:05.457308 I | etcdserver: request \"header:<ID:10592876127946485567 > txn:<compare:<target:MOD key:\\\"/registry/ranges/serviceips\\\" mod_revision:0 > success:<request_put:<key:\\\"/registry/ranges/serviceips\\\" value_size:74 >> failure:<request_range:<key:\\\"/registry/ranges/serviceips\\\" > >>\" with result \"size:14\" took (362.41µs) to execute"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// etcd.log-20200922-1600802418.gz:2020-09-22 19:01:19.601077 I | etcdserver: request "header:<ID:10592876127946897136 > txn:<compare:<target:MOD key:\"/registry/deployments/4trjes-testns/latency-deployment-0\" mod_revision:218600 > success:<request_delete_range:<key:\"/registry/deployments/4trjes-testns/latency-deployment-0\" > > failure:<request_range:<key:\"/registry/deployments/4trjes-testns/latency-deployment-0\" > >>" with result "size:20" took (89.111µs) to execute
	line = "etcd.log:2020-09-22 19:01:19.601077 I | etcdserver: request \"header:<ID:10592876127946897136 > txn:<compare:<target:MOD key:\\\"/registry/deployments/4trjes-testns/latency-deployment-0\\\" mod_revision:218600 > success:<request_delete_range:<key:\\\"/registry/deployments/4trjes-testns/latency-deployment-0\\\" > > failure:<request_range:<key:\\\"/registry/deployments/4trjes-testns/latency-deployment-0\\\" > >>\" with result \"size:20\" took (89.111µs) to execute\n"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// etcd.log-20200923-1600895105.gz:2020-09-23 21:00:31.043176 W | etcdserver: request "header:<ID:10592876152216706621 > txn:<compare:<target:MOD key:\"/registry/configmaps/dyac4z-testns/medium-deployment-1\" mod_revision:0 > success:<request_put:<key:\"/registry/configmaps/dyac4z-testns/medium-deployment-1\" value_size:173 >> failure:<>>" with result "size:18" took too long (118.172147ms) to execute
	line = "etcd.log:2020-09-23 21:00:31.043176 W | etcdserver: request \"header:<ID:10592876152216706621 > txn:<compare:<target:MOD key:\\\"/registry/configmaps/dyac4z-testns/medium-deployment-1\\\" mod_revision:0 > success:<request_put:<key:\\\"/registry/configmaps/dyac4z-testns/medium-deployment-1\\\" value_size:173 >> failure:<>>\" with result \"size:18\" took too long (118.172147ms) to execute\n"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)

	// etcd.log-20200923-1600899026.gz:2020-09-23 21:50:05.805743 W | etcdserver: request "header:<ID:10592876152217224613 > txn:<compare:<target:MOD key:\"/registry/minions/hollow-node-hmhjh\" mod_revision:412936 > success:<request_put:<key:\"/registry/minions/hollow-node-hmhjh\" value_size:1390 >> failure:<request_range:<key:\"/registry/minions/hollow-node-hmhjh\" > >>" with result "size:18" took too long (118.23637ms) to execute
	line = "etcd.log-20200923-1600899026.gz:2020-09-23 21:50:05.805743 W | etcdserver: request \"header:<ID:10592876152217224613 > txn:<compare:<target:MOD key:\\\"/registry/minions/hollow-node-hmhjh\\\" mod_revision:412936 > success:<request_put:<key:\\\"/registry/minions/hollow-node-hmhjh\\\" value_size:1390 >> failure:<request_range:<key:\\\"/registry/minions/hollow-node-hmhjh\\\" > >>\" with result \"size:18\" took too long (118.23637ms) to execute\n"
	hasError, _ = NoReadOnlyRangeRequest(line)
	assert.False(t, hasError)
}