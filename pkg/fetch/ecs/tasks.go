package ecs

import (
	"encoding/json"
	"fmt"

	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

func UnmarshalTask(eventDetail json.RawMessage) (*ecsTypes.Task, error) {
	task := new(ecsTypes.Task)
	err := json.Unmarshal(eventDetail, task)
	if err != nil {
		return nil, fmt.Errorf("unmarshal event Detail into task: %w", err)
	}
	return task, nil
}
