package domain

import "time"

type Person struct {
	OrderNumber    int64      `json:"orderNumber"`
	Name           string     `json:"name"`
	Surname        string     `json:"surname"`
	StudyDirection string     `json:"studyDirection"`
	VisitedEvent   bool       `json:"visitedEvent"`
	CheckInTime    *time.Time `json:"checkInTime,omitempty"`
}

type PeopleFilter struct {
	Visited        *bool
	StudyDirection string
}

type ProgramStat struct {
	StudyDirection string `json:"studyDirection"`
	Count          int64  `json:"count"`
}
