package usr

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zhuboris/never-expires/internal/id/usr/oauth"
	"github.com/zhuboris/never-expires/internal/shared/postgresql"
)

type PostgresqlRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresqlRepository(pool *pgxpool.Pool) (*PostgresqlRepository, error) {
	if pool == nil {
		return nil, postgresql.ErrPoolInitRequired
	}

	return &PostgresqlRepository{
		pool: pool,
	}, nil
}

func (r PostgresqlRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r PostgresqlRepository) addByPassword(ctx context.Context, user User) (*User, error) {
	const sql = `
		WITH new_user AS (
			INSERT INTO users (username)
			VALUES ($1)
			
			RETURNING id, username
		), new_email AS (
		    INSERT INTO emails (email, owner_id) 
		    VALUES (LOWER($2), (SELECT id FROM new_user))
		           
		    RETURNING email, owner_id, is_confirmed 
		), new_password AS (
			INSERT INTO passwords (user_id, encrypted_password)
			VALUES ((SELECT id FROM new_user), $3)
		)
		SELECT u.id, u.username, e.email, e.is_confirmed
		FROM new_user u
		LEFT JOIN new_email e ON u.id = e.owner_id;
	`

	addedUser := new(User)
	err := r.pool.QueryRow(ctx, sql, user.Username, user.Email, user.Password).
		Scan(&addedUser.ID,
			&addedUser.Username,
			&addedUser.Email,
			&addedUser.IsEmailConfirmed,
		)

	return addedUser, postgresql.HandleQueryErr(err)
}

func (r PostgresqlRepository) byID(ctx context.Context, id pgtype.UUID) (*User, error) {
	const sql = `
		SELECT 
		    u.id, 
		    u.username, 
		    e.email,
			e.is_confirmed
		FROM emails e
		LEFT JOIN users u 
		ON e.owner_id = u.id
		WHERE e.is_active = true
		AND e.owner_id = $1;
	`

	user := new(User)
	err := r.pool.
		QueryRow(ctx, sql, id).
		Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.IsEmailConfirmed,
		)

	if errors.Is(err, pgx.ErrNoRows) {
		return user, errors.Join(postgresql.ErrNoMatches, err)
	}

	return user, postgresql.HandleQueryErr(err)
}

func (r PostgresqlRepository) delete(ctx context.Context, id pgtype.UUID) error {
	const sql = `
		WITH saved_deleted_id AS (
		    INSERT INTO users_to_delete (id) VALUES ($1)
			ON CONFLICT (id) DO NOTHING			
		)
		DELETE FROM users
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, sql, id)
	return postgresql.HandleQueryErr(err)
}

func (r PostgresqlRepository) byEmail(ctx context.Context, email string) (*User, error) {
	const sql = `
		SELECT 
		    u.id, 
		    u.username, 
		    e.email, 
		    p.encrypted_password,
			e.is_confirmed
		FROM emails e
		LEFT JOIN users u ON e.owner_id = u.id
		LEFT JOIN passwords p ON u.id = p.user_id
		WHERE e.email ILIKE $1;
	`

	userEntity := new(Entity)
	err := r.pool.
		QueryRow(ctx, sql, email).
		Scan(
			&userEntity.ID,
			&userEntity.Username,
			&userEntity.Email,
			&userEntity.Password,
			&userEntity.IsEmailConfirmed,
		)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.Join(postgresql.ErrNoMatches, err)
	}

	return userEntity.ToUser(), postgresql.HandleQueryErr(err)
}

func (r PostgresqlRepository) encryptedPassword(ctx context.Context, userID pgtype.UUID) (string, error) {
	const sql = `
		SELECT encrypted_password
		FROM passwords
		WHERE user_id = $1;
	`

	var result string
	err := r.pool.
		QueryRow(ctx, sql, userID).
		Scan(&result)

	return result, postgresql.HandleQueryErr(err)
}

func (r PostgresqlRepository) updateColumn(ctx context.Context, userID pgtype.UUID, toUpdate column, new string) error {
	const (
		updatePasswordSql = `
			UPDATE passwords
			SET encrypted_password = $1
			WHERE user_id = $2;
		`
		updateUsernameSql = `
			UPDATE users
			SET username = $1
			WHERE id = $2;
		`
		updateEmailSql = `
			INSERT INTO emails (email, owner_id)
			VALUES ($1, $2);
		`
	)

	var sql string
	switch toUpdate {
	case password:
		sql = updatePasswordSql
	case username:
		sql = updateUsernameSql
	case email:
		sql = updateEmailSql
	default:
		return errors.New("asked to update not existing column")
	}

	result, err := r.pool.Exec(ctx, sql, new, userID)
	if err != nil {
		return postgresql.CheckErrorForUniqueViolation(err)
	}

	if result.RowsAffected() == 0 {
		return postgresql.ErrNoMatches
	}

	return nil
}

func (r PostgresqlRepository) validateEmail(ctx context.Context, validationToken string) error {
	const sql = `
		WITH existing_token AS (
        	SELECT  
            	mct.token,
            	mct.is_used,
            	mct.email,
				e.is_confirmed AS is_email_confirmed,
				mct.expiration 
            FROM mail_confirmation_tokens mct 
            RIGHT JOIN emails e 
            ON e.email = mct.email 
            WHERE mct.token = $1
        ), used_valid_token AS (
        	UPDATE mail_confirmation_tokens 
        	SET is_used = TRUE
			WHERE token IN 
			      		(SELECT token FROM existing_token
			        	WHERE is_used = FALSE 
						AND is_email_confirmed = FALSE 
						AND expiration > CURRENT_TIMESTAMP) 
        	
			RETURNING email 
        ), confirmed_email AS (
        	UPDATE emails 
        	SET is_confirmed = TRUE 
        	WHERE email IN (SELECT email FROM used_valid_token)
        	       
        	RETURNING email
        )
		
        SELECT
            EXISTS(SELECT 1 FROM confirmed_email) AS is_done_successfully,
            t.is_email_confirmed,
            t.token, 
            t.is_used, 
            t.expiration
        FROM (SELECT 1) s
		INNER JOIN existing_token t ON TRUE;
    `

	var (
		isValidated        bool
		isAlreadyConfirmed bool
		existingToken      ConfirmationToken
	)

	err := r.pool.QueryRow(ctx, sql, validationToken).
		Scan(
			&isValidated,
			&isAlreadyConfirmed,
			&existingToken.Value,
			&existingToken.IsUsed,
			&existingToken.ExpirationTime,
		)

	if err != nil {
		return handleSearchingTokenError(err)
	}

	if isAlreadyConfirmed {
		return ErrEmailAlreadyConfirmed
	}

	if !isValidated {
		return existingToken.InvalidityReason()
	}

	return nil
}

func (r PostgresqlRepository) restorePassword(ctx context.Context, validationToken, newPassword string) (userEmail string, err error) {
	const sql = `
		WITH existing_token AS (
			SELECT
				token,
				is_used,
				user_email,
				expiration
			FROM password_restoration_tokens
			WHERE token = $1
		), valid_email AS (
		    SELECT t.user_email AS email, e.owner_id
		    FROM existing_token t
			LEFT JOIN emails e
		    ON t.user_email = e.email
		    WHERE e.is_active = TRUE
		    AND e.is_confirmed = TRUE		    
		), used_valid_token AS (
			UPDATE password_restoration_tokens
			SET is_used = TRUE
			WHERE token IN
				(SELECT token FROM existing_token
				WHERE is_used = FALSE
				AND user_email IN (SELECT email FROM valid_email)
				AND expiration > CURRENT_TIMESTAMP)
			    
			RETURNING user_email AS email
		), changed_password AS (
			INSERT INTO passwords (user_id, encrypted_password)
			SELECT u.id, $2
			FROM valid_email e
			LEFT JOIN users u ON e.owner_id = u.id
			WHERE e.email IN (SELECT email FROM used_valid_token)
			ON CONFLICT (user_id) DO UPDATE
			SET encrypted_password = EXCLUDED.encrypted_password

			RETURNING user_id
		)
		SELECT
			EXISTS(SELECT 1 FROM changed_password) AS is_done_successfully,
			EXISTS(SELECT 1 FROM valid_email) AS is_email_valid,
			t.token,
			t.is_used,
			t.expiration,
			t.user_email
		FROM (SELECT 1) s
		INNER JOIN existing_token t ON TRUE;
	`

	var (
		wasChanged    bool
		isValidEmail  bool
		existingToken = ConfirmationToken{}
	)

	err = r.pool.QueryRow(ctx, sql, validationToken, newPassword).
		Scan(
			&wasChanged,
			&isValidEmail,
			&existingToken.Value,
			&existingToken.IsUsed,
			&existingToken.ExpirationTime,
			&userEmail,
		)

	if err != nil {
		return "", handleSearchingTokenError(err)
	}

	if !isValidEmail {
		return "", ErrNotConfirmedOrChangedEmail
	}

	if !wasChanged {
		return "", errors.Join(existingToken.InvalidityReason())
	}

	return userEmail, nil
}

func (r PostgresqlRepository) isConfirmed(ctx context.Context, email string) (bool, error) {
	const sql = `
		SELECT EXISTS (
		    SELECT 1 FROM emails
		    WHERE email = $1
		    AND  is_active = TRUE
		    AND is_confirmed = TRUE
		) AS is_email_confirmed;
	`

	var result bool
	if err := r.pool.QueryRow(ctx, sql, email).
		Scan(&result); err != nil {
		return false, postgresql.HandleQueryErr(err)
	}

	return result, nil
}

func (r PostgresqlRepository) saveAppleRefreshToken(ctx context.Context, userID pgtype.UUID, token string) error {
	const sql = `
		INSERT INTO apple_refresh_tokens (token, user_id)
		VALUES ($1, $2)
		ON CONFLICT (token) DO NOTHING;
	`

	_, err := r.pool.Exec(ctx, sql, token, userID)
	return postgresql.HandleQueryErr(err)
}

func (r PostgresqlRepository) allAppleRefreshTokens(ctx context.Context, userID pgtype.UUID) ([]string, error) {
	const sql = `
		SELECT token FROM apple_refresh_tokens
		WHERE user_id = $1;
	`

	rows, err := r.pool.Query(ctx, sql, userID)
	if err != nil {
		return nil, postgresql.HandleQueryErr(err)
	}

	tokens := make([]string, 0)
	for rows.Next() {
		var token string
		if scanError := rows.Scan(&token); scanError != nil {
			err = errors.Join(scanError, err)
			continue
		}

		tokens = append(tokens, token)
	}

	return tokens, postgresql.HandleQueryErr(err)
}

func (r PostgresqlRepository) byOAuth(ctx context.Context, userInputted oauth.User, oAuthServiceOption oAuthMethod) (*User, oauth.LoginResultType, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, oauth.Error, postgresql.HandleQueryErr(err)
	}

	defer tx.Rollback(ctx)

	tableName := oAuthServiceOption()
	if user, err := r.tryFindUserByIDOAuth(ctx, tx, userInputted, tableName); !errors.Is(err, pgx.ErrNoRows) {
		return user, oauth.Login, postgresql.HandleQueryErr(err)
	}

	if user, err := r.tryConnectOAuthUserToRegistered(ctx, tx, userInputted, tableName); !errors.Is(err, pgx.ErrNoRows) {
		return user, oauth.Connect, postgresql.HandleQueryErr(err)
	}

	user, err := r.registerUserWithOAuth(ctx, tx, userInputted, tableName)
	return user, oauth.Register, postgresql.HandleQueryErr(err)
}

func (r PostgresqlRepository) addEmailConfirmationToken(ctx context.Context, email string, tempToken ConfirmationToken) error {
	const sql = `
		INSERT INTO mail_confirmation_tokens (token, email, expiration)
		VALUES ($1, LOWER($2), $3);
	`

	_, err := r.pool.Exec(ctx, sql,
		tempToken.Value,
		email,
		tempToken.ExpirationTime,
	)

	return postgresql.HandleQueryErr(err)
}

func (r PostgresqlRepository) addPasswordResetToken(ctx context.Context, email string, tempToken ConfirmationToken) error {
	const sql = `
		INSERT INTO password_restoration_tokens (token, user_email, expiration)
		VALUES ($1, LOWER($2), $3);
	`

	_, err := r.pool.Exec(ctx, sql,
		tempToken.Value,
		email,
		tempToken.ExpirationTime,
	)

	return postgresql.HandleQueryErr(err)
}

func handleSearchingTokenError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.Join(errTokenNotExists, err)
	}

	return postgresql.HandleQueryErr(err)
}

func (r PostgresqlRepository) tryFindUserByIDOAuth(ctx context.Context, tx pgx.Tx, userInputted oauth.User, tableName serviceTableName) (*User, error) {
	const sqlFormat = `
		SELECT user_id FROM %s
		WHERE id = $1;
	`

	sql := fmt.Sprintf(sqlFormat, tableName)
	var userID pgtype.UUID
	err := tx.QueryRow(ctx, sql, userInputted.ID()).
		Scan(&userID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.byID(ctx, userID)
}

func (r PostgresqlRepository) tryConnectOAuthUserToRegistered(ctx context.Context, tx pgx.Tx, userInputted oauth.User, tableName serviceTableName) (*User, error) {
	const sqlFormat = `
		INSERT INTO %s (user_id, id)
		SELECT owner_id, $1 FROM emails
		WHERE email ILIKE $2
			
		RETURNING user_id;
	`

	sql := fmt.Sprintf(sqlFormat, tableName)
	var userID pgtype.UUID
	err := tx.QueryRow(ctx, sql, userInputted.ID(), userInputted.Email()).
		Scan(&userID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.byID(ctx, userID)
}

func (r PostgresqlRepository) registerUserWithOAuth(ctx context.Context, tx pgx.Tx, userInputted oauth.User, tableName serviceTableName) (*User, error) {
	const sqlFormat = `
		WITH new_user AS (
		    INSERT INTO users (username)
			VALUES ($1)
		
		RETURNING id, username
		), new_email AS (
		    INSERT INTO emails (email, owner_id, is_confirmed)
		    VALUES (lower($2), (SELECT id FROM new_user), $3)
		           
		    RETURNING email, owner_id, is_confirmed
		), oauth_connection AS (
			INSERT INTO %s (user_id, id)
			VALUES ((SELECT id FROM new_user), $4)
		)
		SELECT u.id, u.username, e.email, e.is_confirmed
		FROM new_user u
		LEFT JOIN new_email e ON u.id = e.owner_id;
	`

	sql := fmt.Sprintf(sqlFormat, tableName)
	user := new(User)
	err := tx.QueryRow(ctx, sql, userInputted.Name(), userInputted.Email(), userInputted.IsEmailVerified(), userInputted.ID()).
		Scan(&user.ID, &user.Username, &user.Email, &user.IsEmailConfirmed)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return user, nil
}
