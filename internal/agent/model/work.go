package model

type Output struct {
	Stream string
	Chunk  string
}
type Result struct {
	Success  bool
	ExitCode int32
	Message  string
}
type Event struct {
	Output *Output
	Result *Result
}
