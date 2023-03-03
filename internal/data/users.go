package data

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrRecordNotFound = errors.New("record not found")
	ErrBadPassword    = errors.New("bad password")
)

type UserLayer struct {
	DB *pgxpool.Pool
}

type password struct {
	plaintext *string
	hash      []byte
}

var AnonymousUser = &User{}

type User struct {
	User_id    int       `json:"user_id"`
	Created_at time.Time `json:"created_at"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Password   password  `json:"-"`
	Activated  bool      `json:"activated"`
	Is_admin   bool      `json:"is_admin"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, ErrBadPassword
		default:
			return false, err
		}
	}

	return true, nil
}

func (u UserLayer) Insert(user *User) error {
	query := `
		INSERT INTO users (username, email, password_hash, activated, is_admin)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING user_id, created_at
	`
	args := []interface{}{user.Username, user.Email, user.Password.hash, user.Activated, user.Is_admin}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRow(ctx, query, args...).Scan(&user.User_id, &user.Created_at)
	if err != nil {
		switch {
		case err.Error() == `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (u UserLayer) Update(user *User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, password_hash = $3, activated = $4, is_admin = $5
		WHERE user_id = $6
	`

	args := []interface{}{
		user.Username,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.Is_admin,
		user.User_id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := u.DB.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (u UserLayer) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
		SELECT users.user_id, users.created_at, users.username, users.email, users.password_hash, users.activated, users.is_admin
		FROM users
		INNER JOIN tokens
		ON users.user_id = tokens.user_id
		WHERE tokens.hash = $1
		AND tokens.scope = $2
		AND tokens.expiry > $3
	`

	args := []interface{}{tokenHash[:], tokenScope, time.Now()}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user := User{}

	err := u.DB.QueryRow(ctx, query, args...).Scan(
		&user.User_id,
		&user.Created_at,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Is_admin,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u UserLayer) GetByEmail(email string) (*User, error) {
	query := `
		SELECT user_id, created_at, username, email, password_hash, activated, is_admin
		FROM users
		WHERE email = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user := User{}

	err := u.DB.QueryRow(ctx, query, email).Scan(
		&user.User_id,
		&user.Created_at,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Is_admin,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u UserLayer) GetById(id int) (*User, error) {
	query := `
		SELECT user_id, created_at, username, email, password_hash, activated, is_admin
		FROM users
		WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user := User{}

	err := u.DB.QueryRow(ctx, query, id).Scan(
		&user.User_id,
		&user.Created_at,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Is_admin,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u UserLayer) GetAll() ([]User, error) {
	query := `
		SELECT user_id, username, email, created_at, activated, is_admin
		FROM users
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := u.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}

	for rows.Next() {
		user := User{}
		err := rows.Scan(
			&user.User_id,
			&user.Username,
			&user.Email,
			&user.Created_at,
			&user.Activated,
			&user.Is_admin,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (u UserLayer) Delete(id int) error {
	query := `
		DELETE FROM users
		WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := u.DB.Exec(ctx, query, id)
	fmt.Println(err)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			fmt.Println(err)
			return ErrRecordNotFound
		default:
			fmt.Println(err)
			return err
		}
	}

	return nil
}
