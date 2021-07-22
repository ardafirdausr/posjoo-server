package mysql

import (
	"ardafirdausr/posjoo-server/internal/entity"
	"context"
	"database/sql"
	"errors"
	"log"
	"time"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(DB *sql.DB) *UserRepository {
	repo := new(UserRepository)
	repo.DB = DB
	return repo
}

func (repo *UserRepository) GetUserByID(ID int64) (*entity.User, error) {
	ctx := context.TODO()
	row := repo.DB.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?", ID)

	var user entity.User
	var err = row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PhotoUrl,
		&user.Role,
		&user.Position,
		&user.Password,
		&user.MerchantID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		err := entity.ErrNotFound{
			Message: "User not found",
			Err:     err,
		}
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (repo *UserRepository) GetUserByEmail(email string) (*entity.User, error) {
	ctx := context.TODO()
	row := repo.DB.QueryRowContext(ctx, "SELECT * FROM users WHERE email = ?", email)

	var user entity.User
	var err = row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PhotoUrl,
		&user.Role,
		&user.Position,
		&user.Password,
		&user.MerchantID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		err := entity.ErrNotFound{
			Message: "User not found",
			Err:     err,
		}
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (repo *UserRepository) GetUsersByMerchantID(merchantID int64) ([]*entity.User, error) {
	ctx := context.TODO()
	rows, err := repo.DB.QueryContext(ctx, "SELECT * FROM users WHERE merchant_id = ?", merchantID)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	rows.Close()

	users := []*entity.User{}
	for rows.Next() {
		var user entity.User
		var err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.PhotoUrl,
			&user.Role,
			&user.Position,
			&user.Password,
			&user.MerchantID,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			log.Println(err.Error())
			return nil, err
		}

		users = append(users, &user)
	}
	if err = rows.Err(); err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return users, nil
}

func (repo *UserRepository) CreateUser(param entity.CreateUserParam) (*entity.User, error) {
	res, err := repo.DB.Exec(
		"INSERT INTO users(name, email, role, password, merchant_id) VALUES(?, ?, ?, ?, ?)",
		param.Name,
		param.Email,
		param.Role,
		param.Position,
		param.Password,
		param.MerchantID)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	ID, err := res.LastInsertId()
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	user := &entity.User{
		ID:         ID,
		Name:       param.Name,
		Email:      param.Email,
		Role:       param.Role,
		Position:   param.Position,
		Password:   param.Password,
		MerchantID: param.MerchantID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return user, nil
}

func (repo *UserRepository) UpdateByID(ID int64, param entity.UpdateUserParam) error {
	res, err := repo.DB.Exec(
		"UPDATE users SET name = ?, email = ?, role = ?, position = ?, password = ? WHERE id = ?",
		param.Name, param.Email, param.Role, param.Position, param.Password, ID)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	if num, _ := res.RowsAffected(); num < 1 {
		err := errors.New("failed update user")
		enf := entity.ErrNotFound{
			Message: "User not found",
			Err:     err,
		}
		log.Println(enf.Error())
		return enf
	}

	return nil
}

func (repo *UserRepository) DeleteUserByID(ID int64) error {
	res, err := repo.DB.Exec("DELETE FROM users WHERE id = ?", ID)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	if num, _ := res.RowsAffected(); num < 1 {
		err := errors.New("failed update user")
		log.Println(err.Error())
		return err
	}

	return nil
}