package types

type StringOption struct {
	Value <-chan string
	Error <-chan error
}
