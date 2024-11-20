package main

import (
	"errors"
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
		return kmsgparser.Message{}, errors.New("invalid kmsg; must contain a ';'")
	}

	metadata, message := parts[0], parts[1]

	metadataParts := strings.Split(metadata, ",")
	if len(metadataParts) < 3 {
		return kmsgparser.Message{}, errors.New("invalid kmsg: must contain at least 3 ',' separated pieces at the start")
	}

	return kmsgparser.Message{
		Message: message,
	}, nil
}

func getTestData() (klog, podUIDs, containerIDs []string) {
	klog = []string{
		"6,22743,6115623303887,-;oom-kill:constraint=CONSTRAINT_MEMCG,nodemask=(null),cpuset=9f02d9fa0049eb2655fc83c765f142362b2cb403b57b70ba3185071015ca3b64,mems_allowed=0-1,oom_memcg=/kubepods/burstable/podd11ab7b0-d6db-4a24-a7de-4a2faf1e6980/9f02d9fa0049eb2655fc83c765f142362b2cb403b57b70ba3185071015ca3b64,task_memcg=/kubepods/burstable/podd11ab7b0-d6db-4a24-a7de-4a2faf1e6980/9f02d9fa0049eb2655fc83c765f142362b2cb403b57b70ba3185071015ca3b64,task=prometheus-conf,pid=3401999,uid=0",
		"6,23800,6780904484233,-;oom-kill:constraint=CONSTRAINT_MEMCG,nodemask=(null),cpuset=cri-containerd-2260b35b008a15bd118e629c0c5d74e7f3a1fe18c724fbac61a54862fea196dc.scope,mems_allowed=0,oom_memcg=/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-poddfc377c9_c533_4d51_af9e_6e0e0b3db83b.slice,task_memcg=/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-poddfc377c9_c533_4d51_af9e_6e0e0b3db83b.slice/cri-containerd-2260b35b008a15bd118e629c0c5d74e7f3a1fe18c724fbac61a54862fea196dc.scope,task=stress,pid=255629,uid=0",
	}
	podUIDs = []string{
		"d11ab7b0-d6db-4a24-a7de-4a2faf1e6980",
		"dfc377c9_c533_4d51_af9e_6e0e0b3db83b",
	}
	containerIDs = []string{
		"9f02d9fa0049eb2655fc83c765f142362b2cb403b57b70ba3185071015ca3b64",
		"2260b35b008a15bd118e629c0c5d74e7f3a1fe18c724fbac61a54862fea196dc",
	}
	return
}
