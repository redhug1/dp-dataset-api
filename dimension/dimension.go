package dimension

import (
	"encoding/json"
	"net/http"

	"errors"
	"io"
	"io/ioutil"

	errs "github.com/ONSdigital/dp-dataset-api/apierrors"
	"github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-dataset-api/store"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

//Store provides a backend for dimensions
type Store struct {
	store.Storer
}

//GetNodes list from a specified instance
func (s *Store) GetNodes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	results, err := s.GetDimensionNodesFromInstance(id)
	if err != nil {
		log.ErrorC("GetNodes get", err, nil)
		handleErrorType(err, w)
		return
	}

	bytes, err := json.Marshal(results)
	if err != nil {
		internalError(w, err)
		return
	}

	writeBody(w, bytes)
	log.Debug("get dimension nodes", log.Data{"instance": id})
}

//GetUnique dimension values from a specified dimension
func (s *Store) GetUnique(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	dimension := vars["dimension"]

	values, err := s.GetUniqueDimensionValues(id, dimension)
	if err != nil {
		log.ErrorC("GetUnique get", err, nil)
		handleErrorType(err, w)
		return
	}

	bytes, err := json.Marshal(values)
	if err != nil {
		internalError(w, err)
		return
	}

	writeBody(w, bytes)
	log.Debug("get dimension values", log.Data{"instance": id})
}

//Add dimension to a specific instance
func (s *Store) Add(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	option, err := unmarshalDimensionCache(r.Body)
	if err != nil {
		log.ErrorC("Add json", err, nil)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	option.InstanceID = id
	if err := s.AddDimensionToInstance(option); err != nil {
		log.ErrorC("Add add", err, nil)
		handleErrorType(err, w)
	}
}

//AddNodeID against a specific value for dimension
func (s *Store) AddNodeID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	dimensionName := vars["dimension"]
	value := vars["value"]
	nodeID := vars["node_id"]

	dim := models.DimensionOption{Name: dimensionName, Option: value, NodeID: nodeID, InstanceID: id}
	if err := s.UpdateDimensionNodeID(&dim); err != nil {
		log.ErrorC("AddNodeID up", err, nil)
		handleErrorType(err, w)
	}
}

// unmarshalDimensionCache manages the creation of a dataset from a reader
func unmarshalDimensionCache(reader io.Reader) (*models.CachedDimensionOption, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.New("Failed to read message body")
	}

	var option models.CachedDimensionOption

	err = json.Unmarshal(bytes, &option)
	if err != nil {
		return nil, errors.New("Failed to parse json body")

	}
	if option.Name == "" || (option.Option == "" && option.CodeList == "") {
		return nil, errors.New("Missing properties in JSON")
	}

	return &option, nil
}

func handleErrorType(err error, w http.ResponseWriter) {
	status := http.StatusInternalServerError

	if err == errs.DatasetNotFound || err == errs.EditionNotFound || err == errs.VersionNotFound || err == errs.DimensionNodeNotFound || err == errs.InstanceNotFound {
		status = http.StatusNotFound
	}

	http.Error(w, err.Error(), status)

}

func internalError(w http.ResponseWriter, err error) {
	log.ErrorC("internalError", err, nil)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func writeBody(w http.ResponseWriter, bytes []byte) {
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(bytes); err != nil {
		log.ErrorC("writeBody", err, nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
