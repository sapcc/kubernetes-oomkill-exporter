package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/euank/go-kmsg-parser/kmsgparser"
	"github.com/stretchr/testify/require"
)

func TestGetPodUIDFromLog(t *testing.T) {
	klog, podUIDs, containerIDs := getTestData()

	var extractedContainerIDs []string
	var extractedPodUIDs []string

	for _, msg := range klog {
		parsedMsg, err := parseMessage(msg)
		require.NoError(t, err, "There should be no error while parsing kernel log")
		uid, cid := getContainerIDFromLog(parsedMsg.Message)
		fmt.Println(uid)
		extractedContainerIDs = append(extractedContainerIDs, cid)
		extractedPodUIDs = append(extractedPodUIDs, uid)
	}

	require.Equal(t, containerIDs, extractedContainerIDs, "Extracted container ids do not match the expected result")
	require.Equal(t, podUIDs, extractedPodUIDs, "Extracted container ids do not match the expected result")
}

func parseMessage(input string) (kmsgparser.Message, error) {
	// Format:
	//   PRIORITY,SEQUENCE_NUM,TIMESTAMP,-;MESSAGE
	parts := strings.SplitN(input, ";", 2)
	if len(parts) != 2 {
		return kmsgparser.Message{}, fmt.Errorf("invalid kmsg; must contain a ';'")
	}

	metadata, message := parts[0], parts[1]

	metadataParts := strings.Split(metadata, ",")
	if len(metadataParts) < 3 {
		return kmsgparser.Message{}, fmt.Errorf("invalid kmsg: must contain at least 3 ',' separated pieces at the start")
	}

	return kmsgparser.Message{
		Message: message,
	}, nil
}

func getTestData() ([]string, []string, []string) {
	return []string{
			"6,22743,6115623303887,-;oom-kill:constraint=CONSTRAINT_MEMCG,nodemask=(null),cpuset=9f02d9fa0049eb2655fc83c765f142362b2cb403b57b70ba3185071015ca3b64,mems_allowed=0-1,oom_memcg=/kubepods/burstable/podd11ab7b0-d6db-4a24-a7de-4a2faf1e6980/9f02d9fa0049eb2655fc83c765f142362b2cb403b57b70ba3185071015ca3b64,task_memcg=/kubepods/burstable/podd11ab7b0-d6db-4a24-a7de-4a2faf1e6980/9f02d9fa0049eb2655fc83c765f142362b2cb403b57b70ba3185071015ca3b64,task=prometheus-conf,pid=3401999,uid=0",
			"6,23800,6780904484233,-;oom-kill:constraint=CONSTRAINT_MEMCG,nodemask=(null),cpuset=docker-2260b35b008a15bd118e629c0c5d74e7f3a1fe18c724fbac61a54862fea196dc.scope,mems_allowed=0,oom_memcg=/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-poddfc377c9_c533_4d51_af9e_6e0e0b3db83b.slice,task_memcg=/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-poddfc377c9_c533_4d51_af9e_6e0e0b3db83b.slice/docker-2260b35b008a15bd118e629c0c5d74e7f3a1fe18c724fbac61a54862fea196dc.scope,task=stress,pid=255629,uid=0",
			"6,23800,6780904484233,-;oom-kill:constraint=CONSTRAINT_MEMCG,nodemask=(null),cpuset=8322349cef4bf170dd786ad06ddf91c3871c4c097d335eaa1a5a588edfcfeb67,mems_allowed=0,oom_memcg=/kubepods/burstable/poda545513b-a800-49da-8c45-915a953b7e78,task_memcg=/kubepods/burstable/poda545513b-a800-49da-8c45-915a953b7e78/8322349cef4bf170dd786ad06ddf91c3871c4c097d335eaa1a5a588edfcfeb67,task=perl,pid=7173,uid=0",
			"6,23800,6780904484233,-;oom-kill:constraint=CONSTRAINT_MEMCG,nodemask=(null),cpuset=a33e224c8e2f6ed2cd43ffdd03ab1aad260261ea978acf210fb9be342c0a7c24,mems_allowed=0,oom_memcg=/kubepods/podce695e83-cccc-4fb7-a7a1-1c6dbe0f60df,task_memcg=/kubepods/podce695e83-cccc-4fb7-a7a1-1c6dbe0f60df/a33e224c8e2f6ed2cd43ffdd03ab1aad260261ea978acf210fb9be342c0a7c24,task=perl,pid=31097,uid=0",
			"6,23800,6780904484233,-;oom-kill:constraint=CONSTRAINT_MEMCG,nodemask=(null),cpuset=cri-containerd-bd65f4bcf85453bfede70a93c056441aa857ad145e2948d6ea0be1d6b4fbe7d2.scope,mems_allowed=0,oom_memcg=/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod2646e12c_77c3_4d44_a574_6ab00819cb32.slice,task_memcg=/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod2646e12c_77c3_4d44_a574_6ab00819cb32.slice/cri-containerd-bd65f4bcf85453bfede70a93c056441aa857ad145e2948d6ea0be1d6b4fbe7d2.scope,task=perl,pid=1400,uid=0",
			"6,23800,6780904484233,-;oom-kill:constraint=CONSTRAINT_MEMCG,nodemask=(null),cpuset=cri-containerd-3fe4c92b3859cfd884d3d336561b813533219cd0cf8ed4b525ab38e9c17d43b9.scope,mems_allowed=0,oom_memcg=/kubepods.slice/kubepods-pod0286b1b7_3c5a_45da_b2f0_1018921d9229.slice,task_memcg=/kubepods.slice/kubepods-pod0286b1b7_3c5a_45da_b2f0_1018921d9229.slice/cri-containerd-3fe4c92b3859cfd884d3d336561b813533219cd0cf8ed4b525ab38e9c17d43b9.scope,task=perl,pid=31717,uid=0",
		},
		[]string{
			"d11ab7b0-d6db-4a24-a7de-4a2faf1e6980",
			"dfc377c9_c533_4d51_af9e_6e0e0b3db83b",
			"a545513b-a800-49da-8c45-915a953b7e78",
			"ce695e83-cccc-4fb7-a7a1-1c6dbe0f60df",
			"2646e12c_77c3_4d44_a574_6ab00819cb32",
			"0286b1b7_3c5a_45da_b2f0_1018921d9229",
		},
		[]string{
			"9f02d9fa0049eb2655fc83c765f142362b2cb403b57b70ba3185071015ca3b64",
			"2260b35b008a15bd118e629c0c5d74e7f3a1fe18c724fbac61a54862fea196dc",
			"8322349cef4bf170dd786ad06ddf91c3871c4c097d335eaa1a5a588edfcfeb67",
			"a33e224c8e2f6ed2cd43ffdd03ab1aad260261ea978acf210fb9be342c0a7c24",
			"bd65f4bcf85453bfede70a93c056441aa857ad145e2948d6ea0be1d6b4fbe7d2",
			"3fe4c92b3859cfd884d3d336561b813533219cd0cf8ed4b525ab38e9c17d43b9",
		}
}
