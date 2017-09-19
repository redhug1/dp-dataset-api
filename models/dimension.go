package models

import "time"

// DatasetDimensionResults represents a structure for a list of dimensions
type DatasetDimensionResults struct {
	Items []Dimension `json:"items"`
}

// DimensionOptionResults represents a structure for a list of dimension options
type DimensionOptionResults struct {
	Items []PublicDimensionOption `json:"items"`
}

// Dimension represents an overview for a single dimension. This includes a link to the code list API
// which provides metadata about the dimension and all possible values.
type Dimension struct {
	Links       DimensionLink `bson:"links,omitempty"         json:"links,omitempty"`
	Name        string        `bson:"name,omitempty"          json:"dimension_id,omitempty"`
	LastUpdated time.Time     `bson:"last_updated,omitempty"  json:"-"`
}

// DimensionLink contains all links needed for a dimension
type DimensionLink struct {
	CodeList LinkObject `bson:"code_list,omitempty"     json:"code_list,omitempty"`
	Options  LinkObject `bson:"options,omitempty"       json:"options,omitempty"`
	Version  LinkObject `bson:"version,omitempty"       json:"version,omitempty"`
}

//
type CachedDimensionOption struct {
	Name       string `bson:"name,omitempty"           json:"dimension_id"`
	Code       string `bson:"code,omitempty"           json:"code"`
	NodeID     string `bson:"node_id,omitempty"        json:"node_id"`
	InstanceID string `bson:"instance_id,omitempty"    json:"instance_id,omitempty"`
	CodeList   string `bson:"code_list,omitempty"      json:"code_list,omitempty"`
	Value      string `bson:"value,omitempty"          json:"value"`
}

// DimensionOption
type DimensionOption struct {
	Name        string               `bson:"name,omitempty"           json:"dimension_id"`
	Label       string               `bson:"label,omitempty"          json:"label"`
	Links       DimensionOptionLinks `bson:"links,omitempty"          json:"links"`
	Value       string               `bson:"value,omitempty"          json:"value"`
	NodeID      string               `bson:"node_id,omitempty"        json:"node_id"`
	InstanceID  string               `bson:"instance_id,omitempty"    json:"instance_id,omitempty"`
	LastUpdated time.Time            `bson:"last_updated,omitempty"    json:"-"`
}

// DimensionOption
type PublicDimensionOption struct {
	Name  string               `bson:"name,omitempty"           json:"dimension_id"`
	Label string               `bson:"label,omitempty"          json:"label"`
	Links DimensionOptionLinks `bson:"links,omitempty"          json:"links"`
	Value string               `bson:"value,omitempty"          json:"value"`
}

type DimensionOptionLinks struct {
	Code     LinkObject `bson:"code,omitempty"              json:"code"`
	Version  LinkObject `bson:"version,omitempty"           json:"version"`
	CodeList LinkObject `bson:"code_list,omitempty"         json:"code_list"`
}

// DimensionNodeResults wraps dimension node objects for pagination
type DimensionNodeResults struct {
	Items []DimensionOption `json:"items"`
}

// DimensionValues holds all unique values for a dimension
type DimensionValues struct {
	Name   string   `json:"dimension_id"`
	Values []string `json:"values"`
}
