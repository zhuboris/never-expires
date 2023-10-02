package usr

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhuboris/never-expires/internal/id/usr/oauth"
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

func TestPostgresqlRepository_addByPassword(t *testing.T) {
	tests := []struct {
		name         string
		arrangeQuery string
		userToAdd    User
		expectedUser User
		requireError require.ErrorAssertionFunc
	}{
		{
			name: "email should be not confirmed",
			userToAdd: User{
				Username: "user1",
				Email:    "a@a.com",
				Password: "1234",
			},
			expectedUser: User{
				Username:         "user1",
				Email:            "a@a.com",
				IsEmailConfirmed: false,
			},
			requireError: require.NoError,
		},
		{
			name: "email is lowered",
			userToAdd: User{
				Username: "user1",
				Email:    "A@A.com",
				Password: "1234",
			},
			expectedUser: User{
				Username:         "user1",
				Email:            "a@a.com",
				IsEmailConfirmed: false,
			},
			requireError: require.NoError,
		},
		{
			name: "email already in db",
			arrangeQuery: `
				WITH new_user AS (
					INSERT INTO users (id, username)
					VALUES ('f01a2932-181a-468a-9226-95ae131ca1cc', 'user2')
				)
				INSERT INTO emails (email, owner_id) 
				VALUES ('email1@e.com', 'f01a2932-181a-468a-9226-95ae131ca1cc');
			`,
			userToAdd: User{
				Username: "user1",
				Email:    "email1@e.com",
				Password: "1234",
			},
			requireError: require.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := arrangeRepoWithTestDB(t)
			if tt.arrangeQuery != "" {
				_, err := repo.pool.Exec(context.Background(), tt.arrangeQuery)
				require.NoError(t, err, "error arranging db content")
			}

			result, err := repo.addByPassword(context.Background(), tt.userToAdd)

			tt.requireError(t, err)
			if err != nil {
				return
			}

			assertUsersAreEqual(t, tt.expectedUser, *result)
			assert.NotEqual(t, result.ID, tt.expectedUser.ID, "ID must be set in db")
			assert.True(t, result.ID.Valid, "UUID is invalid or not set")
		})
	}
}

func TestPostgresqlRepository_byID(t *testing.T) {
	const arrangingSQL = `
		WITH existing_user AS (
			INSERT INTO users (id, username) 
			VALUES ($1, $2) 
		)
		INSERT INTO emails (email, owner_id, is_confirmed) 
		VALUES ($3, $1, $4);
	`

	var (
		existingUserID = stringToUUID(t, "5b9049ae-3565-414f-acab-b7cdd26bf0bb")
		existingUser   = User{
			ID:               existingUserID,
			Username:         "user1",
			Email:            "qwe@rty.com",
			IsEmailConfirmed: true,
		}
		userToAdd = User{
			Username: "user2",
			Email:    "asd@rty.com",
			Password: "123",
		}
	)

	repo := arrangeRepoWithTestDB(t)
	_, err := repo.pool.Exec(context.Background(), arrangingSQL, existingUser.ID, existingUser.Username, existingUser.Email, existingUser.IsEmailConfirmed)
	require.NoError(t, err, "arranging repo error")

	newUser, err := repo.addByPassword(context.Background(), userToAdd)
	require.NoError(t, err, "arranging repo error")

	tests := []struct {
		name          string
		id            pgtype.UUID
		expectedUser  *User
		requireError  require.ErrorAssertionFunc
		expectedError error
	}{
		{
			name: "invalid uuid",
			id: pgtype.UUID{
				Bytes: [16]byte{123, 45, 67, 89, 10, 32, 52, 16, 0, 3, 65, 23, 45, 67, 89, 10},
			},
			requireError: require.Error,
		},
		{
			name: "valid uuid but not presenting in db",
			id: pgtype.UUID{
				Bytes: [16]byte{155, 140, 161, 56, 216, 239, 79, 36, 160, 76, 101, 45, 117, 207, 7, 35},
				Valid: true,
			},
			requireError:  require.Error,
			expectedError: postgresql.ErrNoMatches,
		},
		{
			name:         "uuid of existing user added with query manually",
			id:           existingUserID,
			expectedUser: &existingUser,
			requireError: require.NoError,
		},
		{
			name:         `uuid of existing user added with repo method "addByPassword"`,
			id:           newUser.ID,
			expectedUser: newUser,
			requireError: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.byID(context.Background(), tt.id)
			tt.requireError(t, err)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			}

			if err == nil { // if NO error
				assert.Equal(t, tt.expectedUser, result, "unexpected result")
			}
		})
	}
}

func TestPostgresqlRepository_byEmail(t *testing.T) {
	const arrangingSQL = `
		WITH existing_user AS (
			INSERT INTO users (id, username) 
			VALUES ($1, $2) 
		), pw AS (
		    INSERT INTO passwords (user_id, encrypted_password)
		    VALUES ($1, $5)
		)
		INSERT INTO emails (email, owner_id, is_confirmed) 
		VALUES ($3, $1, $4);
	`

	var (
		existingUserID = stringToUUID(t, "5b9049ae-3565-414f-acab-b7cdd26bf0bb")
		existingUser   = User{
			ID:               existingUserID,
			Username:         "user1",
			Email:            "1@test.com",
			IsEmailConfirmed: true,
			Password:         "777",
		}
		userToAdd = User{
			Username: "user2",
			Email:    "2@test.com",
			Password: "123",
		}
	)

	repo := arrangeRepoWithTestDB(t)
	_, err := repo.pool.Exec(context.Background(), arrangingSQL, existingUser.ID, existingUser.Username, existingUser.Email, existingUser.IsEmailConfirmed, existingUser.Password)
	require.NoError(t, err, "arranging repo error")

	newUser, err := repo.addByPassword(context.Background(), userToAdd)
	require.NoError(t, err, "arranging repo error")
	newUser.Password = userToAdd.Password

	tests := []struct {
		name          string
		email         string
		expectedUser  *User
		requireError  require.ErrorAssertionFunc
		expectedError error
	}{
		{
			name:          "not existing in db email",
			email:         "not_exist@test.com",
			requireError:  require.Error,
			expectedError: postgresql.ErrNoMatches,
		},
		{
			name:         "email of existing user added with query manually",
			email:        existingUser.Email,
			expectedUser: &existingUser,
			requireError: require.NoError,
		},
		{
			name:         `email of existing user added with repo method "addByPassword"`,
			email:        newUser.Email,
			expectedUser: newUser,
			requireError: require.NoError,
		},
		{
			name:         "email of existing user with changed case",
			email:        strings.ToUpper(existingUser.Email),
			expectedUser: &existingUser,
			requireError: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.byEmail(context.Background(), tt.email)
			tt.requireError(t, err)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			}

			if err == nil { // if NO error
				assert.Equal(t, tt.expectedUser, result, "unexpected result")
			}
		})
	}
}

func TestPostgresqlRepository_encryptedPassword(t *testing.T) {
	const arrangingSQL = `
		WITH existing_user AS (
			INSERT INTO users (id, username) 
			VALUES ($1, $2) 
		)
		INSERT INTO passwords (user_id, encrypted_password)
		VALUES ($1, $3);
	`

	var (
		existingUserID   = stringToUUID(t, "5b9049ae-3565-414f-acab-b7cdd26bf0bb")
		existingTestUser = User{
			ID:       existingUserID,
			Username: "user1",
			Password: "777",
		}
		userToAdd = User{
			Username: "user2",
			Email:    "2@test.com",
			Password: "123",
		}
	)

	repo := arrangeRepoWithTestDB(t)
	_, err := repo.pool.Exec(context.Background(), arrangingSQL, existingTestUser.ID, existingTestUser.Username, existingTestUser.Password)
	require.NoError(t, err, "arranging repo error")

	newUser, err := repo.addByPassword(context.Background(), userToAdd)
	require.NoError(t, err, "arranging repo error")
	newUser.Password = userToAdd.Password

	tests := []struct {
		name             string
		id               pgtype.UUID
		expectedPassword string
		requireError     require.ErrorAssertionFunc
	}{
		{
			name: "invalid user id",
			id: pgtype.UUID{
				Bytes: [16]byte{123, 45, 67, 89, 10, 32, 52, 16, 0, 3, 65, 23, 45, 67, 89, 10},
			},
			requireError: require.Error,
		},
		{
			name: "not existing in db user id",
			id: pgtype.UUID{
				Bytes: [16]byte{155, 140, 161, 56, 216, 239, 79, 36, 160, 76, 101, 45, 117, 207, 7, 35},
				Valid: true,
			},
			requireError: require.Error,
		},
		{
			name:             "id of existing user added with query manually",
			id:               existingTestUser.ID,
			expectedPassword: existingTestUser.Password,
			requireError:     require.NoError,
		},
		{
			name:             `id of existing user added with repo method "addByPassword"`,
			id:               newUser.ID,
			expectedPassword: newUser.Password,
			requireError:     require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.encryptedPassword(context.Background(), tt.id)
			tt.requireError(t, err)

			if err == nil { // if NO error
				assert.Equal(t, tt.expectedPassword, result, "unexpected result")
			}
		})
	}
}

func TestPostgresqlRepository_updateColumn(t *testing.T) {
	const arrangingSQL = `
		WITH existing_user AS (
			INSERT INTO users (id, username) 
			VALUES ($1, $2), ($3, $4)
		), pw AS (
		    INSERT INTO passwords (user_id, encrypted_password)
		    VALUES ($1, $5), ($3, $6)
		)
		INSERT INTO emails (email, owner_id) 
		VALUES ($7, $1), ($8, $3);
	`

	var (
		user1ID = stringToUUID(t, "5b9049ae-3565-414f-acab-b7cdd26bf0bb")
		user2ID = stringToUUID(t, "f39ff96c-c39c-4dee-bcc8-0ae35d7186dd")
		user1   = User{
			ID:       user1ID,
			Username: "user1",
			Email:    "1@test.com",
			Password: "777",
		}
		user2 = User{
			ID:       user2ID,
			Username: "user2",
			Email:    "2@test.com",
			Password: "123",
		}
		newFreeEmail = "3@test.com"
		newUsername  = "user3"
		newPassword  = "12345"
	)

	type args struct {
		userID   pgtype.UUID
		toUpdate column
		new      string
	}

	tests := []struct {
		name          string
		input         args
		requireError  require.ErrorAssertionFunc
		expectedError error
	}{
		{
			name: "invalid user id",
			input: args{
				userID: pgtype.UUID{
					Bytes: [16]byte{123, 45, 67, 89, 10, 32, 52, 16, 0, 3, 65, 23, 45, 67, 89, 10},
				},
			},
			requireError: require.Error,
		},
		{
			name: "not existing in db user id",
			input: args{
				userID: pgtype.UUID{
					Bytes: [16]byte{155, 140, 161, 56, 216, 239, 79, 36, 160, 76, 101, 45, 117, 207, 7, 35},
					Valid: true,
				},
			},
			requireError:  require.Error,
			expectedError: postgresql.ErrNoMatches,
		},
		{
			name: "change email to already existing",
			input: args{
				userID:   user1ID,
				toUpdate: email,
				new:      user2.Email,
			},
			requireError:  require.Error,
			expectedError: postgresql.ErrAddedDuplicateOfUnique,
		},
		{
			name: "change email to free",
			input: args{
				userID:   user1ID,
				toUpdate: email,
				new:      newFreeEmail,
			},
			requireError: require.NoError,
		},
		{
			name: "change username to already existing",
			input: args{
				userID:   user1ID,
				toUpdate: username,
				new:      user2.Username,
			},
			requireError: require.NoError,
		},
		{
			name: "change username to free",
			input: args{
				userID:   user1ID,
				toUpdate: username,
				new:      newUsername,
			},
			requireError: require.NoError,
		},
		{
			name: "change password to already existing",
			input: args{
				userID:   user1ID,
				toUpdate: password,
				new:      user2.Password,
			},
			requireError: require.NoError,
		},
		{
			name: "change password to free",
			input: args{
				userID:   user1ID,
				toUpdate: password,
				new:      newPassword,
			},
			requireError: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := arrangeRepoWithTestDB(t)
			_, err := repo.pool.Exec(context.Background(), arrangingSQL,
				user1.ID,
				user1.Username,
				user2.ID,
				user2.Username,
				user1.Password,
				user2.Password,
				user1.Email,
				user2.Email,
			)
			require.NoError(t, err, "arranging repo error")

			err = repo.updateColumn(context.Background(), tt.input.userID, tt.input.toUpdate, tt.input.new)
			tt.requireError(t, err)
		})
	}
}

func TestPostgresqlRepository_validateEmail(t *testing.T) {
	const (
		arrangeQuery = `
			WITH test_user AS (
			    INSERT INTO users (username) 
				VALUES ('user1')
			    
				RETURNING id
			), users_emails AS (
			    INSERT INTO emails (email, owner_id, is_confirmed)
			    VALUES 
			        ('not_confirmed@test.com', (SELECT id FROM test_user), false),
			        ('confirmed@test.com', (SELECT id FROM test_user), true)
			)
			INSERT INTO mail_confirmation_tokens (token, email, is_used, expiration)
			VALUES 
				('valid1', 'not_confirmed@test.com', false, NOW() + INTERVAL '1 day'),
				('valid2', 'confirmed@test.com', false, NOW() + INTERVAL '1 day'),
				('expired', 'not_confirmed@test.com', false, NOW() - INTERVAL '1 day'),
				('used', 'not_confirmed@test.com', true, NOW() + INTERVAL '1 day');
		`
		checkResultQuery = `
			WITH is_token_used AS (
			    SELECT is_used FROM mail_confirmation_tokens
			    WHERE token = $1
			), is_email_confirmed AS (
				SELECT e.is_confirmed
				FROM mail_confirmation_tokens mct
				LEFT JOIN emails e ON e.email = mct.email
				WHERE mct.token = $1
			)
			SELECT t.is_used AND e.is_confirmed AS is_correct
			FROM is_token_used t, is_email_confirmed e;
		`
	)

	tests := []struct {
		name            string
		validationToken string
		requireError    require.ErrorAssertionFunc
		expectedError   error
	}{
		{
			name:            "empty token",
			validationToken: "",
			requireError:    require.Error,
		},
		{
			name:            "token not exist",
			validationToken: "notExist",
			requireError:    require.Error,
			expectedError:   errTokenNotExists,
		},
		{
			name:            "expired token",
			validationToken: "expired",
			requireError:    require.Error,
			expectedError:   errTokenExpired,
		},
		{
			name:            "token already used",
			validationToken: "used",
			requireError:    require.Error,
			expectedError:   errTokenAlreadyUsed,
		},
		{
			name:            "email already confirmed",
			validationToken: "valid2",
			requireError:    require.Error,
			expectedError:   ErrEmailAlreadyConfirmed,
		},
		{
			name:            "email confirmed correctly",
			validationToken: "valid1",
			requireError:    require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := arrangeRepoWithTestDB(t)
			_, err := repo.pool.Exec(context.Background(), arrangeQuery)
			require.NoError(t, err, "error arranging db content")

			err = repo.validateEmail(context.Background(), tt.validationToken)

			tt.requireError(t, err)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			}

			if err != nil {
				return
			}

			var isConfirmed bool
			err = repo.pool.QueryRow(context.Background(), checkResultQuery, tt.validationToken).
				Scan(&isConfirmed)
			require.NoError(t, err, "check result query error")
			assert.True(t, isConfirmed, "email was not confirmed")
		})
	}
}

func TestPostgresqlRepository_restorePassword(t *testing.T) {
	const (
		arrangeQuery = `
			WITH test_user_without_pw AS (
			    INSERT INTO users (username) 
				VALUES ('user1')
			    
				RETURNING id
			), test_user_with_pw AS (
			    INSERT INTO users (username) 
				VALUES ('user2')
			    
				RETURNING id
			), old_password AS (
			    INSERT INTO passwords (user_id, encrypted_password) 
			    VALUES ((SELECT id FROM test_user_with_pw), 'old')
			), users_emails AS (
			    INSERT INTO emails (email, owner_id, is_confirmed)
			    VALUES 
			        ('not_confirmed@test.com', (SELECT id FROM test_user_without_pw), false),
			        ('not_active@test.com', (SELECT id FROM test_user_without_pw), true),
			        ('confirmed_user_without_pw@test.com', (SELECT id FROM test_user_without_pw), true),
			        ('confirmed_user_with_pw@test.com', (SELECT id FROM test_user_with_pw), true)
			)
			INSERT INTO password_restoration_tokens (token, user_email, is_used, expiration)
			VALUES 
				('valid1', 'not_confirmed@test.com', false, NOW() + INTERVAL '1 day'),
				('valid2', 'not_active@test.com', false, NOW() + INTERVAL '1 day'),
				('valid3', 'confirmed_user_without_pw@test.com', false, NOW() + INTERVAL '1 day'),
				('valid4', 'confirmed_user_with_pw@test.com', false, NOW() + INTERVAL '1 day'),
				('expired', 'confirmed_user_without_pw@test.com', false, NOW() - INTERVAL '1 day'),
				('used', 'confirmed_user_without_pw@test.com', true, NOW() + INTERVAL '1 day');
		`
		checkResultQuery = `
			WITH is_token_used AS (
			    SELECT is_used FROM password_restoration_tokens
			    WHERE token = $1
			), is_email_correct AS (
			   SELECT e.email = $2 AS is_correct_email
				FROM password_restoration_tokens mct
				LEFT JOIN emails e ON e.email = mct.user_email
				WHERE mct.token = $1 
			), is_password_new AS (
				SELECT p.encrypted_password = $3 AS is_correct_new_password
				FROM password_restoration_tokens mct
				LEFT JOIN emails e ON e.email = mct.user_email
				LEFT JOIN passwords p ON e.owner_id = p.user_id
				WHERE mct.token = $1 
			)
			SELECT t.is_used AND e.is_correct_email AND p.is_correct_new_password AS is_done
			FROM is_token_used t, is_email_correct e, is_password_new p;
		`
	)

	type args struct {
		validationToken string
		newPassword     string
	}

	tests := []struct {
		name          string
		input         args
		requireError  require.ErrorAssertionFunc
		expectedError error
	}{
		{
			name: "empty token",
			input: args{
				validationToken: "",
				newPassword:     "123",
			},
			requireError: require.Error,
		},
		{
			name: "token not exist",
			input: args{
				validationToken: "notExist",
				newPassword:     "123",
			},
			requireError:  require.Error,
			expectedError: errTokenNotExists,
		},
		{
			name: "expired token",
			input: args{
				validationToken: "expired",
				newPassword:     "123",
			},
			requireError:  require.Error,
			expectedError: errTokenExpired,
		},
		{
			name: "token already used",
			input: args{
				validationToken: "used",
				newPassword:     "123",
			},
			requireError:  require.Error,
			expectedError: errTokenAlreadyUsed,
		},
		{
			name: "email not confirmed",
			input: args{
				validationToken: "valid1",
				newPassword:     "123",
			},
			requireError:  require.Error,
			expectedError: ErrNotConfirmedOrChangedEmail,
		},
		{
			name: "email was already changed",
			input: args{
				validationToken: "valid2",
				newPassword:     "123",
			},
			requireError:  require.Error,
			expectedError: ErrNotConfirmedOrChangedEmail,
		},
		{
			name: "email confirmed, user have no password set before",
			input: args{
				validationToken: "valid3",
				newPassword:     "123",
			},
			requireError: require.NoError,
		},
		{
			name: "email confirmed, user have a password set before",
			input: args{
				validationToken: "valid4",
				newPassword:     "123",
			},
			requireError: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := arrangeRepoWithTestDB(t)
			_, err := repo.pool.Exec(context.Background(), arrangeQuery)
			require.NoError(t, err, "error arranging db content")

			userEmail, err := repo.restorePassword(context.Background(), tt.input.validationToken, tt.input.newPassword)

			tt.requireError(t, err)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			}

			if err != nil {
				return
			}

			var isChanged bool
			err = repo.pool.QueryRow(context.Background(), checkResultQuery, tt.input.validationToken, userEmail, tt.input.newPassword).
				Scan(&isChanged)
			require.NoError(t, err, "check result query error")
			assert.True(t, isChanged, "password was not changed")
		})
	}
}

func TestPostgresqlRepository_isConfirmed(t *testing.T) {
	const (
		arrangeQuery = `
			WITH test_user_with_active_confirmed AS (
			    INSERT INTO users (username) 
				VALUES ('user1')
			    
				RETURNING id
			), test_user_with_active_not_confirmed AS (
			    INSERT INTO users (username) 
				VALUES ('user2')
			    
				RETURNING id
			)
			INSERT INTO emails (email, owner_id, is_confirmed)
			VALUES 
				('not_active_confirmed@test.com', (SELECT id FROM test_user_with_active_confirmed), true),
			    ('not_active_not_confirmed@test.com', (SELECT id FROM test_user_with_active_confirmed), false),
			   	('active_confirmed@test.com', (SELECT id FROM test_user_with_active_confirmed), true),
			    ('active_not_confirmed@test.com', (SELECT id FROM test_user_with_active_not_confirmed), false);
		`
		checkResultQuery = `
			SELECT is_confirmed AND is_active AS correct_result
			FROM emails
			WHERE email = $1;
		`
	)

	tests := []struct {
		name           string
		arrangeQuery   string
		email          string
		expectedResult bool
	}{
		{
			name:           "email not exist in db",
			email:          "not_exist@test.com",
			expectedResult: false,
		},
		{
			name:           "email not confirmed and not active",
			email:          "not_active_not_confirmed@test.com",
			expectedResult: false,
		},
		{
			name:           "email active but not confirmed",
			email:          "not_active_confirmed@test.com",
			expectedResult: false,
		},
		{
			name:           "email confirmed but not active",
			email:          "active_not_confirmed@test.com",
			expectedResult: false,
		},
		{
			name:           "active email confirmed",
			email:          "active_confirmed@test.com",
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := arrangeRepoWithTestDB(t)
			_, err := repo.pool.Exec(context.Background(), arrangeQuery)
			require.NoError(t, err, "error arranging db content")

			isConfirmed, err := repo.isConfirmed(context.Background(), tt.email)
			require.NoError(t, err, "error in query")
			assert.Equal(t, tt.expectedResult, isConfirmed)
			if !isConfirmed {
				return
			}

			err = repo.pool.QueryRow(context.Background(), checkResultQuery, tt.email).
				Scan(&isConfirmed)
			require.NoError(t, err, "check result query error")
			assert.True(t, isConfirmed, "email was not confirmed")
		})
	}
}

func TestPostgresqlRepository_byOAuth(t *testing.T) {
	const arrangeQuery = `
		WITH test_google_id_user AS (
			INSERT INTO users (username) 
				VALUES ('user1')
			    
			RETURNING id
		), test_apple_id_user AS (
			INSERT INTO users (username) 
			VALUES ('user2')
			    
			RETURNING id
		), test_no_oauth_user AS (
			INSERT INTO users (username) 
			VALUES ('user2')
			    
			RETURNING id
		), google_id AS (
			INSERT INTO google_ids (user_id, id)
			VALUES ((SELECT id FROM test_google_id_user), 'existing_google_id')
		), apple_id AS (
		    INSERT INTO apple_ids (user_id, id)
			VALUES ((SELECT id FROM test_google_id_user), 'existing_apple_id')
		)
		INSERT INTO emails (email, owner_id)
		VALUES 
			('1@test.com', (SELECT id FROM test_google_id_user)),
			('2@test.com', (SELECT id FROM test_apple_id_user)),
			('3@test.com', (SELECT id FROM test_no_oauth_user));
	`

	checkResultQuery := func(method oAuthMethod) string {
		const sqlFormat = `
			SELECT EXISTS (
			    SELECT 1 FROM %s
			    WHERE id = $1
			    AND user_id = $2
			)
		`

		t.Helper()
		tableName := method()
		return fmt.Sprintf(sqlFormat, tableName)
	}

	tests := []struct {
		name           string
		oAuthMethod    oAuthMethod
		user           oauth.User
		expectedResult oauth.LoginResultType
	}{
		{
			name:           "oAuth id is already registered google",
			oAuthMethod:    withGoogle(),
			user:           oauth.NewUser("existing_google_id", "1@test.com", "user", false /*isEmailVerified*/),
			expectedResult: oauth.Login,
		},
		{
			name:           "oAuth id is already registered apple",
			oAuthMethod:    withApple(),
			user:           oauth.NewUser("existing_apple_id", "2@test.com", "user", false /*isEmailVerified*/),
			expectedResult: oauth.Login,
		},
		{
			name:           "oAuth id connecting to existing account by matching email google",
			oAuthMethod:    withGoogle(),
			user:           oauth.NewUser("new_id1", "3@test.com", "user", false /*isEmailVerified*/),
			expectedResult: oauth.Connect,
		},
		{
			name:           "oAuth id connecting to existing account by matching email apple",
			oAuthMethod:    withApple(),
			user:           oauth.NewUser("new_id2", "3@test.com", "user", false /*isEmailVerified*/),
			expectedResult: oauth.Connect,
		},
		{
			name:           "register new account with oAuth google",
			oAuthMethod:    withGoogle(),
			user:           oauth.NewUser("new_id3", "new1@test.com", "user", false /*isEmailVerified*/),
			expectedResult: oauth.Register,
		},
		{
			name:           "register new account with oAuth apple",
			oAuthMethod:    withApple(),
			user:           oauth.NewUser("new_id4", "new2@test.com", "user", false /*isEmailVerified*/),
			expectedResult: oauth.Register,
		},
		{
			name:           "connect different types of oAuth by matching email",
			oAuthMethod:    withGoogle(),
			user:           oauth.NewUser("existing_apple_id", "2@test.com", "user", false /*isEmailVerified*/),
			expectedResult: oauth.Connect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := arrangeRepoWithTestDB(t)
			_, err := repo.pool.Exec(context.Background(), arrangeQuery)
			require.NoError(t, err, "error arranging db content")

			user, queryResult, err := repo.byOAuth(context.Background(), tt.user, tt.oAuthMethod)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, queryResult)

			var isRegistered bool
			err = repo.pool.QueryRow(context.Background(), checkResultQuery(tt.oAuthMethod), tt.user.ID(), user.ID).
				Scan(&isRegistered)
			require.NoError(t, err, "check result query error")
			assert.True(t, isRegistered, "password was not changed")
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

func assertUsersAreEqual(t *testing.T, first, second User) {
	t.Helper()
	assert.Equal(t, first.Email, second.Email)
	assert.Equal(t, first.Username, second.Username)
	assert.Equal(t, first.IsEmailConfirmed, second.IsEmailConfirmed)
}
