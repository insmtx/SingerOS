package nodetools

import (
	"fmt"
	"strings"

	"github.com/insmtx/SingerOS/backend/tools"
)

type assistantNodeInfo struct {
	ContainerID string
}

func nodeInfoForAssistant(toolCtx tools.ToolContext) (*assistantNodeInfo, error) {
	if strings.TrimSpace(toolCtx.AssistantID) == "" {
		return nil, fmt.Errorf("请先与一个已绑定工作节点的 AI 员工对话。")
	}

	// TODO: Resolve the assistant-bound node container from the node manager or database.
	nodeInfo := &assistantNodeInfo{
		ContainerID: "b327e241316c2a2f62cbee986edd0e71235205f0fde5dc7a4543f5344396b351",
	}
	if strings.TrimSpace(nodeInfo.ContainerID) == "" {
		return nil, fmt.Errorf("当前没有工作电脑，无法执行此操作。")
	}

	return nodeInfo, nil
}
