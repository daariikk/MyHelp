package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/daariikk/MyHelp/services/account-service/internal/domain"
	"github.com/daariikk/MyHelp/services/account-service/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockPgxConn is a mock implementation of pgx.Conn
type MockPgxConn struct {
	mock.Mock
}

func (m *MockPgxConn) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	arguments := m.Called(ctx, sql, args)
	return arguments.Get(0).(pgx.Row)
}

func (m *MockPgxConn) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	arguments := m.Called(ctx, sql, args)
	return arguments.Get(0).(pgx.Rows), arguments.Error(1)
}

func (m *MockPgxConn) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	arguments := m.Called(ctx, sql, args)
	return arguments.Get(0).(pgconn.CommandTag), arguments.Error(1)
}

func (m *MockPgxConn) Close(ctx context.Context) error {
	arguments := m.Called(ctx)
	return arguments.Error(0)
}

// MockPgxRow is a mock implementation of pgx.Row
type MockPgxRow struct {
	mock.Mock
}

func (m *MockPgxRow) Scan(dest ...interface{}) error {
	args := m.Called(dest)
	return args.Error(0)
}

// MockPgxRows is a mock implementation of pgx.Rows
type MockPgxRows struct {
	mock.Mock
	rows []domain.Appointment
	idx  int
}

func (m *MockPgxRows) Close() {
	m.Called()
}

func (m *MockPgxRows) Err() error {
	return m.Called().Error(0)
}

func (m *MockPgxRows) CommandTag() pgconn.CommandTag {
	return m.Called().Get(0).(pgconn.CommandTag)
}

func (m *MockPgxRows) FieldDescriptions() []pgconn.FieldDescription {
	return m.Called().Get(0).([]pgconn.FieldDescription)
}

func (m *MockPgxRows) Next() bool {
	m.idx++
	return m.idx <= len(m.rows)
}

func (m *MockPgxRows) Scan(dest ...interface{}) error {
	if m.idx > len(m.rows) {
		return errors.New("no more rows")
	}

	app := m.rows[m.idx-1]

	if len(dest) >= 1 {
		*(dest[0].(*int)) = app.Id
	}
	if len(dest) >= 2 {
		*(dest[1].(*string)) = app.DoctorFIO
	}
	if len(dest) >= 3 {
		*(dest[2].(*string)) = app.DoctorSpecialization
	}
	if len(dest) >= 4 {
		*(dest[3].(*time.Time)) = app.Date
	}
	if len(dest) >= 5 {
		*(dest[4].(*time.Time)) = app.Time
	}
	if len(dest) >= 6 {
		*(dest[5].(*string)) = app.Status
	}
	if len(dest) >= 7 {
		if app.Rating != 5.0 {
			*(dest[6].(*sql.NullFloat64)) = sql.NullFloat64{Float64: app.Rating, Valid: true}
		} else {
			*(dest[6].(*sql.NullFloat64)) = sql.NullFloat64{Valid: false}
		}
	}

	return nil
}

func (m *MockPgxRows) Values() ([]interface{}, error) {
	return m.Called().Get(0).([]interface{}), m.Called().Error(1)
}

func (m *MockPgxRows) RawValues() [][]byte {
	return m.Called().Get(0).([][]byte)
}

func (m *MockPgxRows) Conn() *pgx.Conn {
	return m.Called().Get(0).(*pgx.Conn)
}

func TestStorage_GetPatientById(t *testing.T) {
	tests := []struct {
		name        string
		patientID   int
		mockSetup   func(*MockPgxConn, *MockPgxRow)
		expected    domain.Patient
		expectedErr error
	}{
		{
			name:      "successful retrieval",
			patientID: 1,
			mockSetup: func(conn *MockPgxConn, row *MockPgxRow) {
				conn.On("QueryRow", context.Background(),
					`SELECT id, surname, name, patronymic, email, polic, is_deleted FROM patients WHERE id=$1`,
					[]interface{}{1}).
					Return(row)

				row.On("Scan", mock.AnythingOfType("[]interface {}")).
					Run(func(args mock.Arguments) {
						dest := args.Get(0).([]interface{})
						*(dest[0].(*int)) = 1
						*(dest[1].(*sql.NullString)) = sql.NullString{String: "Doe", Valid: true}
						*(dest[2].(*sql.NullString)) = sql.NullString{String: "John", Valid: true}
						*(dest[3].(*sql.NullString)) = sql.NullString{String: "Smith", Valid: true}
						*(dest[4].(*string)) = "john.doe@example.com"
						*(dest[5].(*string)) = "123456789"
						*(dest[6].(*bool)) = false
					}).
					Return(nil)
			},
			expected: domain.Patient{
				Id:         1,
				Surname:    "Doe",
				Name:       "John",
				Patronymic: "Smith",
				Email:      "john.doe@example.com",
				Polic:      "123456789",
				IsDeleted:  false,
			},
			expectedErr: nil,
		},
		{
			name:      "patient not found",
			patientID: 2,
			mockSetup: func(conn *MockPgxConn, row *MockPgxRow) {
				conn.On("QueryRow", context.Background(),
					"SELECT id, surname, name, patronymic, email, polic, is_deleted FROM patients WHERE id=$1",
					[]interface{}{2}).
					Return(row)

				row.On("Scan", mock.AnythingOfType("[]interface {}")).
					Return(pgx.ErrNoRows)
			},
			expected:    domain.Patient{},
			expectedErr: fmt.Errorf("%w: patient with id %d not found", repository.ErrorNotFound, 2),
		},
		{
			name:      "database error",
			patientID: 3,
			mockSetup: func(conn *MockPgxConn, row *MockPgxRow) {
				conn.On("QueryRow", context.Background(),
					"SELECT id, surname, name, patronymic, email, polic, is_deleted FROM patients WHERE id=$1",
					[]interface{}{3}).
					Return(row)

				row.On("Scan", mock.AnythingOfType("[]interface {}")).
					Return(errors.New("database error"))
			},
			expected:    domain.Patient{},
			expectedErr: fmt.Errorf("failed to get patient with id %d: database error", 3),
		},
		{
			name:      "null values handling",
			patientID: 4,
			mockSetup: func(conn *MockPgxConn, row *MockPgxRow) {
				conn.On("QueryRow", context.Background(),
					"SELECT id, surname, name, patronymic, email, polic, is_deleted FROM patients WHERE id=$1",
					[]interface{}{4}).
					Return(row)

				row.On("Scan", mock.AnythingOfType("[]interface {}")).
					Run(func(args mock.Arguments) {
						dest := args.Get(0).([]interface{})
						*(dest[0].(*int)) = 4
						*(dest[1].(*sql.NullString)) = sql.NullString{Valid: false}
						*(dest[2].(*sql.NullString)) = sql.NullString{Valid: false}
						*(dest[3].(*sql.NullString)) = sql.NullString{Valid: false}
						*(dest[4].(*string)) = "no.name@example.com"
						*(dest[5].(*string)) = "987654321"
						*(dest[6].(*bool)) = false
					}).
					Return(nil)
			},
			expected: domain.Patient{
				Id:         4,
				Surname:    "",
				Name:       "",
				Patronymic: "",
				Email:      "no.name@example.com",
				Polic:      "987654321",
				IsDeleted:  false,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := new(MockPgxConn)
			mockRow := new(MockPgxRow)

			tt.mockSetup(mockConn, mockRow)

			logger := slog.New(slog.NewTextHandler(nil, nil))
			storage := &Storage{
				connection: mockConn,
				logger:     logger,
			}

			patient, err := storage.GetPatientById(tt.patientID)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, patient)
			}

			mockConn.AssertExpectations(t)
			mockRow.AssertExpectations(t)
		})
	}
}

func TestStorage_GetAppointmentByPatientId(t *testing.T) {
	date1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	time1 := time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)
	date2 := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)
	time2 := time.Date(0, 1, 1, 11, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		patientID   int
		mockSetup   func(*MockPgxConn, *MockPgxRows)
		expected    []domain.Appointment
		expectedErr error
	}{
		{
			name:      "successful retrieval with appointments",
			patientID: 1,
			mockSetup: func(conn *MockPgxConn, rows *MockPgxRows) {
				// Setup for the update query
				conn.On("Exec", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{1, domain.COMPLETED}).
					Return(pgconn.CommandTag{}, nil)

				// Setup for the select query
				conn.On("Query", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{1}).
					Return(rows, nil)

				rows.rows = []domain.Appointment{
					{
						Id:                   1,
						DoctorFIO:            "Dr. Smith John",
						DoctorSpecialization: "Cardiology",
						Date:                 date1,
						Time:                 time1,
						Status:               "SCHEDULED",
						Rating:               5.0,
					},
					{
						Id:                   2,
						DoctorFIO:            "Dr. Johnson Alice",
						DoctorSpecialization: "Neurology",
						Date:                 date2,
						Time:                 time2,
						Status:               "COMPLETED",
						Rating:               4.5,
					},
				}

				rows.On("Close").Return()
				rows.On("Err").Return(nil)
			},
			expected: []domain.Appointment{
				{
					Id:                   1,
					DoctorFIO:            "Dr. Smith John",
					DoctorSpecialization: "Cardiology",
					Date:                 date1,
					Time:                 time1,
					Status:               "SCHEDULED",
					Rating:               5.0,
				},
				{
					Id:                   2,
					DoctorFIO:            "Dr. Johnson Alice",
					DoctorSpecialization: "Neurology",
					Date:                 date2,
					Time:                 time2,
					Status:               "COMPLETED",
					Rating:               4.5,
				},
			},
			expectedErr: nil,
		},
		{
			name:      "no appointments found",
			patientID: 2,
			mockSetup: func(conn *MockPgxConn, rows *MockPgxRows) {
				conn.On("Exec", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{2, domain.COMPLETED}).
					Return(pgconn.CommandTag{}, nil)

				conn.On("Query", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{2}).
					Return(rows, nil)

				rows.rows = []domain.Appointment{}
				rows.On("Close").Return()
				rows.On("Err").Return(nil)
			},
			expected:    []domain.Appointment{},
			expectedErr: nil,
		},
		{
			name:      "update query fails",
			patientID: 3,
			mockSetup: func(conn *MockPgxConn, rows *MockPgxRows) {
				conn.On("Exec", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{3, domain.COMPLETED}).
					Return(pgconn.CommandTag{}, errors.New("update failed"))
			},
			expected:    nil,
			expectedErr: fmt.Errorf("failed to update appointment statuses: update failed"),
		},
		{
			name:      "select query fails",
			patientID: 4,
			mockSetup: func(conn *MockPgxConn, rows *MockPgxRows) {
				conn.On("Exec", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{4, domain.COMPLETED}).
					Return(pgconn.CommandTag{}, nil)

				conn.On("Query", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{4}).
					Return(rows, errors.New("query failed"))
			},
			expected:    nil,
			expectedErr: fmt.Errorf("failed to execute query: query failed"),
		},
		{
			name:      "scan error",
			patientID: 5,
			mockSetup: func(conn *MockPgxConn, rows *MockPgxRows) {
				conn.On("Exec", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{5, domain.COMPLETED}).
					Return(pgconn.CommandTag{}, nil)

				conn.On("Query", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{5}).
					Return(rows, nil)

				rows.rows = []domain.Appointment{
					{
						Id:                   1,
						DoctorFIO:            "Dr. Smith John",
						DoctorSpecialization: "Cardiology",
						Date:                 date1,
						Time:                 time1,
						Status:               "SCHEDULED",
						Rating:               5.0,
					},
				}

				rows.On("Close").Return()
				rows.On("Err").Return(errors.New("scan error"))
			},
			expected:    nil,
			expectedErr: fmt.Errorf("error iterating rows: scan error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := new(MockPgxConn)
			mockRows := &MockPgxRows{}

			tt.mockSetup(mockConn, mockRows)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			storage := &Storage{
				connection: mockConn,
				logger:     logger,
			}

			appointments, err := storage.GetAppointmentByPatientId(tt.patientID)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, appointments)
			}

			mockConn.AssertExpectations(t)
			mockRows.AssertExpectations(t)
		})
	}
}

func TestStorage_Close(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*MockPgxConn)
		expectedErr error
	}{
		{
			name: "successful close",
			mockSetup: func(conn *MockPgxConn) {
				conn.On("Close", context.Background()).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "close error",
			mockSetup: func(conn *MockPgxConn) {
				conn.On("Close", context.Background()).Return(errors.New("close failed"))
			},
			expectedErr: errors.New("close failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockConn *MockPgxConn
			if tt.name != "nil connection" {
				mockConn = new(MockPgxConn)
				tt.mockSetup(mockConn)
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			storage := &Storage{
				connection: mockConn,
				logger:     logger,
			}

			err := storage.Close()

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			if mockConn != nil {
				mockConn.AssertExpectations(t)
			}
		})
	}
}

func TestStorage_DeletePatientById(t *testing.T) {
	tests := []struct {
		name        string
		patientID   int
		mockSetup   func(*MockPgxConn, *MockPgxRow)
		expected    bool
		expectedErr error
	}{
		{
			name:      "successful deletion",
			patientID: 1,
			mockSetup: func(conn *MockPgxConn, row *MockPgxRow) {
				conn.On("Exec", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{1}).
					Return(pgconn.CommandTag{}, nil)

				conn.On("QueryRow", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{1}).
					Return(row)

				row.On("Scan", mock.AnythingOfType("[]interface {}")).
					Run(func(args mock.Arguments) {
						dest := args.Get(0).([]interface{})
						*(dest[0].(*bool)) = true
					}).
					Return(nil)
			},
			expected:    true,
			expectedErr: nil,
		},
		{
			name:      "delete query fails",
			patientID: 2,
			mockSetup: func(conn *MockPgxConn, row *MockPgxRow) {
				conn.On("Exec", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{2}).
					Return(pgconn.CommandTag{}, errors.New("update failed"))
			},
			expected:    false,
			expectedErr: errors.New("update failed"),
		},
		{
			name:      "select query returns no rows",
			patientID: 3,
			mockSetup: func(conn *MockPgxConn, row *MockPgxRow) {
				conn.On("Exec", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{3}).
					Return(pgconn.CommandTag{}, nil)

				conn.On("QueryRow", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{3}).
					Return(row)

				row.On("Scan", mock.AnythingOfType("[]interface {}")).
					Return(pgx.ErrNoRows)
			},
			expected:    false,
			expectedErr: nil,
		},
		{
			name:      "select query returns error",
			patientID: 4,
			mockSetup: func(conn *MockPgxConn, row *MockPgxRow) {
				conn.On("Exec", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{4}).
					Return(pgconn.CommandTag{}, nil)

				conn.On("QueryRow", context.Background(),
					mock.AnythingOfType("string"),
					[]interface{}{4}).
					Return(row)

				row.On("Scan", mock.AnythingOfType("[]interface {}")).
					Return(errors.New("db error"))
			},
			expected:    false,
			expectedErr: errors.New("failed to get patient with id 4: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := new(MockPgxConn)
			mockRow := new(MockPgxRow)

			tt.mockSetup(mockConn, mockRow)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			storage := &Storage{
				connection: mockConn,
				logger:     logger,
			}

			isDeleted, err := storage.DeletePatientById(tt.patientID)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, isDeleted)
			}

			mockConn.AssertExpectations(t)
			mockRow.AssertExpectations(t)
		})
	}
}
