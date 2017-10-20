package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/ONSdigital/go-ns/log"
)

// Instance which presents a single dataset being imported
type Instance struct {
	InstanceID           string        `bson:"id,omitempty"                          json:"id,omitempty"`
	CollectionID         string        `bson:"collection_id,omitempty"               json:"collection_id,omitempty"`
	Dimensions           []CodeList    `bson:"dimensions,omitempty"                  json:"dimensions,omitempty"`
	Downloads            *DownloadList `bson:"downloads,omitempty"                   json:"downloads,omitempty"`
	Edition              string        `bson:"edition,omitempty"                     json:"edition,omitempty"`
	Events               *[]Event      `bson:"events,omitempty"                      json:"events,omitempty"`
	Headers              *[]string     `bson:"headers,omitempty"                     json:"headers,omitempty"`
	InsertedObservations *int          `bson:"total_inserted_observations,omitempty" json:"total_inserted_observations,omitempty"`
	License              string        `bson:"license,omitempty"                     json:"license,omitempty"`
	Links                InstanceLinks `bson:"links,omitempty"                       json:"links,omitempty"`
	ReleaseDate          string        `bson:"release_date,omitempty"                json:"release_date,omitempty"`
	State                string        `bson:"state,omitempty"                       json:"state,omitempty"`
	TotalObservations    *int          `bson:"total_observations,omitempty"          json:"total_observations,omitempty"`
	Version              int           `bson:"version,omitempty"                     json:"version,omitempty"`
	LastUpdated          time.Time     `bson:"last_updated,omitempty"                json:"last_updated,omitempty"`
}

// CodeList for a dimension within an instance
type CodeList struct {
	ID   string `json:"id"`
	HRef string `json:"href"`
	Name string `json:"name"`
}

// InstanceLinks holds all links for an instance
type InstanceLinks struct {
	Job        *IDLink `bson:"job,omitempty"        json:"job"`
	Dataset    *IDLink `bson:"dataset,omitempty"    json:"dataset,omitempty"`
	Dimensions *IDLink `bson:"dimensions,omitempty" json:"dimensions,omitempty"`
	Edition    *IDLink `bson:"edition,omitempty"    json:"edition,omitempty"`
	Version    *IDLink `bson:"version,omitempty"    json:"version,omitempty"`
	Self       *IDLink `bson:"self,omitempty"       json:"self,omitempty"`
}

// IDLink holds the id and a link to the resource
type IDLink struct {
	ID   string `bson:"id,omitempty"   json:"id,omitempty"`
	HRef string `bson:"href,omitempty" json:"href,omitempty"`
}

// Event which has happened to an instance
type Event struct {
	Type          string     `bson:"type,omitempty"           json:"type"`
	Time          *time.Time `bson:"time,omitempty"           json:"time"`
	Message       string     `bson:"message,omitempty"        json:"message"`
	MessageOffset string     `bson:"message_offset,omitempty" json:"message_offset"`
}

// InstanceResults wraps instances objects for pagination
type InstanceResults struct {
	Items []Instance `json:"items"`
}

// Validate the event structure
func (e *Event) Validate() error {
	if e.Message == "" || e.MessageOffset == "" || e.Time == nil || e.Type == "" {
		return errors.New("Missing properties")
	}
	return nil
}

var validStates = map[string]int{
	"created":           1,
	"completed":         1,
	"edition-confirmed": 1,
	"associated":        1,
	"published":         1,
}

// ValidateStateFilter checks the list of filter states from a whitelist
func ValidateStateFilter(filterList []string) error {
	var invalidFilterStateValues []string

	for _, filter := range filterList {
		if _, ok := validStates[filter]; !ok {
			invalidFilterStateValues = append(invalidFilterStateValues, filter)
		}
	}

	if invalidFilterStateValues != nil {
		err := fmt.Errorf("invalid filter state values")
		log.Error(err, log.Data{"list-of-invalid-filter-states": invalidFilterStateValues})
		return err
	}

	return nil
}
