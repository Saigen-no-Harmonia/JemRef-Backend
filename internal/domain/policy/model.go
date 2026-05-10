package policy

import "time"

type Policy struct {
	Id          string
	Version     string
	Name        string
	Content     string
	EffectiveDt time.Time
}
