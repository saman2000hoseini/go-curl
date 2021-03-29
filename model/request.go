package model

type Request struct {
	Duration int64
	FilePath string
	FormData []string
	Headers  []string
	JsonData string
	Method   string
	Queries  []string
}
