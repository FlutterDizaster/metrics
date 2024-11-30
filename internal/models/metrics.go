package models

import "time"

type Metrics struct {
	ID     string
	Name   string
	Status string
	State  string

	Time time.Time

	OnlineCPUs    uint32
	TotalCPUUsage uint64
	PerCPUUsage   []uint64

	MemUsage    uint64
	MaxMemUsage uint64
	MemLimit    uint64
}
