package instance_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ONSdigital/dp-dataset-api/api"
	errs "github.com/ONSdigital/dp-dataset-api/apierrors"
	"github.com/ONSdigital/dp-dataset-api/config"
	"github.com/ONSdigital/dp-dataset-api/instance"
	"github.com/ONSdigital/dp-dataset-api/mocks"
	"github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-dataset-api/store"
	"github.com/ONSdigital/dp-dataset-api/store/datastoretest"
	"github.com/ONSdigital/dp-dataset-api/url"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/audit/audit_mock"
	"github.com/ONSdigital/go-ns/common"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

const host = "http://localhost:8080"

var errAudit = errors.New("auditing error")
var updateImportTaskAction = "updateImportTask"

func createRequestWithToken(method, url string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, url, body)
	ctx := r.Context()
	ctx = common.SetCaller(ctx, "someone@ons.gov.uk")
	r = r.WithContext(ctx)
	return r, err
}

func TestGetInstancesReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("Get instances returns a ok status code", t, func() {
		r, err := createRequestWithToken("GET", "http://localhost:21800/instances", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetInstancesFunc: func([]string) (*models.InstanceResults, error) {
				return &models.InstanceResults{}, nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: "http://lochost://8080", Storer: mockedDataStore, Auditor: auditor}
		instanceAPI.GetList(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetInstancesCalls()), ShouldEqual, 1)
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(instance.GetInstancesAction, audit.Attempted, nil),
			audit_mock.NewExpectation(instance.GetInstancesAction, audit.Successful, nil),
		)
	})
}

func TestGetInstancesFiltersOnState(t *testing.T) {
	t.Parallel()
	Convey("Get instances filtered by a single state value returns only instances with that value", t, func() {
		r, err := createRequestWithToken("GET", "http://localhost:21800/instances?state=completed", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		var result []string

		mockedDataStore := &storetest.StorerMock{
			GetInstancesFunc: func(filterString []string) (*models.InstanceResults, error) {
				result = filterString
				return &models.InstanceResults{}, nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: "http://lochost://8080", Storer: mockedDataStore, Auditor: auditor}
		instanceAPI.GetList(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetInstancesCalls()), ShouldEqual, 1)
		So(result, ShouldResemble, []string{models.CompletedState})
		expectedParams := common.Params{"query": "completed"}
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(instance.GetInstancesAction, audit.Attempted, expectedParams),
			audit_mock.NewExpectation(instance.GetInstancesAction, audit.Successful, expectedParams),
		)
	})

	Convey("Get instances filtered by multiple state values returns only instances with those values", t, func() {
		r, err := createRequestWithToken("GET", "http://localhost:21800/instances?state=completed,edition-confirmed", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		var result []string

		mockedDataStore := &storetest.StorerMock{
			GetInstancesFunc: func(filterString []string) (*models.InstanceResults, error) {
				result = filterString
				return &models.InstanceResults{}, nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: "http://lochost://8080", Storer: mockedDataStore, Auditor: auditor}
		instanceAPI.GetList(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetInstancesCalls()), ShouldEqual, 1)
		So(result, ShouldResemble, []string{models.CompletedState, models.EditionConfirmedState})
		expectedParams := common.Params{"query": "completed,edition-confirmed"}
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(instance.GetInstancesAction, audit.Attempted, expectedParams),
			audit_mock.NewExpectation(instance.GetInstancesAction, audit.Successful, expectedParams),
		)
	})
}

func TestGetInstancesReturnsError(t *testing.T) {
	t.Parallel()
	Convey("Get instances returns an internal error", t, func() {
		r, err := createRequestWithToken("GET", "http://localhost:21800/instances", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetInstancesFunc: func([]string) (*models.InstanceResults, error) {
				return nil, errs.ErrInternalServer
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}
		instanceAPI.GetList(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.GetInstancesCalls()), ShouldEqual, 1)
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(instance.GetInstancesAction, audit.Attempted, nil),
			audit_mock.NewExpectation(instance.GetInstancesAction, audit.Unsuccessful, nil),
		)
	})

	Convey("Get instances returns bad request error", t, func() {
		r, err := createRequestWithToken("GET", "http://localhost:21800/instances?state=foo", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}
		instanceAPI.GetList(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(len(mockedDataStore.GetInstancesCalls()), ShouldEqual, 0)
		expectedParams := common.Params{"query": "foo"}
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(instance.GetInstancesAction, audit.Attempted, expectedParams),
			audit_mock.NewExpectation(instance.GetInstancesAction, audit.Unsuccessful, expectedParams),
		)
	})
}

func TestGetInstancesAuditErrors(t *testing.T) {
	t.Parallel()
	Convey("given audit action attempted returns an error", t, func() {

		auditor := audit_mock.NewErroring(instance.GetInstancesAction, audit.Attempted)

		Convey("when get instances is called", func() {
			r, err := createRequestWithToken("GET", "http://localhost:21800/instances", nil)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{}

			instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}
			instanceAPI.GetList(w, r)

			Convey("then a 500 status is returned", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(len(mockedDataStore.GetInstancesCalls()), ShouldEqual, 0)
				auditor.AssertRecordCalls(audit_mock.NewExpectation(instance.GetInstancesAction, audit.Attempted, nil))
			})
		})
	})

	Convey("given audit action unsuccessful returns an error", t, func() {
		auditor := audit_mock.NewErroring(instance.GetInstancesAction, audit.Unsuccessful)

		Convey("when get instances return an error", func() {
			r, err := createRequestWithToken("GET", "http://localhost:21800/instances", nil)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{
				GetInstancesFunc: func([]string) (*models.InstanceResults, error) {
					return nil, errs.ErrInternalServer
				},
			}

			instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}
			instanceAPI.GetList(w, r)

			Convey("then a 500 status is returned", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(len(mockedDataStore.GetInstancesCalls()), ShouldEqual, 1)
				auditor.AssertRecordCalls(
					audit_mock.NewExpectation(instance.GetInstancesAction, audit.Attempted, nil),
					audit_mock.NewExpectation(instance.GetInstancesAction, audit.Unsuccessful, nil),
				)
			})
		})
	})

	Convey("given audit action successful returns an error", t, func() {
		auditor := audit_mock.NewErroring(instance.GetInstancesAction, audit.Successful)

		Convey("when get instances is called", func() {
			r, err := createRequestWithToken("GET", "http://localhost:21800/instances", nil)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{
				GetInstancesFunc: func([]string) (*models.InstanceResults, error) {
					return &models.InstanceResults{}, nil
				},
			}

			instanceAPI := &instance.Store{Host: "http://lochost://8080", Storer: mockedDataStore, Auditor: auditor}
			instanceAPI.GetList(w, r)

			Convey("then a 500 status is returned", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(len(mockedDataStore.GetInstancesCalls()), ShouldEqual, 1)
				auditor.AssertRecordCalls(
					audit_mock.NewExpectation(instance.GetInstancesAction, audit.Attempted, nil),
					audit_mock.NewExpectation(instance.GetInstancesAction, audit.Successful, nil),
				)
			})
		})
	})
}

func TestGetInstanceReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("Get instance returns a ok status code", t, func() {
		r, err := createRequestWithToken("GET", "http://localhost:21800/instances/123", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(ID string) (*models.Instance, error) {
				return &models.Instance{State: models.CreatedState}, nil
			},
		}

		auditor := audit_mock.New()
		datasetAPI := getAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditor, &mocks.ObservationStoreMock{})
		datasetAPI.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)

		auditParams := common.Params{"instance_id": "123"}
		So(len(auditor.RecordCalls()), ShouldEqual, 2)
		auditor.AssertRecordCalls(
			audit_mock.Expected{instance.GetInstanceAction, audit.Attempted, auditParams},
			audit_mock.Expected{instance.GetInstanceAction, audit.Successful, auditParams},
		)
	})
}

func TestGetInstanceReturnsInternalError(t *testing.T) {
	t.Parallel()
	Convey("Given an internal error is returned from mongo, then response returns an internal error", t, func() {
		r, err := createRequestWithToken("GET", "http://localhost:21800/instances/123", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(ID string) (*models.Instance, error) {
				return nil, errs.ErrInternalServer
			},
		}

		auditor := audit_mock.New()
		datasetAPI := getAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditor, &mocks.ObservationStoreMock{})
		datasetAPI.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(auditor.RecordCalls()), ShouldEqual, 2)

		auditParams := common.Params{"instance_id": "123"}
		auditor.AssertRecordCalls(
			audit_mock.Expected{instance.GetInstanceAction, audit.Attempted, auditParams},
			audit_mock.Expected{instance.GetInstanceAction, audit.Unsuccessful, auditParams},
		)
	})

	Convey("Given instance state is invalid, then response returns an internal error", t, func() {
		r, err := createRequestWithToken("GET", "http://localhost:21800/instances/123", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(ID string) (*models.Instance, error) {
				return &models.Instance{State: "gobbly gook"}, nil
			},
		}

		auditor := audit_mock.New()
		datasetAPI := getAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditor, &mocks.ObservationStoreMock{})
		datasetAPI.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(auditor.RecordCalls()), ShouldEqual, 2)

		auditParams := common.Params{"instance_id": "123"}
		auditor.AssertRecordCalls(
			audit_mock.Expected{instance.GetInstanceAction, audit.Attempted, auditParams},
			audit_mock.Expected{instance.GetInstanceAction, audit.Unsuccessful, auditParams},
		)
	})
}

func TestGetInstanceAuditErrors(t *testing.T) {
	t.Parallel()
	Convey("Given audit action 'attempted' fails", t, func() {

		auditor := audit_mock.NewErroring(instance.GetInstanceAction, audit.Attempted)

		Convey("When a GET request is made to get an instance resource", func() {
			r, err := createRequestWithToken("GET", "http://localhost:21800/instances/123", nil)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{
				GetInstanceFunc: func(ID string) (*models.Instance, error) {
					return nil, errs.ErrInternalServer
				},
			}

			datasetAPI := getAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditor, &mocks.ObservationStoreMock{})
			datasetAPI.Router.ServeHTTP(w, r)

			Convey("Then response returns internal server error (500)", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)

				So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 0)
				So(len(auditor.RecordCalls()), ShouldEqual, 1)

				auditParams := common.Params{"instance_id": "123"}
				auditor.AssertRecordCalls(
					audit_mock.Expected{instance.GetInstanceAction, audit.Attempted, auditParams},
				)
			})
		})
	})

	Convey("Given audit action 'unsuccessful' fails", t, func() {

		auditor := audit_mock.NewErroring(instance.GetInstanceAction, audit.Unsuccessful)

		Convey("When a GET request is made to get an instance resource", func() {
			r, err := createRequestWithToken("GET", "http://localhost:21800/instances/123", nil)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{
				GetInstanceFunc: func(ID string) (*models.Instance, error) {
					return nil, errs.ErrInternalServer
				},
			}

			datasetAPI := getAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditor, &mocks.ObservationStoreMock{})
			datasetAPI.Router.ServeHTTP(w, r)

			Convey("Then response returns internal server error (500)", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)

				So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
				So(len(auditor.RecordCalls()), ShouldEqual, 2)

				auditParams := common.Params{"instance_id": "123"}
				auditor.AssertRecordCalls(
					audit_mock.Expected{instance.GetInstanceAction, audit.Attempted, auditParams},
					audit_mock.Expected{instance.GetInstanceAction, audit.Unsuccessful, auditParams},
				)
			})
		})
	})

	Convey("Given audit action 'successful' fails", t, func() {

		auditor := audit_mock.NewErroring(instance.GetInstanceAction, audit.Successful)

		Convey("When a GET request is made to get an instance resource", func() {
			r, err := createRequestWithToken("GET", "http://localhost:21800/instances/123", nil)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{
				GetInstanceFunc: func(ID string) (*models.Instance, error) {
					return &models.Instance{State: models.CreatedState}, nil
				},
			}

			datasetAPI := getAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditor, &mocks.ObservationStoreMock{})
			datasetAPI.Router.ServeHTTP(w, r)

			Convey("Then response returns internal server error (500)", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)

				So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
				So(len(auditor.RecordCalls()), ShouldEqual, 2)

				auditParams := common.Params{"instance_id": "123"}
				auditor.AssertRecordCalls(
					audit_mock.Expected{instance.GetInstanceAction, audit.Attempted, auditParams},
					audit_mock.Expected{instance.GetInstanceAction, audit.Successful, auditParams},
				)
			})
		})
	})
}

func TestAddInstancesReturnsCreated(t *testing.T) {
	t.Parallel()
	Convey("Add instance returns a created code", t, func() {
		body := strings.NewReader(`{"links": { "job": { "id":"123-456", "href":"http://localhost:2200/jobs/123-456" } } }`)
		r, err := createRequestWithToken("POST", "http://localhost:21800/instances", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			AddInstanceFunc: func(*models.Instance) (*models.Instance, error) {
				return &models.Instance{}, nil
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.Add(w, r)

		So(w.Code, ShouldEqual, http.StatusCreated)
		So(len(mockedDataStore.AddInstanceCalls()), ShouldEqual, 1)
	})
}

func TestAddInstancesReturnsBadRequest(t *testing.T) {
	t.Parallel()
	Convey("Add instance returns a bad request with invalid json", t, func() {
		body := strings.NewReader(`{`)
		r, err := createRequestWithToken("POST", "http://localhost:21800/instances", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			AddInstanceFunc: func(*models.Instance) (*models.Instance, error) {
				return &models.Instance{}, nil
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.Add(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Add instance returns a bad request with a empty json", t, func() {
		body := strings.NewReader(`{}`)
		r, err := createRequestWithToken("POST", "http://localhost:21800/instances", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			AddInstanceFunc: func(*models.Instance) (*models.Instance, error) {
				return &models.Instance{}, nil
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.Add(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})
}

func TestAddInstancesReturnsInternalError(t *testing.T) {
	t.Parallel()
	Convey("Add instance returns an internal error", t, func() {
		body := strings.NewReader(`{"links": {"job": { "id":"123-456", "href":"http://localhost:2200/jobs/123-456" } } }`)
		r, err := createRequestWithToken("POST", "http://localhost:21800/instances", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			AddInstanceFunc: func(*models.Instance) (*models.Instance, error) {
				return nil, errs.ErrInternalServer
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.Add(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.AddInstanceCalls()), ShouldEqual, 1)
	})
}

func TestUpdateInstanceReturnsOk(t *testing.T) {
	t.Parallel()
	Convey("when an instance has a state of created", t, func() {
		body := strings.NewReader(`{"state":"created"}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return &models.Instance{State: models.CreatedState}, nil
			},
			UpdateInstanceFunc: func(id string, i *models.Instance) error {
				return nil
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.Update(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 0)
	})

	Convey("when an instance changes its state to edition-confirmed", t, func() {
		body := strings.NewReader(`{"state":"edition-confirmed", "edition": "2017"}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		currentInstanceTestData := &models.Instance{
			Edition: "2017",
			Links: &models.InstanceLinks{
				Job: &models.IDLink{
					ID:   "7654",
					HRef: "job-link",
				},
				Dataset: &models.IDLink{
					ID:   "4567",
					HRef: "dataset-link",
				},
				Self: &models.IDLink{
					HRef: "self-link",
				},
			},
			State: models.CompletedState,
		}

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return currentInstanceTestData, nil
			},
			GetEditionFunc: func(datasetID string, edition string, state string) (*models.EditionUpdate, error) {
				return nil, errs.ErrEditionNotFound
			},
			UpsertEditionFunc: func(datasetID, edition string, editionDoc *models.EditionUpdate) error {
				return nil
			},
			GetNextVersionFunc: func(string, string) (int, error) {
				return 1, nil
			},
			UpdateInstanceFunc: func(id string, i *models.Instance) error {
				return nil
			},
			AddVersionDetailsToInstanceFunc: func(ctx context.Context, instanceID string, datasetID string, edition string, version int) error {
				return nil
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.Update(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpsertEditionCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetNextVersionCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 1)
	})
}

func TestUpdateInstanceReturnsInternalError(t *testing.T) {
	t.Parallel()
	Convey("Given an internal error is returned from mongo, then response returns an internal error", t, func() {
		body := strings.NewReader(`{"state":"created"}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return nil, errs.ErrInternalServer
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.Update(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldResemble, "internal error\n")

		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 0)
	})

	Convey("Given the current instance state is invalid, then response returns an internal error", t, func() {
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123", strings.NewReader(`{"state":"completed", "edition": "2017"}`))
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return &models.Instance{State: "gobbly gook"}, nil
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.Update(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldResemble, "internal error\n")

		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 0)
	})
}

func TestUpdateInstanceFailure(t *testing.T) {
	t.Parallel()
	Convey("when the json body is in the incorrect structure return a bad request error", t, func() {
		body := strings.NewReader(`{"state":`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.Update(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 0)
	})

	Convey("when the instance does not exist return status not found", t, func() {
		body := strings.NewReader(`{"edition": "2017"}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return nil, errs.ErrInstanceNotFound
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.Update(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 0)
	})
	Convey("when store.AddVersionDetailsToInstance return an error", t, func() {
		body := strings.NewReader(`{"state":"edition-confirmed", "edition": "2017"}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		currentInstanceTestData := &models.Instance{
			Edition: "2017",
			Links: &models.InstanceLinks{
				Job: &models.IDLink{
					ID:   "7654",
					HRef: "job-link",
				},
				Dataset: &models.IDLink{
					ID:   "4567",
					HRef: "dataset-link",
				},
				Self: &models.IDLink{
					HRef: "self-link",
				},
			},
			State: models.CompletedState,
		}

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return currentInstanceTestData, nil
			},
			GetEditionFunc: func(datasetID string, edition string, state string) (*models.EditionUpdate, error) {
				return nil, errs.ErrEditionNotFound
			},
			UpsertEditionFunc: func(datasetID, edition string, editionDoc *models.EditionUpdate) error {
				return nil
			},
			GetNextVersionFunc: func(string, string) (int, error) {
				return 1, nil
			},
			UpdateInstanceFunc: func(id string, i *models.Instance) error {
				return nil
			},
			AddVersionDetailsToInstanceFunc: func(ctx context.Context, instanceID string, datasetID string, edition string, version int) error {
				return errors.New("boom")
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.Update(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpsertEditionCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetNextVersionCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 0)
	})
}

func TestUpdatePublishedInstanceToCompletedReturnsForbidden(t *testing.T) {
	t.Parallel()
	Convey("Given a 'published' instance, when we update to 'completed' then we get a bad-request error", t, func() {
		body := strings.NewReader(`{"state":"completed"}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/1235", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		currentInstanceTestData := &models.Instance{
			Edition: "2017",
			Links: &models.InstanceLinks{
				Dataset: &models.IDLink{
					ID: "4567",
				},
			},
			State: models.PublishedState,
		}

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return currentInstanceTestData, nil
			},
			UpdateInstanceFunc: func(id string, i *models.Instance) error {
				return nil
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.Update(w, r)

		So(w.Code, ShouldEqual, http.StatusForbidden)
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 0)
	})
}

func TestUpdateEditionConfirmedInstanceToCompletedReturnsForbidden(t *testing.T) {
	t.Parallel()
	Convey("update to an instance returns an internal error", t, func() {
		body := strings.NewReader(`{"state":"completed"}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		currentInstanceTestData := &models.Instance{
			Edition: "2017",
			Links: &models.InstanceLinks{
				Dataset: &models.IDLink{
					ID: "4567",
				},
			},
			State: models.EditionConfirmedState,
		}

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return currentInstanceTestData, nil
			},
			UpdateInstanceFunc: func(id string, i *models.Instance) error {
				return nil
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.Update(w, r)

		So(w.Code, ShouldEqual, http.StatusForbidden)
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 0)
	})
}

func TestInsertedObservationsReturnsOk(t *testing.T) {
	t.Parallel()
	Convey("Updateding the inserted observations returns ok", t, func() {
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/inserted_observations/200", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateObservationInsertedFunc: func(id string, ob int64) error {
				return nil
			},
		}
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}

		router := mux.NewRouter()
		router.HandleFunc("/instances/{id}/inserted_observations/{inserted_observations}", instanceAPI.UpdateObservations).Methods("PUT")
		router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.UpdateObservationInsertedCalls()), ShouldEqual, 1)
	})
}

func TestInsertedObservationsReturnsBadRequest(t *testing.T) {
	t.Parallel()
	Convey("Updateding the inserted observations returns bad request", t, func() {
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/inserted_observations/aa12a", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.UpdateObservations(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})
}

func TestInsertedObservationsReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("Updating the inserted observations returns not found", t, func() {
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/inserted_observations/200", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateObservationInsertedFunc: func(id string, ob int64) error {
				return errs.ErrInstanceNotFound
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}

		router := mux.NewRouter()
		router.HandleFunc("/instances/{id}/inserted_observations/{inserted_observations}", instanceAPI.UpdateObservations).Methods("PUT")
		router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.UpdateObservationInsertedCalls()), ShouldEqual, 1)
	})
}

func TestStore_UpdateImportTask_UpdateImportObservations(t *testing.T) {

	t.Parallel()
	Convey("update to an import task returns http 200 response if no errors occur", t, func() {
		body := strings.NewReader(`{"import_observations":{"state":"completed"}}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateImportObservationsTaskStateFunc: func(id string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string. )
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Successful, ap),
		)
	})
}

func TestStore_UpdateImportTask_UpdateImportObservations_Failure(t *testing.T) {

	t.Parallel()
	Convey("update to an import task with invalid json returns http 400 response", t, func() {
		body := strings.NewReader(`{`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateImportObservationsTaskStateFunc: func(id string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "failed to parse json body: unexpected end of JSON input")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})

	Convey("update to an import task but missing mandatory field, 'state' returns http 400 response", t, func() {
		body := strings.NewReader(`{"import_observations":{}}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateImportObservationsTaskStateFunc: func(id string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "bad request - invalid import observation task, must include state")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})

	Convey("update to an import task with an invalid state returns http 400 response", t, func() {
		body := strings.NewReader(`{"import_observations":{"state":"notvalid"}}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateImportObservationsTaskStateFunc: func(id string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "bad request - invalid task state value for import observations: notvalid")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.

		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})
}

func TestStore_UpdateImportTask_UpdateBuildHierarchyTask_Failure(t *testing.T) {

	t.Parallel()
	Convey("update to an import task with invalid json returns http 400 response", t, func() {
		body := strings.NewReader(`{`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildHierarchyTaskStateFunc: func(id string, dimension string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "failed to parse json body: unexpected end of JSON input")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})

	Convey("update to an import task with an empty request body returns http 400 response", t, func() {
		body := strings.NewReader(`{}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildHierarchyTaskStateFunc: func(id string, dimension string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "bad request - request body does not contain any import tasks")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})

	Convey("update to an import task without specifying a task returns http 400 response", t, func() {
		body := strings.NewReader(`{"build_hierarchies":[]}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildHierarchyTaskStateFunc: func(id string, dimension string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "bad request - missing hierarchy task")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})

	Convey("update to an import task without a 'dimension_name' returns http 400 response", t, func() {
		body := strings.NewReader(`{"build_hierarchies":[{"state":"completed"}]}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildHierarchyTaskStateFunc: func(id string, dimension string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "bad request - missing mandatory fields: [dimension_name]")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})

	Convey("update to an import task without a 'dimension_name' returns http 400 response", t, func() {
		body := strings.NewReader(`{"build_hierarchies":[{"dimension_name":"geography"}]}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildHierarchyTaskStateFunc: func(id string, dimension string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "bad request - missing mandatory fields: [state]")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})

	Convey("update to an import task with an invalid state returns http 400 response", t, func() {
		body := strings.NewReader(`{"build_hierarchies":[{"state":"notvalid", "dimension_name": "geography"}]}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildHierarchyTaskStateFunc: func(id string, dimension string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "bad request - invalid task state value: notvalid")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})

	Convey("update to an import task with an invalid state returns http 400 response", t, func() {
		body := strings.NewReader(`{"build_hierarchies":[{"state":"completed", "dimension_name": "geography"}]}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildHierarchyTaskStateFunc: func(id string, dimension string, state string) error {
				return errors.New("not found")
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, "geography hierarchy import task does not exist")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})

	Convey("update to an import task but lose connection to datastore when updating resource", t, func() {
		body := strings.NewReader(`{"build_hierarchies":[{"state":"completed", "dimension_name": "geography"}]}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildHierarchyTaskStateFunc: func(id string, dimension string, state string) error {
				return errors.New("internal error")
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldContainSubstring, "internal error")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})
}

func TestStore_UpdateImportTask_UpdateBuildHierarchyTask(t *testing.T) {

	t.Parallel()
	Convey("update to an import task returns http 200 response if no errors occur", t, func() {
		body := strings.NewReader(`{"build_hierarchies":[{"state":"completed", "dimension_name":"geography"}]}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildHierarchyTaskStateFunc: func(id string, dimension string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 1)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Successful, ap),
		)
	})
}

func TestStore_UpdateImportTask_ReturnsInternalError(t *testing.T) {

	t.Parallel()
	Convey("update to an import task returns an internal error", t, func() {
		body := strings.NewReader(`{"import_observations":{"state":"completed"}}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateImportObservationsTaskStateFunc: func(id string, state string) error {
				return errs.ErrInternalServer
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 1)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})
}

func TestUpdateInstanceReturnsErrorWhenStateIsPublished(t *testing.T) {
	t.Parallel()
	Convey("when an instance has a state of published, then put request to change to it to completed ", t, func() {
		body := strings.NewReader(`{"state":"completed"}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return &models.Instance{State: models.PublishedState}, nil
			},
			UpdateInstanceFunc: func(id string, i *models.Instance) error {
				return nil
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.Update(w, r)

		So(w.Code, ShouldEqual, http.StatusForbidden)
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 0)
	})
}

func TestUpdateDimensionReturnsInternalError(t *testing.T) {
	t.Parallel()
	Convey("Given an internal error is returned from mongo, then response returns an internal error", t, func() {
		body := strings.NewReader(`{"label":"ages"}`)
		r, err := createRequestWithToken("PUT", "http://localhost:22000/instances/123/dimensions/age", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return nil, errs.ErrInternalServer
			},
			UpdateInstanceFunc: func(id string, i *models.Instance) error {
				return nil
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.UpdateDimension(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 0)
	})

	Convey("Given the instance state is invalid, then response returns an internal error", t, func() {
		body := strings.NewReader(`{"label":"ages"}`)
		r, err := createRequestWithToken("PUT", "http://localhost:22000/instances/123/dimensions/age", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return &models.Instance{State: "gobbly gook"}, nil
			},
			UpdateInstanceFunc: func(id string, i *models.Instance) error {
				return nil
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.UpdateDimension(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 0)
	})
}

func TestUpdateDimensionReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When update dimension return status not found", t, func() {
		r, err := createRequestWithToken("PUT", "http://localhost:22000/instances/123/dimensions/age", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return nil, errs.ErrInstanceNotFound
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.UpdateDimension(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 0)
	})
}

func TestUpdateDimensionReturnsForbidden(t *testing.T) {
	t.Parallel()
	Convey("When update dimension returns forbidden (for already published) ", t, func() {
		r, err := createRequestWithToken("PUT", "http://localhost:22000/instances/123/dimensions/age", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		currentInstanceTestData := &models.Instance{
			State: models.PublishedState,
		}

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return currentInstanceTestData, nil
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.UpdateDimension(w, r)

		So(w.Code, ShouldEqual, http.StatusForbidden)
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 0)
	})
}

func TestUpdateDimensionReturnsBadRequest(t *testing.T) {
	t.Parallel()
	Convey("When update dimension returns bad request", t, func() {
		body := strings.NewReader("{")
		r, err := createRequestWithToken("PUT", "http://localhost:22000/instances/123/dimensions/age", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return &models.Instance{State: models.CompletedState}, nil
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		instanceAPI.UpdateDimension(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})
}

func TestUpdateDimensionReturnsNotFoundWithWrongName(t *testing.T) {
	t.Parallel()
	Convey("When update dimension fails to update an instance", t, func() {
		body := strings.NewReader(`{"label":"notages"}`)
		r, err := createRequestWithToken("PUT", "http://localhost:22000/instances/123/dimensions/notage", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return &models.Instance{State: models.EditionConfirmedState,
					InstanceID: "123",
					Dimensions: []models.CodeList{{Name: "age", ID: "age"}}}, nil
			},
			UpdateInstanceFunc: func(id string, i *models.Instance) error {
				return nil
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}

		router := mux.NewRouter()
		router.HandleFunc("/instances/{id}/dimensions/{dimension}", instanceAPI.UpdateDimension).Methods("PUT")

		router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 0)
	})
}

func TestUpdateDimensionReturnsOk(t *testing.T) {
	t.Parallel()
	Convey("When update dimension fails to update an instance", t, func() {
		body := strings.NewReader(`{"label":"ages"}`)
		r, err := createRequestWithToken("PUT", "http://localhost:22000/instances/123/dimensions/age", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetInstanceFunc: func(id string) (*models.Instance, error) {
				return &models.Instance{State: models.EditionConfirmedState,
					InstanceID: "123",
					Dimensions: []models.CodeList{{Name: "age", ID: "age"}}}, nil
			},
			UpdateInstanceFunc: func(id string, i *models.Instance) error {
				return nil
			},
		}

		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore}
		router := mux.NewRouter()
		router.HandleFunc("/instances/{id}/dimensions/{dimension}", instanceAPI.UpdateDimension).Methods("PUT")

		router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 1)
	})
}

func TestStore_UpdateImportTask_UpdateBuildSearchIndexTask_Failure(t *testing.T) {

	t.Parallel()
	Convey("update to an import task with invalid json returns http 400 response", t, func() {
		body := strings.NewReader(`{`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildSearchTaskStateFunc: func(id string, dimension string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "failed to parse json body: unexpected end of JSON input")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})

	Convey("update to an import task with an empty request body returns http 400 response", t, func() {
		body := strings.NewReader(`{}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildSearchTaskStateFunc: func(id string, dimension string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "bad request - request body does not contain any import tasks")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})

	Convey("update to an import task without specifying a task returns http 400 response", t, func() {
		body := strings.NewReader(`{"build_search_indexes":[]}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildSearchTaskStateFunc: func(id string, dimension string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "bad request - missing search index task")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})

	Convey("update to an import task without a 'dimension_name' returns http 400 response", t, func() {
		body := strings.NewReader(`{"build_search_indexes":[{"state":"completed"}]}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildSearchTaskStateFunc: func(id string, dimension string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "bad request - missing mandatory fields: [dimension_name]")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)
	})

	Convey("update to an import task without a 'dimension_name' returns http 400 response", t, func() {
		body := strings.NewReader(`{"build_search_indexes":[{"dimension_name":"geography"}]}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildSearchTaskStateFunc: func(id string, dimension string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "bad request - missing mandatory fields: [state]")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})

	Convey("update to an import task with an invalid state returns http 400 response", t, func() {
		body := strings.NewReader(`{"build_search_indexes":[{"state":"notvalid", "dimension_name": "geography"}]}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildSearchTaskStateFunc: func(id string, dimension string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "bad request - invalid task state value: notvalid")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})

	Convey("update to an import task with a dimension that does not exist returns http 404 response", t, func() {
		body := strings.NewReader(`{"build_search_indexes":[{"state":"completed", "dimension_name": "geography"}]}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildSearchTaskStateFunc: func(id string, dimension string, state string) error {
				return errors.New("not found")
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, "geography search index import task does not exist")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 1)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})

	Convey("update to an import task but lose connection to datastore when updating resource", t, func() {
		body := strings.NewReader(`{"build_search_indexes":[{"state":"completed", "dimension_name": "geography"}]}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildSearchTaskStateFunc: func(id string, dimension string, state string) error {
				return errors.New("internal error")
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldContainSubstring, "internal error")
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 1)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
		)
	})
}

func TestStore_UpdateImportTask_UpdateBuildSearchIndexTask(t *testing.T) {

	t.Parallel()
	Convey("update to an import task returns http 200 response if no errors occur", t, func() {
		body := strings.NewReader(`{"build_search_indexes":[{"state":"completed", "dimension_name": "geography"}]}`)
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateBuildSearchTaskStateFunc: func(id string, dimension string, state string) error {
				return nil
			},
		}

		auditor := audit_mock.New()
		instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}

		instanceAPI.UpdateImportTask(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 1)

		ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
		// router so no URL params are available in the test - hence empty string.
		auditor.AssertRecordCalls(
			audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
			audit_mock.NewExpectation(updateImportTaskAction, audit.Successful, ap),
		)
	})
}

func TestStore_UpdateImportTask_AuditAttemptedError(t *testing.T) {
	t.Parallel()
	Convey("given audit action attempted returns an error", t, func() {
		auditor := audit_mock.NewErroring(updateImportTaskAction, audit.Attempted)

		Convey("when update import task is called", func() {
			body := strings.NewReader(`{"build_search_indexes":[{"state":"completed"}]}`)
			r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{}
			instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}
			instanceAPI.UpdateImportTask(w, r)

			Convey("then a 500 status is returned", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

				ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
				// router so no URL params are available in the test - hence empty string.
				auditor.AssertRecordCalls(audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap))
			})
		})
	})
}

func TestStore_UpdateImportTask_AuditUnsuccessfulError(t *testing.T) {
	t.Parallel()
	Convey("given audit action unsuccessful returns an error", t, func() {
		Convey("when the request body fails to marshal into the updateImportTask model", func() {
			auditor := audit_mock.NewErroring(updateImportTaskAction, audit.Unsuccessful)
			body := strings.NewReader(`THIS IS NOT JSON`)
			r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{}
			instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}
			instanceAPI.UpdateImportTask(w, r)

			Convey("then a 500 status is returned", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

				ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
				// router so no URL params are available in the test - hence empty string.
				auditor.AssertRecordCalls(
					audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
					audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
				)
			})
		})

		Convey("when UpdateImportObservationsTaskState returns an error", func() {
			auditor := audit_mock.NewErroring(updateImportTaskAction, audit.Unsuccessful)
			body := strings.NewReader(`{"import_observations":{"state":"completed"}}`)
			r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{
				UpdateImportObservationsTaskStateFunc: func(id string, state string) error {
					return errors.New("error")
				},
			}

			instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}
			instanceAPI.UpdateImportTask(w, r)

			Convey("then a 500 status is returned", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

				ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
				// router so no URL params are available in the test - hence empty string.
				auditor.AssertRecordCalls(
					audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
					audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
				)
			})
		})

		Convey("when UpdateBuildHierarchyTaskState returns an error", func() {
			auditor := audit_mock.NewErroring(updateImportTaskAction, audit.Unsuccessful)
			body := strings.NewReader(`{"build_hierarchies":[{"dimension_name": "geography", "state":"completed"}]}`)
			r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{
				UpdateBuildHierarchyTaskStateFunc: func(id string, dimension string, state string) error {
					return errors.New("error")
				},
			}

			instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}
			instanceAPI.UpdateImportTask(w, r)

			Convey("then a 500 status is returned", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, "internal error")
				So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

				ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
				// router so no URL params are available in the test - hence empty string.
				auditor.AssertRecordCalls(
					audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
					audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
				)
			})
		})

		Convey("when UpdateBuildSearchTaskState returns an error", func() {
			auditor := audit_mock.NewErroring(updateImportTaskAction, audit.Unsuccessful)
			body := strings.NewReader(`{"build_search_indexes":[{"dimension_name": "geography", "state":"completed"}]}`)
			r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{
				UpdateBuildSearchTaskStateFunc: func(id string, dimension string, state string) error {
					return errors.New("error")
				},
			}
			instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}
			instanceAPI.UpdateImportTask(w, r)

			Convey("then a 500 status is returned", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 1)

				ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
				// router so no URL params are available in the test - hence empty string.
				auditor.AssertRecordCalls(
					audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
					audit_mock.NewExpectation(updateImportTaskAction, audit.Unsuccessful, ap),
				)
			})
		})
	})
}

func TestStore_UpdateImportTask_AuditSuccessfulError(t *testing.T) {
	t.Parallel()
	Convey("given audit action successful returns an error", t, func() {
		auditor := audit_mock.NewErroring(updateImportTaskAction, audit.Successful)

		Convey("when update import task is called", func() {
			body := strings.NewReader(`{"import_observations":{"state":"completed"}}`)
			r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/import_tasks", body)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{
				UpdateImportObservationsTaskStateFunc: func(id string, state string) error {
					return nil
				},
			}
			instanceAPI := &instance.Store{Host: host, Storer: mockedDataStore, Auditor: auditor}
			instanceAPI.UpdateImportTask(w, r)

			Convey("then a 500 status is returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(len(mockedDataStore.UpdateImportObservationsTaskStateCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.UpdateBuildHierarchyTaskStateCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.UpdateBuildSearchTaskStateCalls()), ShouldEqual, 0)

				ap := common.Params{"ID": ""} //NOTE: ID comes from mux router url params but the test is not invoked via the
				// router so no URL params are available in the test - hence empty string.
				auditor.AssertRecordCalls(
					audit_mock.NewExpectation(updateImportTaskAction, audit.Attempted, ap),
					audit_mock.NewExpectation(updateImportTaskAction, audit.Successful, ap),
				)
			})
		})
	})
}

var urlBuilder = url.NewBuilder("localhost:20000")

func getAPIWithMockedDatastore(mockedDataStore store.Storer, mockedGeneratedDownloads api.DownloadsGenerator, mockAuditor api.Auditor, mockedObservationStore api.ObservationStore) *api.DatasetAPI {
	cfg, err := config.Get()
	So(err, ShouldBeNil)
	cfg.ServiceAuthToken = "dataset"
	cfg.DatasetAPIURL = "http://localhost:22000"
	cfg.EnablePrivateEnpoints = true
	cfg.HealthCheckTimeout = 2 * time.Second

	return api.Routes(*cfg, mux.NewRouter(), store.DataStore{Backend: mockedDataStore}, urlBuilder, mockedGeneratedDownloads, mockAuditor, mockedObservationStore)
}
