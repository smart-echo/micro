// Package os runs processes locally
package os

import (
	"github.com/smart-echo/micro/runtime/local/process"
)

type Process struct{}

func NewProcess(opts ...process.Option) process.Process {
	return &Process{}
}
