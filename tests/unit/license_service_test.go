package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"license-server/internal/domain/license"
	"license-server/internal/infrastructure/database"
)

// MockDatabase is a mock implementation of the database interface
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) AddLicense(ctx context.Context, req *database.AddLicenseRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockDatabase) GetLicense(ctx context.Context, key, product string) (*database.License, error) {
	args := m.Called(ctx, key, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.License), args.Error(1)
}

func (m *MockDatabase) UpdateLicense(ctx context.Context, req *database.UpdateLicenseRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockDatabase) DeleteLicense(ctx context.Context, key, product string) error {
	args := m.Called(ctx, key, product)
	return args.Error(0)
}

func (m *MockDatabase) ListLicenses(ctx context.Context, req *database.ListLicensesRequest) ([]*database.License, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*database.License), args.Error(1)
}

func (m *MockDatabase) GetAuditLogs(ctx context.Context, req *database.GetAuditLogsRequest) ([]*database.AuditLog, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*database.AuditLog), args.Error(1)
}

// TestLicenseService_CreateLicense tests license creation
func TestLicenseService_CreateLicense(t *testing.T) {
	tests := []struct {
		name           string
		request        *license.CreateLicenseRequest
		mockSetup      func(*MockDatabase)
		expectedError  string
		expectedResult *license.License
	}{
		{
			name: "successful license creation",
			request: &license.CreateLicenseRequest{
				Product:     "MyApp",
				OwnerEmail:  "user@example.com",
				OwnerName:   "John Doe",
				ExpiresIn:   "1y",
				Features:    []string{"premium", "api_access"},
				MaxActivations: 3,
			},
			mockSetup: func(m *MockDatabase) {
				m.On("AddLicense", mock.Anything, mock.AnythingOfType("*database.AddLicenseRequest")).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name: "invalid email format",
			request: &license.CreateLicenseRequest{
				Product:     "MyApp",
				OwnerEmail:  "invalid-email",
				OwnerName:   "John Doe",
				ExpiresIn:   "1y",
			},
			mockSetup:     func(m *MockDatabase) {},
			expectedError: "invalid email format",
		},
		{
			name: "empty product name",
			request: &license.CreateLicenseRequest{
				Product:     "",
				OwnerEmail:  "user@example.com",
				OwnerName:   "John Doe",
				ExpiresIn:   "1y",
			},
			mockSetup:     func(m *MockDatabase) {},
			expectedError: "product name is required",
		},
		{
			name: "database error",
			request: &license.CreateLicenseRequest{
				Product:     "MyApp",
				OwnerEmail:  "user@example.com",
				OwnerName:   "John Doe",
				ExpiresIn:   "1y",
			},
			mockSetup: func(m *MockDatabase) {
				m.On("AddLicense", mock.Anything, mock.AnythingOfType("*database.AddLicenseRequest")).
					Return(assert.AnError)
			},
			expectedError: "failed to create license",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDatabase)
			tt.mockSetup(mockDB)
			
			service := license.NewService(mockDB, nil, nil)
			
			// Execute
			result, err := service.CreateLicense(context.Background(), tt.request)
			
			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Product, result.Product)
				assert.Equal(t, tt.request.OwnerEmail, result.OwnerEmail)
				assert.Equal(t, tt.request.OwnerName, result.OwnerName)
			}
			
			mockDB.AssertExpectations(t)
		})
	}
}

// TestLicenseService_VerifyLicense tests license verification
func TestLicenseService_VerifyLicense(t *testing.T) {
	tests := []struct {
		name           string
		request        *license.VerifyLicenseRequest
		mockSetup      func(*MockDatabase)
		expectedError  string
		expectedResult *license.VerifyLicenseResponse
	}{
		{
			name: "valid license verification",
			request: &license.VerifyLicenseRequest{
				Key:      "ABCD-EFGH-IJKL-MNOP",
				Product:  "MyApp",
				Version:  "1.0.0",
				MachineID: "machine-123",
			},
			mockSetup: func(m *MockDatabase) {
				license := &database.License{
					ID:          "license-123",
					Key:         "ABCD-EFGH-IJKL-MNOP",
					Product:     "MyApp",
					OwnerEmail:  "user@example.com",
					OwnerName:   "John Doe",
					ExpiresAt:   time.Now().Add(365 * 24 * time.Hour),
					Features:    []string{"premium", "api_access"},
					MaxActivations: 3,
					CurrentActivations: 1,
					Status:      "active",
				}
				m.On("GetLicense", mock.Anything, "ABCD-EFGH-IJKL-MNOP", "MyApp").
					Return(license, nil)
			},
			expectedError: "",
			expectedResult: &license.VerifyLicenseResponse{
				Valid: true,
				License: &license.License{
					ID:          "license-123",
					Key:         "ABCD-EFGH-IJKL-MNOP",
					Product:     "MyApp",
					OwnerEmail:  "user@example.com",
					OwnerName:   "John Doe",
					ExpiresAt:   time.Now().Add(365 * 24 * time.Hour),
					Features:    []string{"premium", "api_access"},
					MaxActivations: 3,
					CurrentActivations: 1,
					Status:      "active",
				},
			},
		},
		{
			name: "license not found",
			request: &license.VerifyLicenseRequest{
				Key:     "INVALID-KEY",
				Product: "MyApp",
			},
			mockSetup: func(m *MockDatabase) {
				m.On("GetLicense", mock.Anything, "INVALID-KEY", "MyApp").
					Return(nil, database.ErrLicenseNotFound)
			},
			expectedError: "license not found",
			expectedResult: &license.VerifyLicenseResponse{
				Valid:  false,
				Reason: "license not found",
			},
		},
		{
			name: "expired license",
			request: &license.VerifyLicenseRequest{
				Key:     "ABCD-EFGH-IJKL-MNOP",
				Product: "MyApp",
			},
			mockSetup: func(m *MockDatabase) {
				expiredLicense := &database.License{
					ID:          "license-123",
					Key:         "ABCD-EFGH-IJKL-MNOP",
					Product:     "MyApp",
					OwnerEmail:  "user@example.com",
					OwnerName:   "John Doe",
					ExpiresAt:   time.Now().Add(-24 * time.Hour), // Expired yesterday
					Status:      "expired",
				}
				m.On("GetLicense", mock.Anything, "ABCD-EFGH-IJKL-MNOP", "MyApp").
					Return(expiredLicense, nil)
			},
			expectedError: "",
			expectedResult: &license.VerifyLicenseResponse{
				Valid:  false,
				Reason: "license expired",
			},
		},
		{
			name: "max activations exceeded",
			request: &license.VerifyLicenseRequest{
				Key:      "ABCD-EFGH-IJKL-MNOP",
				Product:  "MyApp",
				MachineID: "new-machine-456",
			},
			mockSetup: func(m *MockDatabase) {
				maxedLicense := &database.License{
					ID:          "license-123",
					Key:         "ABCD-EFGH-IJKL-MNOP",
					Product:     "MyApp",
					OwnerEmail:  "user@example.com",
					OwnerName:   "John Doe",
					ExpiresAt:   time.Now().Add(365 * 24 * time.Hour),
					MaxActivations: 2,
					CurrentActivations: 2, // Already at max
					Status:      "active",
				}
				m.On("GetLicense", mock.Anything, "ABCD-EFGH-IJKL-MNOP", "MyApp").
					Return(maxedLicense, nil)
			},
			expectedError: "",
			expectedResult: &license.VerifyLicenseResponse{
				Valid:  false,
				Reason: "maximum activations exceeded",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDatabase)
			tt.mockSetup(mockDB)
			
			service := license.NewService(mockDB, nil, nil)
			
			// Execute
			result, err := service.VerifyLicense(context.Background(), tt.request)
			
			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Valid, result.Valid)
				if tt.expectedResult.Reason != "" {
					assert.Equal(t, tt.expectedResult.Reason, result.Reason)
				}
				if tt.expectedResult.License != nil {
					assert.Equal(t, tt.expectedResult.License.Key, result.License.Key)
					assert.Equal(t, tt.expectedResult.License.Product, result.License.Product)
				}
			}
			
			mockDB.AssertExpectations(t)
		})
	}
}

// TestLicenseService_ListLicenses tests license listing
func TestLicenseService_ListLicenses(t *testing.T) {
	tests := []struct {
		name           string
		request        *license.ListLicensesRequest
		mockSetup      func(*MockDatabase)
		expectedError  string
		expectedCount  int
	}{
		{
			name: "successful license listing",
			request: &license.ListLicensesRequest{
				Page:   1,
				Limit:  10,
				Product: "MyApp",
			},
			mockSetup: func(m *MockDatabase) {
				licenses := []*database.License{
					{
						ID:          "license-1",
						Key:         "ABCD-EFGH-IJKL-MNOP",
						Product:     "MyApp",
						OwnerEmail:  "user1@example.com",
						OwnerName:   "John Doe",
						Status:      "active",
					},
					{
						ID:          "license-2",
						Key:         "WXYZ-1234-5678-9ABC",
						Product:     "MyApp",
						OwnerEmail:  "user2@example.com",
						OwnerName:   "Jane Smith",
						Status:      "active",
					},
				}
				m.On("ListLicenses", mock.Anything, mock.AnythingOfType("*database.ListLicensesRequest")).
					Return(licenses, nil)
			},
			expectedError: "",
			expectedCount: 2,
		},
		{
			name: "empty result",
			request: &license.ListLicensesRequest{
				Page:   1,
				Limit:  10,
				Product: "NonExistentApp",
			},
			mockSetup: func(m *MockDatabase) {
				m.On("ListLicenses", mock.Anything, mock.AnythingOfType("*database.ListLicensesRequest")).
					Return([]*database.License{}, nil)
			},
			expectedError: "",
			expectedCount: 0,
		},
		{
			name: "database error",
			request: &license.ListLicensesRequest{
				Page:   1,
				Limit:  10,
			},
			mockSetup: func(m *MockDatabase) {
				m.On("ListLicenses", mock.Anything, mock.AnythingOfType("*database.ListLicensesRequest")).
					Return(nil, assert.AnError)
			},
			expectedError: "failed to list licenses",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDatabase)
			tt.mockSetup(mockDB)
			
			service := license.NewService(mockDB, nil, nil)
			
			// Execute
			result, err := service.ListLicenses(context.Background(), tt.request)
			
			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Licenses, tt.expectedCount)
			}
			
			mockDB.AssertExpectations(t)
		})
	}
}

// TestLicenseService_DeleteLicense tests license deletion
func TestLicenseService_DeleteLicense(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		product       string
		mockSetup     func(*MockDatabase)
		expectedError string
	}{
		{
			name:    "successful license deletion",
			key:     "ABCD-EFGH-IJKL-MNOP",
			product: "MyApp",
			mockSetup: func(m *MockDatabase) {
				m.On("DeleteLicense", mock.Anything, "ABCD-EFGH-IJKL-MNOP", "MyApp").
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:    "license not found",
			key:     "INVALID-KEY",
			product: "MyApp",
			mockSetup: func(m *MockDatabase) {
				m.On("DeleteLicense", mock.Anything, "INVALID-KEY", "MyApp").
					Return(database.ErrLicenseNotFound)
			},
			expectedError: "license not found",
		},
		{
			name:    "database error",
			key:     "ABCD-EFGH-IJKL-MNOP",
			product: "MyApp",
			mockSetup: func(m *MockDatabase) {
				m.On("DeleteLicense", mock.Anything, "ABCD-EFGH-IJKL-MNOP", "MyApp").
					Return(assert.AnError)
			},
			expectedError: "failed to delete license",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDatabase)
			tt.mockSetup(mockDB)
			
			service := license.NewService(mockDB, nil, nil)
			
			// Execute
			err := service.DeleteLicense(context.Background(), tt.key, tt.product)
			
			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
			
			mockDB.AssertExpectations(t)
		})
	}
}

// Benchmark tests
func BenchmarkLicenseService_VerifyLicense(b *testing.B) {
	mockDB := new(MockDatabase)
	license := &database.License{
		ID:          "license-123",
		Key:         "ABCD-EFGH-IJKL-MNOP",
		Product:     "MyApp",
		OwnerEmail:  "user@example.com",
		OwnerName:   "John Doe",
		ExpiresAt:   time.Now().Add(365 * 24 * time.Hour),
		Status:      "active",
	}
	mockDB.On("GetLicense", mock.Anything, "ABCD-EFGH-IJKL-MNOP", "MyApp").
		Return(license, nil)
	
	service := license.NewService(mockDB, nil, nil)
	request := &license.VerifyLicenseRequest{
		Key:     "ABCD-EFGH-IJKL-MNOP",
		Product: "MyApp",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.VerifyLicense(context.Background(), request)
	}
}

// Test helper functions
func TestValidateLicenseKey(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"ABCD-EFGH-IJKL-MNOP", true},
		{"1234-5678-9ABC-DEF0", true},
		{"invalid-key", false},
		{"ABCD-EFGH-IJKL", false},
		{"ABCD-EFGH-IJKL-MNOP-EXTRA", false},
		{"", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := license.ValidateLicenseKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"user@example.com", true},
		{"test.email+tag@domain.co.uk", true},
		{"invalid-email", false},
		{"@domain.com", false},
		{"user@", false},
		{"", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := license.ValidateEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}