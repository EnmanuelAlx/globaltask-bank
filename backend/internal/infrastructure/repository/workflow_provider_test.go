package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRows implements pgx.Rows for testing
type MockRows struct {
	mock.Mock
	currentIndex int
	data         [][]interface{}
}

func (m *MockRows) Close()                                       { m.Called() }
func (m *MockRows) Err() error                                   { return m.Called().Error(0) }
func (m *MockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *MockRows) Next() bool {
	if m.currentIndex < len(m.data) {
		return true
	}
	return false
}
func (m *MockRows) Scan(dest ...any) error {
	row := m.data[m.currentIndex]
	for i, val := range row {
		if val == nil {
			continue
		}
		// Simplified scan for the test
		switch d := dest[i].(type) {
		case *int:
			*d = val.(int)
		}
	}
	m.currentIndex++
	return nil
}
func (m *MockRows) Values() ([]any, error) { return nil, nil }
func (m *MockRows) RawValues() [][]byte    { return nil }
func (m *MockRows) Conn() *pgx.Conn        { return nil }

// MockDB implements DBTX for testing
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	callArgs := m.Called(ctx, sql, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(pgx.Rows), callArgs.Error(1)
}

func (m *MockDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	callArgs := m.Called(ctx, sql, args)
	return callArgs.Get(0).(pgx.Row)
}

func TestWorkflowProviderRepository_GetByWorkflowAndStep_SQL(t *testing.T) {
	db := new(MockDB)
	repo := NewWorkflowProviderRepository(db)
	ctx := context.Background()

	// Capture the SQL query
	db.On("Query", ctx, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		query := args.String(1)
		assert.Contains(t, query, "priority")
		assert.Contains(t, query, "ORDER BY priority ASC")
	}).Return(nil, pgx.ErrNoRows) // We return error to stop scanning

	_, _ = repo.GetByWorkflowAndStep(ctx, "PT", "FETCH_BANK_DATA")

	db.AssertExpectations(t)
}
