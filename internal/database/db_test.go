package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocbotau/api-join-sound/internal/database"
)

func TestNewSQLiteDB(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		dataSource string
		wantErr    bool
	}{
		{
			name:       "In-memory SQLite DB",
			dataSource: ":memory:",
			wantErr:    false,
		},
		{
			name:       "Invalid data source",
			dataSource: "/invalid/path/to/db.sqlite",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, err := database.NewSQLiteDB(tt.dataSource)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, db)
			} else {
				require.NoError(t, err)
				require.NotNil(t, db)
			}
		})
	}
}
