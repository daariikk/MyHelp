package postgres

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/daariikk/MyHelp/services/api-gateway/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Моки ---
type MockPgxConn struct{ mock.Mock }
type MockPgxRow struct{ mock.Mock }
type MockPgxRows struct{ mock.Mock }

func (m *MockPgxConn) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	a := m.Called(ctx, sql, args)
	return a.Get(0).(pgx.Row)
}
func (m *MockPgxConn) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	a := m.Called(ctx, sql, args)
	return a.Get(0).(pgx.Rows), a.Error(1)
}
func (m *MockPgxConn) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	a := m.Called(ctx, sql, args)
	return a.Get(0).(pgconn.CommandTag), a.Error(1)
}
func (m *MockPgxConn) Close(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
func (m *MockPgxRow) Scan(dest ...interface{}) error {
	args := m.Called(dest)
	return args.Error(0)
}
func (m *MockPgxRows) Next() bool { return m.Called().Bool(0) }
func (m *MockPgxRows) Scan(dest ...interface{}) error {
	args := m.Called(dest)
	return args.Error(0)
}
func (m *MockPgxRows) Close()     { m.Called() }
func (m *MockPgxRows) Err() error { return m.Called().Error(0) }

func (m *MockPgxRows) CommandTag() pgconn.CommandTag                { return m.Called().Get(0).(pgconn.CommandTag) }
func (m *MockPgxRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *MockPgxRows) Values() ([]interface{}, error)               { return nil, nil }
func (m *MockPgxRows) RawValues() [][]byte                          { return nil }
func (m *MockPgxRows) Conn() *pgx.Conn                              { return nil }

func TestStorage_RegisterUser(t *testing.T) {
	type args struct {
		user domain.User
	}
	tests := []struct {
		name      string
		setupMock func(*MockPgxConn, *MockPgxRow)
		args      args
		wantErr   bool
		wantId    int
	}{
		{
			name: "success",
			args: args{user: domain.User{
				Surname: "A", Name: "B", Patronymic: "C", Polic: "P", Email: "x@y", Password: "pwd",
			}},
			setupMock: func(conn *MockPgxConn, row *MockPgxRow) {

				rows := new(MockPgxRows)
				rows.On("Next").Return(false)
				rows.On("Close").Return()
				conn.On("Query", context.Background(), mock.Anything, []interface{}{"x@y"}).
					Return(rows, nil)
				conn.On("QueryRow", context.Background(), mock.Anything, mock.Anything).
					Return(row)
				row.On("Scan", mock.AnythingOfType("[]interface {}")).
					Run(func(args mock.Arguments) {
						dest := args.Get(0).([]interface{})
						*(dest[0].(*int)) = 42
					}).Return(nil)
			},

			wantErr: false,
			wantId:  42,
		},
		{
			name: "user already exists",
			args: args{user: domain.User{Email: "x@y"}},
			setupMock: func(conn *MockPgxConn, row *MockPgxRow) {
				mRows := new(MockPgxRows)
				mRows.On("Next").Return(true)
				mRows.On("Close").Return(nil)

				conn.On("Query", context.Background(), mock.Anything, []interface{}{"x@y"}).
					Return(mRows, nil)
			},
			wantErr: true,
		},
		{
			name: "db error",
			args: args{user: domain.User{Surname: "A", Email: "x@y", Password: "pwd"}},
			setupMock: func(conn *MockPgxConn, row *MockPgxRow) {
				mRows := new(MockPgxRows)
				mRows.On("Next").Return(false)
				mRows.On("Close").Return(nil)

				conn.On("Query", context.Background(), mock.Anything, []interface{}{"x@y"}).
					Return(mRows, nil)

				conn.On("QueryRow", context.Background(), mock.Anything, mock.Anything).
					Return(row)
				row.On("Scan", mock.AnythingOfType("[]interface {}")).
					Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := new(MockPgxConn)
			row := new(MockPgxRow)
			if tt.setupMock != nil {
				tt.setupMock(conn, row)
			}
			st := &Storage{connection: conn, logger: slog.New(slog.NewTextHandler(nil, nil))}

			u, err := st.RegisterUser(tt.args.user)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantId, u.Id)
			}
		})
	}
}

func TestStorage_GetPassword(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockPgxConn, *MockPgxRows)
		email     string
		wantId    int
		wantPwd   string
		wantErr   bool
	}{
		{
			name:  "success",
			email: "x@y",
			setupMock: func(conn *MockPgxConn, rows *MockPgxRows) {
				conn.On("Query", context.Background(), mock.Anything, []interface{}{"x@y"}).
					Return(rows, nil)
				rows.On("Next").Return(true)
				rows.On("Scan", mock.AnythingOfType("[]interface {}")).Run(func(args mock.Arguments) {
					dest := args.Get(0).([]interface{})
					*(dest[0].(*int)) = 2
					*(dest[1].(*string)) = "pwd"
				}).Return(nil)
				rows.On("Close").Return()
			},
			wantId: 2, wantPwd: "pwd",
		},
		{
			name:  "user not found",
			email: "x@y",
			setupMock: func(conn *MockPgxConn, rows *MockPgxRows) {
				conn.On("Query", context.Background(), mock.Anything, []interface{}{"x@y"}).
					Return(rows, nil)
				rows.On("Next").Return(false)
				rows.On("Close").Return()
			},
			wantErr: true,
		},
		{
			name:  "db error",
			email: "x@y",
			setupMock: func(conn *MockPgxConn, rows *MockPgxRows) {
				conn.On("Query", context.Background(), mock.Anything, []interface{}{"x@y"}).
					Return(rows, errors.New("fail"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := new(MockPgxConn)
			rows := new(MockPgxRows)
			tt.setupMock(conn, rows)
			st := &Storage{connection: conn, logger: slog.New(slog.NewTextHandler(nil, nil))}
			id, pwd, err := st.GetPassword(tt.email)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantId, id)
				assert.Equal(t, tt.wantPwd, pwd)
			}
		})
	}
}

func TestStorage_UpdatePassword(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockPgxConn)
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func(conn *MockPgxConn) {
				conn.On("Exec", context.Background(), mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)
			},
			wantErr: false,
		},
		{
			name: "fail",
			setupMock: func(conn *MockPgxConn) {
				conn.On("Exec", context.Background(), mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, errors.New("fail"))
			},
			wantErr: true,
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := new(MockPgxConn)
			tt.setupMock(conn)
			st := &Storage{
				connection: conn,
				logger:     logger,
			}
			err := st.UpdatePassword("x@y", "pwd")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStorage_CheckUserByEmail(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockPgxConn, *MockPgxRows)
		wantFound bool
		wantErr   bool
	}{
		{
			name: "found",
			setupMock: func(conn *MockPgxConn, rows *MockPgxRows) {
				conn.On("Query", context.Background(), mock.Anything, mock.Anything).Return(rows, nil)
				rows.On("Next").Return(true)
				rows.On("Close").Return()
			},
			wantFound: true,
		},
		{
			name: "not found",
			setupMock: func(conn *MockPgxConn, rows *MockPgxRows) {
				conn.On("Query", context.Background(), mock.Anything, mock.Anything).Return(rows, nil)
				rows.On("Next").Return(false)
				rows.On("Close").Return()
			},
			wantFound: false,
		},
		{
			name: "fail",
			setupMock: func(conn *MockPgxConn, rows *MockPgxRows) {
				conn.On("Query", context.Background(), mock.Anything, mock.Anything).Return(rows, errors.New("fail"))
			},
			wantErr: true,
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := new(MockPgxConn)
			rows := new(MockPgxRows)
			tt.setupMock(conn, rows)
			st := &Storage{
				connection: conn,
				logger:     logger,
			}

			found, err := st.CheckUserByEmail("x@y")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantFound, found)
			}
		})
	}
}
