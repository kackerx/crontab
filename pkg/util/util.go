package util

import (
	"encoding/json"
	"github.com/kackerx/crontab/pkg/protocol"
)

// 应答方法
func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	// 1, 定义response
	var (
		response protocol.Response
	)
	response.Data = data
	response.Msg = msg
	response.Errno = errno

	// 2, 序列化
	resp, err = json.Marshal(response)
	return
}

func UnPackJob(value []byte) (*protocol.Job, error) {
	var job protocol.Job
	if err := json.Unmarshal(value, &job); err != nil {
		return nil, err
	}
	return &job, nil
}
