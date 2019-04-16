package model

type ApplicationBranchDTO struct {
	AppName    string
	BranchName string
}

type CDPipelineDTO struct {
	Id     int
	Name   string
	Status string
}
