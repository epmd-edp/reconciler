package model

type CodebaseBranchDTO struct {
	CodebaseName string
	BranchName   string
}

type CDPipelineDTO struct {
	Id     int
	Name   string
	Status string
}

type CodebaseDockerStreamReadDTO struct {
	CodebaseDockerStreamId int
	CodebaseId             int
	CodebaseName           string
}

type CodebaseBranchIdDTO struct {
	CodebaseId int
	BranchId   int
}
