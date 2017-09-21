package mongo

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-dataset-api/store"

	errs "github.com/ONSdigital/dp-dataset-api/apierrors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var _ store.Storer = &Mongo{}
var session *mgo.Session

// Mongo represents a simplistic MongoDB configuration.
type Mongo struct {
	Collection string
	Database   string
	URI        string
}

// Init creates a new mgo.Session with a strong consistency and a write mode of "majortiy".
func (m *Mongo) Init() (err error) {
	if session != nil {
		return
	}

	if session, err = mgo.Dial(m.URI); err != nil {
		return
	}

	session.EnsureSafe(&mgo.Safe{WMode: "majority"})
	session.SetMode(mgo.Strong, true)
	return
}

// GetDatasets retrieves all dataset documents
func (m *Mongo) GetDatasets() (*models.DatasetResults, error) {
	s := session.Copy()
	defer s.Close()

	datasets := &models.DatasetResults{}

	iter := s.DB(m.Database).C("datasets").Find(nil).Iter()
	defer iter.Close()

	results := []models.DatasetUpdate{}
	if err := iter.All(&results); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errs.DatasetNotFound
		}
		return nil, err
	}

	datasets.Items = mapResults(results)

	return datasets, nil
}

func mapResults(results []models.DatasetUpdate) []*models.Dataset {
	items := []*models.Dataset{}
	for _, item := range results {
		if item.Current == nil {
			continue
		}

		items = append(items, item.Current)
	}
	return items
}

// GetDataset retrieves a dataset document
func (m *Mongo) GetDataset(id string) (*models.DatasetUpdate, error) {
	s := session.Copy()
	defer s.Clone()
	var dataset models.DatasetUpdate
	err := s.DB(m.Database).C("datasets").Find(bson.M{"_id": id}).One(&dataset)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errs.DatasetNotFound
		}
		return nil, err
	}

	return &dataset, nil
}

// GetEditions retrieves all edition documents for a dataset
func (m *Mongo) GetEditions(id, state string) (*models.EditionResults, error) {
	s := session.Copy()
	defer s.Clone()

	selector := buildEditionsQuery(id, state)

	iter := s.DB(m.Database).C("editions").Find(selector).Iter()
	defer iter.Close()

	var results []models.Edition
	if err := iter.All(&results); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errs.EditionNotFound
		}
		return nil, err
	}

	if len(results) < 1 {
		return nil, errs.EditionNotFound
	}
	return &models.EditionResults{Items: results}, nil
}

func buildEditionsQuery(id, state string) bson.M {
	var selector bson.M
	if state != "" {
		selector = bson.M{
			"links.dataset.id": id,
			"state":            state,
		}
	} else {
		selector = bson.M{
			"links.dataset.id": id,
		}
	}

	return selector
}

// GetEdition retrieves an edition document for a dataset
func (m *Mongo) GetEdition(id, editionID, state string) (*models.Edition, error) {
	s := session.Copy()
	defer s.Clone()

	selector := buildEditionQuery(id, editionID, state)

	var edition models.Edition
	err := s.DB(m.Database).C("editions").Find(selector).One(&edition)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errs.EditionNotFound
		}
		return nil, err
	}
	return &edition, nil
}

func buildEditionQuery(id, editionID, state string) bson.M {
	var selector bson.M
	if state == "" {
		selector = bson.M{
			"links.dataset.id": id,
			"edition":          editionID,
		}
	} else {
		selector = bson.M{
			"links.dataset.id": id,
			"edition":          editionID,
			"state":            state,
		}
	}

	return selector
}

// GetNextVersion retrieves the latest version for an edition of a dataset
func (m *Mongo) GetNextVersion(datasetID, editionID string) (int, error) {
	s := session.Copy()
	defer s.Clone()
	var version models.Version
	var nextVersion int
	err := s.DB(m.Database).C("versions").Find(bson.M{"links.dataset.id": datasetID, "edition": editionID}).Sort("-version").One(&version)
	if err != nil {
		if err == mgo.ErrNotFound {
			return 1, nil
		}
		return nextVersion, err
	}

	nextVersion = version.Version + 1

	return nextVersion, nil
}

// GetVersions retrieves all version documents for a dataset edition
func (m *Mongo) GetVersions(id, editionID, state string) (*models.VersionResults, error) {
	s := session.Copy()
	defer s.Clone()

	selector := buildVersionsQuery(id, editionID, state)

	iter := s.DB(m.Database).C("versions").Find(selector).Iter()
	defer iter.Close()

	var results []models.Version
	if err := iter.All(&results); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errs.VersionNotFound
		}
		return nil, err
	}

	if len(results) < 1 {
		return nil, errs.VersionNotFound
	}

	return &models.VersionResults{Items: results}, nil
}

func buildVersionsQuery(id, editionID, state string) bson.M {
	var selector bson.M
	if state == "" {
		selector = bson.M{
			"links.dataset.id": id,
			"edition":          editionID,
		}
	} else {
		selector = bson.M{
			"links.dataset.id": id,
			"edition":          editionID,
			"state":            state,
		}
	}

	return selector
}

// GetVersion retrieves a version document for a dataset edition
func (m *Mongo) GetVersion(id, editionID, versionID, state string) (*models.Version, error) {
	s := session.Copy()
	defer s.Clone()

	versionNumber, err := strconv.Atoi(versionID)
	if err != nil {
		return nil, err
	}
	selector := buildVersionQuery(id, editionID, state, versionNumber)

	var version models.Version
	err = s.DB(m.Database).C("versions").Find(selector).One(&version)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errs.VersionNotFound
		}
		return nil, err
	}
	return &version, nil
}

func buildVersionQuery(id, editionID, state string, versionID int) bson.M {
	var selector bson.M
	if state != "published" {
		selector = bson.M{
			"links.dataset.id": id,
			"version":          versionID,
			"edition":          editionID,
		}
	} else {
		selector = bson.M{
			"links.dataset.id": id,
			"edition":          editionID,
			"version":          versionID,
			"state":            state,
		}
	}

	return selector
}

// UpdateDataset updates an existing dataset document
func (m *Mongo) UpdateDataset(id string, dataset *models.Dataset) (err error) {
	s := session.Copy()
	defer s.Close()

	updates := createDatasetUpdateQuery(dataset)

	err = s.DB(m.Database).C("datasets").UpdateId(id, bson.M{"$set": updates, "$setOnInsert": bson.M{"next.last_updated": time.Now()}})
	return
}

func createDatasetUpdateQuery(dataset *models.Dataset) bson.M {
	updates := make(bson.M, 0)

	if dataset.CollectionID != "" {
		updates["next.collection_id"] = dataset.CollectionID
	}

	if dataset.Contact.Email != "" {
		updates["next.contact.email"] = dataset.Contact.Email
	}

	if dataset.Contact.Name != "" {
		updates["next.contact.name"] = dataset.Contact.Name
	}

	if dataset.Contact.Telephone != "" {
		updates["next.contact.telephone"] = dataset.Contact.Telephone
	}

	if dataset.Description != "" {
		updates["next.description"] = dataset.Description
	}

	if dataset.NextRelease != "" {
		updates["next.next_release"] = dataset.NextRelease
	}

	if dataset.Periodicity != "" {
		updates["next.periodicity"] = dataset.Periodicity
	}

	if dataset.Publisher.HRef != "" {
		updates["next.publisher.href"] = dataset.Publisher.HRef
	}

	if dataset.Publisher.Name != "" {
		updates["next.publisher.name"] = dataset.Publisher.Name
	}

	if dataset.Publisher.Type != "" {
		updates["next.publisher.type"] = dataset.Publisher.Type
	}

	if dataset.Theme != "" {
		updates["next.theme"] = dataset.Theme
	}

	if dataset.Title != "" {
		updates["next.title"] = dataset.Title
	}
	return updates
}

// UpdateDatasetWithAssociation updates an existing dataset document with collection data
func (m *Mongo) UpdateDatasetWithAssociation(id, state string, version *models.Version) (err error) {
	s := session.Copy()
	defer s.Close()

	update := bson.M{
		"$set": bson.M{
			"next.state":                     state,
			"next.collection_id":             version.CollectionID,
			"next.links.latest_version.link": version.Links.Self,
			"next.links.latest_version.id":   version.ID,
			"next.last_updated":              time.Now(),
		},
	}

	err = s.DB(m.Database).C("datasets").UpdateId(id, update)
	return
}

// UpdateEdition updates an existing edition document
func (m *Mongo) UpdateEdition(id, state string) (err error) {
	s := session.Copy()
	defer s.Close()

	update := bson.M{
		"$set": bson.M{
			"state": state,
		},
		"$setOnInsert": bson.M{
			"last_updated": time.Now(),
		},
	}

	err = s.DB(m.Database).C("editions").UpdateId(id, update)
	return
}

// UpdateVersion updates an existing version document
func (m *Mongo) UpdateVersion(id string, version *models.Version) (err error) {
	s := session.Copy()
	defer s.Close()

	updates := createVersionUpdateQuery(version)

	err = s.DB(m.Database).C("versions").UpdateId(id, bson.M{"$set": updates, "$setOnInsert": bson.M{"last_updated": time.Now()}})
	return
}

func createVersionUpdateQuery(version *models.Version) bson.M {
	updates := make(bson.M, 0)

	if version.CollectionID != "" {
		updates["collection_id"] = version.CollectionID
	}

	if version.InstanceID != "" {
		updates["instance_id"] = version.InstanceID
	}

	if version.License != "" {
		updates["license"] = version.License
	}

	if version.ReleaseDate != "" {
		updates["release_date"] = version.ReleaseDate
	}

	if version.State != "" {
		updates["state"] = version.State
	}

	return updates
}

// UpsertDataset adds or overides an existing dataset document
func (m *Mongo) UpsertDataset(id string, datasetDoc *models.DatasetUpdate) (err error) {
	s := session.Copy()
	defer s.Close()

	update := bson.M{
		"$set": datasetDoc,
		"$setOnInsert": bson.M{
			"last_updated": time.Now(),
		},
	}

	_, err = s.DB(m.Database).C("datasets").UpsertId(id, update)
	return
}

// UpsertEdition adds or overides an existing edition document
func (m *Mongo) UpsertEdition(editionID string, editionDoc *models.Edition) (err error) {
	s := session.Copy()
	defer s.Close()

	update := bson.M{
		"$set": editionDoc,
		"$setOnInsert": bson.M{
			"last_updated": time.Now(),
		},
	}

	_, err = s.DB(m.Database).C("editions").Upsert(bson.M{"edition": editionID}, update)
	return
}

// UpsertVersion adds or overides an existing version document
func (m *Mongo) UpsertVersion(id string, version *models.Version) (err error) {
	s := session.Copy()
	defer s.Close()

	update := bson.M{
		"$set": version,
		"$setOnInsert": bson.M{
			"last_updated": time.Now(),
		},
	}

	_, err = s.DB(m.Database).C("versions").UpsertId(id, update)
	return
}

// UpsertContact adds or overides an existing contact document
func (m *Mongo) UpsertContact(id string, update interface{}) (err error) {
	s := session.Copy()
	defer s.Close()

	_, err = s.DB(m.Database).C("contacts").UpsertId(id, update)
	return
}

func (m *Mongo) Close(ctx context.Context) error {
	closedChannel := make(chan bool)
	defer close(closedChannel)
	go func() {
		session.Close()
		closedChannel <- true
	}()
	timeLeft := 1000 * time.Millisecond
	if deadline, ok := ctx.Deadline(); ok {
		timeLeft = deadline.Sub(time.Now())
	}
	for {
		select {
		case <-time.After(timeLeft):
			return errors.New("closing mongo timed out")
		case <-closedChannel:
			return nil
		}
	}
}