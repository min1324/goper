package goper

type poolErrType int

const (
	nameExistsErr poolErrType = iota
	nameNotExistsErr

	errHdNotSet
	errNotRun
	errSafeCall
	errChanClose

	errWorkerFunc
	errInValidTask
)

type poolError struct {
	name   string
	reason poolErrType
	value  interface{}
}

func (pe poolError) Error() string {
	switch pe.reason {
	case nameExistsErr:
		return "GoPool " + pe.name + " exists."

	case nameNotExistsErr:
		return "GoPool " + pe.name + " not exists."

	case errHdNotSet:
		return "GoPool " + pe.name + " Handler func not set."

	case errNotRun:
		return "GoPool " + pe.name + " not run."

	case errSafeCall:
		return "[WARN]GoPool " + pe.name + " :"

	case errChanClose:
		return "GoPool " + pe.name + " chan closed."

	case errWorkerFunc:
		return "arg " + pe.name + " not WorkerFunc."

	case errInValidTask:
		return "Invalid task."
	}
	return ""
}
