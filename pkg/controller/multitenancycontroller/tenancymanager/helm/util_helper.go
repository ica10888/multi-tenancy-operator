package helm

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"time"
)

type Options struct {
	ReleaseOptions ReleaseOptions
	KubeVersion    string
	APIVersions    []string
}

// ReleaseOptions represents the additional release options needed
// for the composition of the final values struct
type ReleaseOptions struct {
	Name      string
	Time      *timestamp.Timestamp
	Namespace string
	IsUpgrade bool
	IsInstall bool
	Revision  int
}



func Now() *timestamp.Timestamp {
	return Timestamp(time.Now())
}

func Timestamp(t time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}