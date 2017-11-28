// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package mocks

import (
	"github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/go-ns/clients/filter"
	"sync"
)

var (
	lockFilterClientMockCreateBlueprint sync.RWMutex
	lockFilterClientMockGetJobState     sync.RWMutex
	lockFilterClientMockGetOutput       sync.RWMutex
	lockFilterClientMockUpdateBlueprint sync.RWMutex
)

// FilterClientMock is a mock implementation of FilterClient.
//
//     func TestSomethingThatUsesFilterClient(t *testing.T) {
//
//         // make and configure a mocked FilterClient
//         mockedFilterClient := &FilterClientMock{
//             CreateBlueprintFunc: func(instanceID string, names []string) (string, error) {
// 	               panic("TODO: mock out the CreateBlueprint method")
//             },
//             GetJobStateFunc: func(filterID string) (filter.Model, error) {
// 	               panic("TODO: mock out the GetJobState method")
//             },
//             GetOutputFunc: func(filterOutputID string) (filter.Model, error) {
// 	               panic("TODO: mock out the GetOutput method")
//             },
//             UpdateBlueprintFunc: func(m filter.Model, doSubmit bool) (filter.Model, error) {
// 	               panic("TODO: mock out the UpdateBlueprint method")
//             },
//         }
//
//         // TODO: use mockedFilterClient in code that requires FilterClient
//         //       and then make assertions.
//
//     }
type FilterClientMock struct {
	// CreateBlueprintFunc mocks the CreateBlueprint method.
	CreateBlueprintFunc func(instanceID string, names []string) (string, error)

	// GetJobStateFunc mocks the GetJobState method.
	GetJobStateFunc func(filterID string) (filter.Model, error)

	// GetOutputFunc mocks the GetOutput method.
	GetOutputFunc func(filterOutputID string) (filter.Model, error)

	// UpdateBlueprintFunc mocks the UpdateBlueprint method.
	UpdateBlueprintFunc func(m filter.Model, doSubmit bool) (filter.Model, error)

	// calls tracks calls to the methods.
	calls struct {
		// CreateBlueprint holds details about calls to the CreateBlueprint method.
		CreateBlueprint []struct {
			// InstanceID is the instanceID argument value.
			InstanceID string
			// Names is the names argument value.
			Names []string
		}
		// GetJobState holds details about calls to the GetJobState method.
		GetJobState []struct {
			// FilterID is the filterID argument value.
			FilterID string
		}
		// GetOutput holds details about calls to the GetOutput method.
		GetOutput []struct {
			// FilterOutputID is the filterOutputID argument value.
			FilterOutputID string
		}
		// UpdateBlueprint holds details about calls to the UpdateBlueprint method.
		UpdateBlueprint []struct {
			// M is the m argument value.
			M filter.Model
			// DoSubmit is the doSubmit argument value.
			DoSubmit bool
		}
	}
}

// CreateBlueprint calls CreateBlueprintFunc.
func (mock *FilterClientMock) CreateBlueprint(instanceID string, names []string) (string, error) {
	if mock.CreateBlueprintFunc == nil {
		panic("moq: FilterClientMock.CreateBlueprintFunc is nil but FilterClient.CreateBlueprint was just called")
	}
	callInfo := struct {
		InstanceID string
		Names      []string
	}{
		InstanceID: instanceID,
		Names:      names,
	}
	lockFilterClientMockCreateBlueprint.Lock()
	mock.calls.CreateBlueprint = append(mock.calls.CreateBlueprint, callInfo)
	lockFilterClientMockCreateBlueprint.Unlock()
	return mock.CreateBlueprintFunc(instanceID, names)
}

// CreateBlueprintCalls gets all the calls that were made to CreateBlueprint.
// Check the length with:
//     len(mockedFilterClient.CreateBlueprintCalls())
func (mock *FilterClientMock) CreateBlueprintCalls() []struct {
	InstanceID string
	Names      []string
} {
	var calls []struct {
		InstanceID string
		Names      []string
	}
	lockFilterClientMockCreateBlueprint.RLock()
	calls = mock.calls.CreateBlueprint
	lockFilterClientMockCreateBlueprint.RUnlock()
	return calls
}

// GetJobState calls GetJobStateFunc.
func (mock *FilterClientMock) GetJobState(filterID string) (filter.Model, error) {
	if mock.GetJobStateFunc == nil {
		panic("moq: FilterClientMock.GetJobStateFunc is nil but FilterClient.GetJobState was just called")
	}
	callInfo := struct {
		FilterID string
	}{
		FilterID: filterID,
	}
	lockFilterClientMockGetJobState.Lock()
	mock.calls.GetJobState = append(mock.calls.GetJobState, callInfo)
	lockFilterClientMockGetJobState.Unlock()
	return mock.GetJobStateFunc(filterID)
}

// GetJobStateCalls gets all the calls that were made to GetJobState.
// Check the length with:
//     len(mockedFilterClient.GetJobStateCalls())
func (mock *FilterClientMock) GetJobStateCalls() []struct {
	FilterID string
} {
	var calls []struct {
		FilterID string
	}
	lockFilterClientMockGetJobState.RLock()
	calls = mock.calls.GetJobState
	lockFilterClientMockGetJobState.RUnlock()
	return calls
}

// GetOutput calls GetOutputFunc.
func (mock *FilterClientMock) GetOutput(filterOutputID string) (filter.Model, error) {
	if mock.GetOutputFunc == nil {
		panic("moq: FilterClientMock.GetOutputFunc is nil but FilterClient.GetOutput was just called")
	}
	callInfo := struct {
		FilterOutputID string
	}{
		FilterOutputID: filterOutputID,
	}
	lockFilterClientMockGetOutput.Lock()
	mock.calls.GetOutput = append(mock.calls.GetOutput, callInfo)
	lockFilterClientMockGetOutput.Unlock()
	return mock.GetOutputFunc(filterOutputID)
}

// GetOutputCalls gets all the calls that were made to GetOutput.
// Check the length with:
//     len(mockedFilterClient.GetOutputCalls())
func (mock *FilterClientMock) GetOutputCalls() []struct {
	FilterOutputID string
} {
	var calls []struct {
		FilterOutputID string
	}
	lockFilterClientMockGetOutput.RLock()
	calls = mock.calls.GetOutput
	lockFilterClientMockGetOutput.RUnlock()
	return calls
}

// UpdateBlueprint calls UpdateBlueprintFunc.
func (mock *FilterClientMock) UpdateBlueprint(m filter.Model, doSubmit bool) (filter.Model, error) {
	if mock.UpdateBlueprintFunc == nil {
		panic("moq: FilterClientMock.UpdateBlueprintFunc is nil but FilterClient.UpdateBlueprint was just called")
	}
	callInfo := struct {
		M        filter.Model
		DoSubmit bool
	}{
		M:        m,
		DoSubmit: doSubmit,
	}
	lockFilterClientMockUpdateBlueprint.Lock()
	mock.calls.UpdateBlueprint = append(mock.calls.UpdateBlueprint, callInfo)
	lockFilterClientMockUpdateBlueprint.Unlock()
	return mock.UpdateBlueprintFunc(m, doSubmit)
}

// UpdateBlueprintCalls gets all the calls that were made to UpdateBlueprint.
// Check the length with:
//     len(mockedFilterClient.UpdateBlueprintCalls())
func (mock *FilterClientMock) UpdateBlueprintCalls() []struct {
	M        filter.Model
	DoSubmit bool
} {
	var calls []struct {
		M        filter.Model
		DoSubmit bool
	}
	lockFilterClientMockUpdateBlueprint.RLock()
	calls = mock.calls.UpdateBlueprint
	lockFilterClientMockUpdateBlueprint.RUnlock()
	return calls
}

var (
	lockStoreMockGetVersion    sync.RWMutex
	lockStoreMockUpdateVersion sync.RWMutex
)

// StoreMock is a mock implementation of Store.
//
//     func TestSomethingThatUsesStore(t *testing.T) {
//
//         // make and configure a mocked Store
//         mockedStore := &StoreMock{
//             GetVersionFunc: func(datasetID string, editionID string, version string, state string) (*models.Version, error) {
// 	               panic("TODO: mock out the GetVersion method")
//             },
//             UpdateVersionFunc: func(ID string, version *models.Version) error {
// 	               panic("TODO: mock out the UpdateVersion method")
//             },
//         }
//
//         // TODO: use mockedStore in code that requires Store
//         //       and then make assertions.
//
//     }
type StoreMock struct {
	// GetVersionFunc mocks the GetVersion method.
	GetVersionFunc func(datasetID string, editionID string, version string, state string) (*models.Version, error)

	// UpdateVersionFunc mocks the UpdateVersion method.
	UpdateVersionFunc func(ID string, version *models.Version) error

	// calls tracks calls to the methods.
	calls struct {
		// GetVersion holds details about calls to the GetVersion method.
		GetVersion []struct {
			// DatasetID is the datasetID argument value.
			DatasetID string
			// EditionID is the editionID argument value.
			EditionID string
			// Version is the version argument value.
			Version string
			// State is the state argument value.
			State string
		}
		// UpdateVersion holds details about calls to the UpdateVersion method.
		UpdateVersion []struct {
			// ID is the ID argument value.
			ID string
			// Version is the version argument value.
			Version *models.Version
		}
	}
}

// GetVersion calls GetVersionFunc.
func (mock *StoreMock) GetVersion(datasetID string, editionID string, version string, state string) (*models.Version, error) {
	if mock.GetVersionFunc == nil {
		panic("moq: StoreMock.GetVersionFunc is nil but Store.GetVersion was just called")
	}
	callInfo := struct {
		DatasetID string
		EditionID string
		Version   string
		State     string
	}{
		DatasetID: datasetID,
		EditionID: editionID,
		Version:   version,
		State:     state,
	}
	lockStoreMockGetVersion.Lock()
	mock.calls.GetVersion = append(mock.calls.GetVersion, callInfo)
	lockStoreMockGetVersion.Unlock()
	return mock.GetVersionFunc(datasetID, editionID, version, state)
}

// GetVersionCalls gets all the calls that were made to GetVersion.
// Check the length with:
//     len(mockedStore.GetVersionCalls())
func (mock *StoreMock) GetVersionCalls() []struct {
	DatasetID string
	EditionID string
	Version   string
	State     string
} {
	var calls []struct {
		DatasetID string
		EditionID string
		Version   string
		State     string
	}
	lockStoreMockGetVersion.RLock()
	calls = mock.calls.GetVersion
	lockStoreMockGetVersion.RUnlock()
	return calls
}

// UpdateVersion calls UpdateVersionFunc.
func (mock *StoreMock) UpdateVersion(ID string, version *models.Version) error {
	if mock.UpdateVersionFunc == nil {
		panic("moq: StoreMock.UpdateVersionFunc is nil but Store.UpdateVersion was just called")
	}
	callInfo := struct {
		ID      string
		Version *models.Version
	}{
		ID:      ID,
		Version: version,
	}
	lockStoreMockUpdateVersion.Lock()
	mock.calls.UpdateVersion = append(mock.calls.UpdateVersion, callInfo)
	lockStoreMockUpdateVersion.Unlock()
	return mock.UpdateVersionFunc(ID, version)
}

// UpdateVersionCalls gets all the calls that were made to UpdateVersion.
// Check the length with:
//     len(mockedStore.UpdateVersionCalls())
func (mock *StoreMock) UpdateVersionCalls() []struct {
	ID      string
	Version *models.Version
} {
	var calls []struct {
		ID      string
		Version *models.Version
	}
	lockStoreMockUpdateVersion.RLock()
	calls = mock.calls.UpdateVersion
	lockStoreMockUpdateVersion.RUnlock()
	return calls
}
