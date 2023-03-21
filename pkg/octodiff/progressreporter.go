package octodiff

type ProgressReporter interface {
	ReportProgress(operation string, currentPosition int64, total int64)
}

type nopProgressReporter struct {
}

func (n nopProgressReporter) ReportProgress(operation string, currentPosition int64, total int64) {
	// safely does nothing
}

var nopProgressReporterInstance = &nopProgressReporter{}

func NopProgressReporter() ProgressReporter {
	return nopProgressReporterInstance
}
