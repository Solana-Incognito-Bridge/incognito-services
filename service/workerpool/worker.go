package workerpool

type Worker interface {
	Task()
	TrackHistory(status int, message string, responseData string)
}

