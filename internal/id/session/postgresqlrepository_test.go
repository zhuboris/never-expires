package session

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhuboris/never-expires/internal/id/usr"
	"github.com/zhuboris/never-expires/internal/shared/postgresql"
	"github.com/zhuboris/never-expires/internal/test"
)

func TestNewPostgresqlRepository(t *testing.T) {
	config := test.PostgresConfig{
		Username: "postgres",
		Password: "12345",
		Host:     "localhost",
		Port:     5432,
		DBName:   "test" + strconv.Itoa(rand.Int()),
	}

	pool := test.SetupDatabase(t, config, "")

	tests := []struct {
		name          string
		pool          *pgxpool.Pool
		requireError  require.ErrorAssertionFunc
		expectedError error
	}{
		{
			name:          "pool is nil",
			pool:          nil,
			requireError:  require.Error,
			expectedError: postgresql.ErrPoolInitRequired,
		},
		{
			name:         "pool is valid",
			pool:         pool,
			requireError: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewPostgresqlRepository(tt.pool)

			tt.requireError(t, err)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			}

			if err == nil { // if NO error
				err := repo.pool.Ping(context.Background())
				assert.NoError(t, err, "not connected to db")
			}
		})
	}
}

func TestPostgresqlRepository_Ping(t *testing.T) {
	repo := arrangeRepoWithTestDB(t)

	err := repo.Ping(context.Background())

	assert.NoError(t, err, "no connection")
}

func TestPostgresqlRepository_add(t *testing.T) {
	const (
		arrangeQuery = `
			WITH test_user1 AS (
			    INSERT INTO users (id, username) 
			    VALUES ('c171f212-6520-4a34-8761-c450b46cf853', 'user1')
			), test_user2 AS (
			     INSERT INTO users (username) 
			    VALUES ('user2')
			    
			    RETURNING id
			)
			INSERT INTO sessions (id, user_id, refresh_jwt)
			VALUES ('b14f265a-7317-42fd-8d14-b7862bf3cb37', (SELECT id FROM test_user2), 'existing');
		`
		checkResultQuery = `
			SELECT id, user_id, device, refresh_jwt, is_active
			FROM sessions
			WHERE id = $1;
		`
	)

	var (
		existingUserID    = stringToUUID(t, "c171f212-6520-4a34-8761-c450b46cf853")
		notExistingUserID = stringToUUID(t, "916cdb99-fad2-4e17-b067-0828b6bf9311")

		existingRefreshToken = "existing"
		newRefreshToken      = "new"
	)

	tests := []struct {
		name         string
		session      Session
		requireError require.ErrorAssertionFunc
	}{
		{
			name: "session with invalid user id",
			session: Session{
				UserID:     pgtype.UUID{Valid: false},
				Device:     "device",
				RefreshJWT: newRefreshToken,
			},
			requireError: require.Error,
		},
		{
			name: "session with not existing user id",
			session: Session{
				UserID:     notExistingUserID,
				Device:     "device",
				RefreshJWT: newRefreshToken,
			},
			requireError: require.Error,
		},
		{
			name: "session with already existing refresh token",
			session: Session{
				UserID:     existingUserID,
				Device:     "device",
				RefreshJWT: existingRefreshToken,
			},
			requireError: require.Error,
		},
		{
			name: "valid session",
			session: Session{
				UserID:     existingUserID,
				Device:     "device",
				RefreshJWT: newRefreshToken,
			},
			requireError: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := arrangeRepoWithTestDB(t)
			_, err := repo.pool.Exec(context.Background(), arrangeQuery)
			require.NoError(t, err, "error arranging db content")

			id, err := repo.add(context.Background(), tt.session)

			tt.requireError(t, err)
			if err != nil {
				return
			}

			tt.session.ID = id
			var (
				newSession Session
				isActive   bool
			)
			err = repo.pool.QueryRow(context.Background(), checkResultQuery, id).
				Scan(&newSession.ID, &newSession.UserID, &newSession.Device, &newSession.RefreshJWT, &isActive)
			require.NoError(t, err, "check result query error")
			assert.Equal(t, tt.session, newSession, "new session in db is incorrect")
			assert.True(t, isActive, "session is  not active")
		})
	}
}

func TestPostgresqlRepository_contains(t *testing.T) {
	const arrangeQuery = `
		WITH test_user AS (
			INSERT INTO users (id, username) 
			VALUES ('c171f212-6520-4a34-8761-c450b46cf853', 'user1')
			    
			RETURNING id
		)
		INSERT INTO sessions (id, user_id, refresh_jwt, device, is_active)
		VALUES 
		    ('b14f265a-7317-42fd-8d14-b7862bf3cb37', (SELECT id FROM test_user), 'active', 'device', true),
		    ('92627f5b-1ac8-4b70-9e1e-edd2ca14b471',  (SELECT id FROM test_user), 'not_active', 'device', false);
	`

	repo := arrangeRepoWithTestDB(t)
	_, err := repo.pool.Exec(context.Background(), arrangeQuery)
	require.NoError(t, err, "error arranging db content")

	var (
		userID = stringToUUID(t, "c171f212-6520-4a34-8761-c450b46cf853")

		notExistingSessionID = stringToUUID(t, "bda4f7ca-c10b-4a63-94b8-00c97dea08fd")
		activeSessionID      = stringToUUID(t, "b14f265a-7317-42fd-8d14-b7862bf3cb37")
		notActiveSessionID   = stringToUUID(t, "92627f5b-1ac8-4b70-9e1e-edd2ca14b471")

		activeSessionRefreshToken    = "active"
		notActiveSessionRefreshToken = "not_active"
	)

	tests := []struct {
		name          string
		session       Session
		requireError  require.ErrorAssertionFunc
		expectedError error
	}{
		{
			name: "session with invalid id",
			session: Session{
				ID:         pgtype.UUID{Valid: false},
				UserID:     userID,
				RefreshJWT: activeSessionRefreshToken,
			},
			requireError:  require.Error,
			expectedError: postgresql.ErrNoMatches,
		},
		{
			name: "session with not existing id",
			session: Session{
				ID:         notExistingSessionID,
				UserID:     userID,
				RefreshJWT: activeSessionRefreshToken,
			},
			requireError:  require.Error,
			expectedError: postgresql.ErrNoMatches,
		},
		{
			name: "not active session",
			session: Session{
				ID:         notActiveSessionID,
				UserID:     userID,
				RefreshJWT: notActiveSessionRefreshToken,
			},
			requireError:  require.Error,
			expectedError: postgresql.ErrNoMatches,
		},
		{
			name: "active session",
			session: Session{
				ID:         activeSessionID,
				UserID:     userID,
				RefreshJWT: activeSessionRefreshToken,
			},
			requireError: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.contains(context.Background(), tt.session)

			tt.requireError(t, err)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			}
		})
	}
}

func TestPostgresqlRepository_deactivate(t *testing.T) {
	const (
		arrangeQuery = `
			WITH test_user_with_active AS (
			    INSERT INTO users (id, username) 
			    VALUES ('c171f212-6520-4a34-8761-c450b46cf853', 'user1')
			    
				RETURNING id
			), test_without_active AS (
			    INSERT INTO users (id, username) 
			    VALUES ('31d6f3ce-36d0-4b78-b096-59e68cd0e6e3', 'user2')
			    
			    RETURNING id
			)
			INSERT INTO sessions (id, user_id, refresh_jwt, is_active)
			VALUES
			    ('b14f265a-7317-42fd-8d14-b7862bf3cb37', (SELECT id FROM test_user_with_active), '1', true),
			    ('7f5077df-c3f3-49d6-a943-d4124b18799c', (SELECT id FROM test_user_with_active), '2', true),
			    ('71862613-1d6c-4ba7-b90f-7493bf403e14', (SELECT id FROM test_user_with_active), '3', false),
			    ('c89e4974-26e3-4d3f-a309-97d87fadecee', (SELECT id FROM test_without_active), '4', false),
			    ('17662206-86dc-4f1f-b4bd-d1f305d1de03', (SELECT id FROM test_without_active), '5', false);
		`
		checkResultQueryForBySessionOpt = `
			SELECT NOT EXISTS(
			    SELECT 1
				FROM sessions
				WHERE id = $1
				AND is_active = true
			)  AS is_correct_result
		`
		checkResultQueryForByUserOpt = `
			SELECT NOT EXISTS (
			    SELECT 1 FROM sessions
			    WHERE user_id = $1
			    AND is_active = true
			) AS is_correct_result;
		`
	)

	var (
		notExistingUserID            = stringToUUID(t, "916cdb99-fad2-4e17-b067-0828b6bf9311")
		userIDWithSomeActiveSessions = stringToUUID(t, "c171f212-6520-4a34-8761-c450b46cf853")
		userIDWithNoActiveSessions   = stringToUUID(t, "31d6f3ce-36d0-4b78-b096-59e68cd0e6e3")

		notExistingSessionID = stringToUUID(t, "db66a39d-5c14-445f-b5ec-de6261582173")
		activeSessionID      = stringToUUID(t, "7f5077df-c3f3-49d6-a943-d4124b18799c")
		notActiveSessionID   = stringToUUID(t, "c89e4974-26e3-4d3f-a309-97d87fadecee")
	)

	tests := []struct {
		name          string
		opt           option
		id            pgtype.UUID
		checkingQuery string
	}{
		{
			name:          "by existing user with some active sessions",
			opt:           byUser(),
			id:            userIDWithSomeActiveSessions,
			checkingQuery: checkResultQueryForByUserOpt,
		},
		{
			name:          "by existing user without active sessions",
			opt:           byUser(),
			id:            userIDWithNoActiveSessions,
			checkingQuery: checkResultQueryForByUserOpt,
		},
		{
			name:          "by not existing user",
			opt:           byUser(),
			id:            notExistingUserID,
			checkingQuery: checkResultQueryForByUserOpt,
		},
		{
			name:          "by existing id of active session",
			opt:           bySession(activeSessionID),
			id:            activeSessionID,
			checkingQuery: checkResultQueryForBySessionOpt,
		},
		{
			name:          "by existing id of not active session",
			opt:           bySession(notActiveSessionID),
			id:            notActiveSessionID,
			checkingQuery: checkResultQueryForBySessionOpt,
		},
		{
			name:          "by not existing id",
			opt:           bySession(notExistingSessionID),
			id:            notExistingSessionID,
			checkingQuery: checkResultQueryForBySessionOpt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := arrangeRepoWithTestDB(t)
			_, err := repo.pool.Exec(context.Background(), arrangeQuery)
			require.NoError(t, err, "error arranging db content")
			ctx := usr.WithUserID(context.Background(), tt.id)

			err = repo.deactivate(ctx, tt.opt)

			require.NoError(t, err)

			var isDoneCorrectly bool

			err = repo.pool.QueryRow(context.Background(), tt.checkingQuery, tt.id).
				Scan(&isDoneCorrectly)
			require.NoError(t, err, "check result query error")
			assert.True(t, isDoneCorrectly, "required session is active")
		})
	}
}

func TestPostgresqlRepository_isDeviceNewWhenSessionIsNotFirst(t *testing.T) {
	const arrangeQuery = `
		WITH test_user_with_sessions AS (
			INSERT INTO users (id, username) 
			VALUES ('c171f212-6520-4a34-8761-c450b46cf853', 'user1')
			    
			RETURNING id
		), test_user_without_sessions AS(
			INSERT INTO users (id, username) 
			VALUES ('ff93c7d9-0b1b-4b74-b102-ef1893cf4b89', 'user2')
		), previous_sessions AS (
			INSERT INTO sessions (id, user_id, refresh_jwt, device, is_active)
			VALUES 
		    	('b14f265a-7317-42fd-8d14-b7862bf3cb37', (SELECT id FROM test_user_with_sessions), '1', 'device1', true),
		    	('92627f5b-1ac8-4b70-9e1e-edd2ca14b471',  (SELECT id FROM test_user_with_sessions), '2', 'device2', false)
		), new_session_with_old_device AS (
			INSERT INTO sessions (id, user_id, refresh_jwt, device, is_active)
			VALUES ('823fd46e-fa36-4157-b147-a121006fc21e', (SELECT id FROM test_user_with_sessions), '3', 'device1', true)
		), new_session_with_new_device AS (
		    INSERT INTO sessions (id, user_id, refresh_jwt, device, is_active)
			VALUES ('3ae909e7-7e8c-4724-b96a-472a515ba755', (SELECT id FROM test_user_with_sessions), '4', 'device3', true)
		)
		SELECT 1;
	`

	repo := arrangeRepoWithTestDB(t)
	_, err := repo.pool.Exec(context.Background(), arrangeQuery)
	require.NoError(t, err, "error arranging db content")

	var (
		notExistingUserID     = stringToUUID(t, "b068d861-13ce-4c1c-a853-6a3cc0d93421")
		userIDWithSessions    = stringToUUID(t, "c171f212-6520-4a34-8761-c450b46cf853")
		userIDWithoutSessions = stringToUUID(t, "ff93c7d9-0b1b-4b74-b102-ef1893cf4b89")

		notExistingSessionID      = stringToUUID(t, "bda4f7ca-c10b-4a63-94b8-00c97dea08fd")
		newSessionWithOldDeviceID = stringToUUID(t, "823fd46e-fa36-4157-b147-a121006fc21e")
		newSessionWithNewDeviceID = stringToUUID(t, "3ae909e7-7e8c-4724-b96a-472a515ba755")
	)

	tests := []struct {
		name           string
		session        Session
		expectedResult bool
	}{
		{
			name: "user not exist",
			session: Session{
				ID:     newSessionWithNewDeviceID,
				UserID: notExistingUserID,
				Device: "device3",
			},
			expectedResult: false,
		},
		{
			name: "session is first for user",
			session: Session{
				ID:     newSessionWithNewDeviceID,
				UserID: userIDWithoutSessions,
				Device: "device3",
			},
			expectedResult: false,
		},
		{
			name: "device existed",
			session: Session{
				ID:     newSessionWithOldDeviceID,
				UserID: userIDWithSessions,
				Device: "device1",
			},
			expectedResult: false,
		},
		{
			name: "device is new and session is not first for user",
			session: Session{
				ID:     newSessionWithNewDeviceID,
				UserID: userIDWithSessions,
				Device: "device3",
			},
			expectedResult: true,
		},
		{
			name: "session is not exist before but user has sessions and device is new",
			session: Session{
				ID:     notExistingSessionID,
				UserID: userIDWithSessions,
				Device: "device5",
			},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.isDeviceNewWhenUserHadSessionsBefore(context.Background(), tt.session)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result, "wrong result")
		})
	}
}

func arrangeRepoWithTestDB(t *testing.T) *PostgresqlRepository {
	config := test.PostgresConfig{
		Username: "postgres",
		Password: "12345",
		Host:     "localhost",
		Port:     5432,
		DBName:   "test" + strconv.Itoa(rand.Int()),
	}

	path := os.Getenv("SQL_AUTHENTICATION_INIT_FILE_PATH")
	require.NotEmpty(t, path, "path to .sql is empty")

	pool := test.SetupDatabase(t, config, path)
	t.Cleanup(func() {
		test.DropDatabase(t, pool, config)
	})

	repo, err := NewPostgresqlRepository(pool)
	require.NoError(t, err, "error creating repo")
	return repo
}

func stringToUUID(t *testing.T, uuidRaw string) pgtype.UUID {
	t.Helper()
	idBytes, err := uuid.Parse(uuidRaw)
	require.NoError(t, err, "arranging uuid error")
	return pgtype.UUID{
		Bytes: idBytes,
		Valid: true,
	}
}
