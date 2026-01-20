// Package repository provides unit tests for data access layer
// Complies with CODING_STANDARDS.md: Test file max 500 lines, 80%+ coverage
package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"

	"github.com/divyang-garg/sentinel-hub-api/internal/models"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo *PostgresUserRepository
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	var err error
	suite.db, suite.mock, err = sqlmock.New()
	suite.Require().NoError(err)

	suite.repo = NewPostgresUserRepository(suite.db)
}

func (suite *UserRepositoryTestSuite) TearDownTest() {
	suite.db.Close()
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (suite *UserRepositoryTestSuite) TestCreate_Success() {
	user := &models.User{
		Email:     "test@example.com",
		Name:      "Test User",
		Password:  "hashedpassword",
		Role:      models.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	suite.mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Email, user.Name, user.Password, user.Role, user.IsActive, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	result, err := suite.repo.Create(context.Background(), user)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(1, result.ID)
	suite.Equal(user.Email, result.Email)
	suite.Equal(user.Name, result.Name)
}

func (suite *UserRepositoryTestSuite) TestCreate_DatabaseError() {
	user := &models.User{
		Email:     "test@example.com",
		Name:      "Test User",
		Password:  "hashedpassword",
		Role:      models.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	suite.mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Email, user.Name, user.Password, user.Role, user.IsActive, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	result, err := suite.repo.Create(context.Background(), user)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "failed to create user")
}

func (suite *UserRepositoryTestSuite) TestGetByID_Success() {
	expectedUser := &models.User{
		ID:        1,
		Email:     "test@example.com",
		Name:      "Test User",
		Password:  "hashedpassword",
		Role:      models.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "email", "name", "password", "role", "is_active", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Name, expectedUser.Password,
			expectedUser.Role, expectedUser.IsActive, expectedUser.CreatedAt, expectedUser.UpdatedAt)

	suite.mock.ExpectQuery(`SELECT .* FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(rows)

	user, err := suite.repo.GetByID(context.Background(), 1)

	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal(expectedUser.ID, user.ID)
	suite.Equal(expectedUser.Email, user.Email)
	suite.Equal(expectedUser.Name, user.Name)
	suite.Equal(expectedUser.Password, user.Password)
}

func (suite *UserRepositoryTestSuite) TestGetByID_NotFound() {
	suite.mock.ExpectQuery(`SELECT .* FROM users WHERE id = \$1`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	user, err := suite.repo.GetByID(context.Background(), 999)

	suite.Error(err)
	suite.Nil(user)
	var notFoundErr *models.NotFoundError
	suite.ErrorAs(err, &notFoundErr)
	suite.Equal("user", notFoundErr.Resource)
	suite.Equal(999, notFoundErr.ID)
}

func (suite *UserRepositoryTestSuite) TestGetByEmail_Success() {
	expectedUser := &models.User{
		ID:        1,
		Email:     "test@example.com",
		Name:      "Test User",
		Password:  "hashedpassword",
		Role:      models.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "email", "name", "password", "role", "is_active", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Name, expectedUser.Password,
			expectedUser.Role, expectedUser.IsActive, expectedUser.CreatedAt, expectedUser.UpdatedAt)

	suite.mock.ExpectQuery(`SELECT .* FROM users WHERE email = \$1`).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	user, err := suite.repo.GetByEmail(context.Background(), "test@example.com")

	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal(expectedUser.Email, user.Email)
	suite.Equal(expectedUser.Name, user.Name)
}

func (suite *UserRepositoryTestSuite) TestUpdate_Success() {
	user := &models.User{
		ID:       1,
		Email:    "updated@example.com",
		Name:     "Updated User",
		Password: "newhashedpassword",
		Role:     models.RoleAdmin,
		IsActive: false,
	}

	suite.mock.ExpectExec(`UPDATE users SET name = \$2, email = \$3, password = \$4, role = \$5, is_active = \$6, updated_at = \$7 WHERE id = \$1`).
		WithArgs(user.ID, user.Name, user.Email, user.Password, user.Role, user.IsActive, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := suite.repo.Update(context.Background(), user)

	suite.NoError(err)
}

func (suite *UserRepositoryTestSuite) TestUpdate_NotFound() {
	user := &models.User{
		ID:   999,
		Name: "Non-existent User",
	}

	suite.mock.ExpectExec(`UPDATE users .* WHERE id = \$1`).
		WithArgs(999, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := suite.repo.Update(context.Background(), user)

	suite.Error(err)
	var notFoundErr *models.NotFoundError
	suite.ErrorAs(err, &notFoundErr)
	suite.Equal("user", notFoundErr.Resource)
	suite.Equal(999, notFoundErr.ID)
}

func (suite *UserRepositoryTestSuite) TestDelete_Success() {
	suite.mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := suite.repo.Delete(context.Background(), 1)

	suite.NoError(err)
}

func (suite *UserRepositoryTestSuite) TestDelete_NotFound() {
	suite.mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := suite.repo.Delete(context.Background(), 999)

	suite.Error(err)
	var notFoundErr *models.NotFoundError
	suite.ErrorAs(err, &notFoundErr)
	suite.Equal("user", notFoundErr.Resource)
	suite.Equal(999, notFoundErr.ID)
}

func (suite *UserRepositoryTestSuite) TestList_Success() {
	users := []*models.User{
		{
			ID:        1,
			Email:     "user1@example.com",
			Name:      "User One",
			Password:  "hash1",
			Role:      models.RoleUser,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        2,
			Email:     "user2@example.com",
			Name:      "User Two",
			Password:  "hash2",
			Role:      models.RoleAdmin,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	rows := sqlmock.NewRows([]string{"id", "email", "name", "password", "role", "is_active", "created_at", "updated_at"})
	for _, user := range users {
		rows.AddRow(user.ID, user.Email, user.Name, user.Password, user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt)
	}

	suite.mock.ExpectQuery(`SELECT .* FROM users ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	result, err := suite.repo.List(context.Background(), 10, 0)

	suite.NoError(err)
	suite.Len(result, 2)
	suite.Equal(users[0].Email, result[0].Email)
	suite.Equal(users[1].Email, result[1].Email)
}

func (suite *UserRepositoryTestSuite) TestGetByEmail_DatabaseError() {
	// Given: Database error (non-NotFound)
	suite.mock.ExpectQuery(`SELECT .* FROM users WHERE email = \$1`).
		WithArgs("test@example.com").
		WillReturnError(sql.ErrConnDone)

	// When: Getting user by email
	user, err := suite.repo.GetByEmail(context.Background(), "test@example.com")

	// Then: Should return error
	suite.Error(err)
	suite.Nil(user)
	suite.Contains(err.Error(), "failed to get user by email")
}

func (suite *UserRepositoryTestSuite) TestUpdate_DatabaseError() {
	// Given: User to update but database error occurs
	user := &models.User{
		ID:       1,
		Email:    "updated@example.com",
		Name:     "Updated User",
		Password: "newhashedpassword",
		Role:     models.RoleAdmin,
		IsActive: false,
	}

	suite.mock.ExpectExec(`UPDATE users SET name = \$2, email = \$3, password = \$4, role = \$5, is_active = \$6, updated_at = \$7 WHERE id = \$1`).
		WithArgs(user.ID, user.Name, user.Email, user.Password, user.Role, user.IsActive, sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	// When: Updating user
	err := suite.repo.Update(context.Background(), user)

	// Then: Should return error
	suite.Error(err)
	suite.Contains(err.Error(), "failed to update user")
}

func (suite *UserRepositoryTestSuite) TestUpdate_RowsAffectedError() {
	// Given: User to update but error getting rows affected
	user := &models.User{
		ID:       1,
		Email:    "updated@example.com",
		Name:     "Updated User",
		Password: "newhashedpassword",
		Role:     models.RoleAdmin,
		IsActive: false,
	}

	result := sqlmock.NewErrorResult(errors.New("rows affected error"))
	suite.mock.ExpectExec(`UPDATE users SET name = \$2, email = \$3, password = \$4, role = \$5, is_active = \$6, updated_at = \$7 WHERE id = \$1`).
		WithArgs(user.ID, user.Name, user.Email, user.Password, user.Role, user.IsActive, sqlmock.AnyArg()).
		WillReturnResult(result)

	// When: Updating user
	err := suite.repo.Update(context.Background(), user)

	// Then: Should return error
	suite.Error(err)
	suite.Contains(err.Error(), "failed to get rows affected")
}

func (suite *UserRepositoryTestSuite) TestDelete_DatabaseError() {
	// Given: Database error during delete
	suite.mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	// When: Deleting user
	err := suite.repo.Delete(context.Background(), 1)

	// Then: Should return error
	suite.Error(err)
	suite.Contains(err.Error(), "failed to delete user")
}

func (suite *UserRepositoryTestSuite) TestDelete_RowsAffectedError() {
	// Given: Error getting rows affected
	result := sqlmock.NewErrorResult(errors.New("rows affected error"))
	suite.mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(result)

	// When: Deleting user
	err := suite.repo.Delete(context.Background(), 1)

	// Then: Should return error
	suite.Error(err)
	suite.Contains(err.Error(), "failed to get rows affected")
}

func (suite *UserRepositoryTestSuite) TestList_DatabaseError() {
	// Given: Database error during query
	suite.mock.ExpectQuery(`SELECT .* FROM users ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnError(sql.ErrConnDone)

	// When: Listing users
	result, err := suite.repo.List(context.Background(), 10, 0)

	// Then: Should return error
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "failed to list users")
}

func (suite *UserRepositoryTestSuite) TestList_ScanError() {
	// Given: Rows with invalid data causing scan error
	rows := sqlmock.NewRows([]string{"id", "email", "name", "password", "role", "is_active", "created_at", "updated_at"}).
		AddRow("invalid", "user1@example.com", "User One", "hash1", "user", true, time.Now(), time.Now())

	suite.mock.ExpectQuery(`SELECT .* FROM users ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	// When: Listing users
	result, err := suite.repo.List(context.Background(), 10, 0)

	// Then: Should return error
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "failed to scan user")
}

func (suite *UserRepositoryTestSuite) TestList_RowsError() {
	// Given: Rows iteration error
	rows := sqlmock.NewRows([]string{"id", "email", "name", "password", "role", "is_active", "created_at", "updated_at"}).
		AddRow(1, "user1@example.com", "User One", "hash1", "user", true, time.Now(), time.Now()).
		RowError(0, errors.New("row iteration error"))

	suite.mock.ExpectQuery(`SELECT .* FROM users ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	// When: Listing users
	result, err := suite.repo.List(context.Background(), 10, 0)

	// Then: Should return error
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "error iterating users")
}
