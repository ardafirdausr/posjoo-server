package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/ardafirdausr/posjoo-server/internal"
	"github.com/ardafirdausr/posjoo-server/internal/entity"
)

type UserUsecase struct {
	userRepo internal.UserRepository
	storage  internal.Storage
}

func NewUserUsecase(userRepo internal.UserRepository, storage internal.Storage) *UserUsecase {
	usecase := new(UserUsecase)
	usecase.userRepo = userRepo
	usecase.storage = storage
	return usecase
}

func (uc UserUsecase) GetMerchantUsers(ctx context.Context, merchantID int64) ([]*entity.User, error) {
	users, err := uc.userRepo.GetUsersByMerchantID(ctx, merchantID)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return users, nil
}

func (uc UserUsecase) GetUser(ctx context.Context, userID int64) (*entity.User, error) {
	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return user, nil
}

func (uc UserUsecase) CreateUser(ctx context.Context, param entity.CreateUserParam) (*entity.User, error) {
	existUser, err := uc.userRepo.GetUserByEmail(ctx, param.Email)
	_, errNotFound := err.(entity.ErrNotFound)
	if err != nil && !errNotFound {
		return nil, err
	}

	if existUser != nil && existUser.Email == param.Email {
		err := entity.ErrInvalidData{
			Message: "Email is already registered",
			Err:     errors.New("email is already registered"),
		}
		return nil, err
	}

	isRoleAllowed := false
	allowedRoles := []entity.UserRole{entity.UserRoleManager, entity.UserRoleStaff}
	for _, role := range allowedRoles {
		if param.Role == role {
			isRoleAllowed = true
		}
	}
	if !isRoleAllowed {
		err := entity.ErrInvalidData{
			Message: "Cannot create user with requested role",
			Err:     errors.New("cannot create user with requested rol"),
		}
		return nil, err
	}

	param.Password = hashString(param.Password)
	param.CreatedAt = time.Now()
	user, err := uc.userRepo.CreateUser(ctx, param)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (uc UserUsecase) UpdateUser(ctx context.Context, userID int64, param entity.UpdateUserParam) (*entity.User, error) {
	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	existUser, err := uc.userRepo.GetUserByEmail(ctx, param.Email)
	_, errNotFound := err.(entity.ErrNotFound)
	if err != nil && !errNotFound {
		return nil, err
	}

	if existUser != nil && existUser.ID != user.ID && existUser.Email == param.Email {
		err := entity.ErrInvalidData{
			Message: "Email is already registered",
			Err:     errors.New("email is already registered"),
		}
		return nil, err
	}

	ownerTriedToChangeRole := user.Role == entity.UserRoleOwner && param.Role != entity.UserRoleOwner
	ownerTriedToMakeOtherOwner := user.Role != entity.UserRoleOwner && param.Role == entity.UserRoleOwner
	if ownerTriedToChangeRole && ownerTriedToMakeOtherOwner {
		err := entity.ErrForbidden{
			Message: "Cannot change owner or create new orner",
			Err:     errors.New("cannot change owner or create new orner"),
		}
		return nil, err
	}

	isRoleAllowed := false
	allowedRoles := []entity.UserRole{entity.UserRoleManager, entity.UserRoleStaff}
	for _, role := range allowedRoles {
		if param.Role == role {
			isRoleAllowed = true
		}
	}
	if !isRoleAllowed {
		err := entity.ErrInvalidData{
			Message: "Cannot create user with requested role",
			Err:     errors.New("cannot create user with requested rol"),
		}
		return nil, err
	}

	param.UpdatedAt = time.Now()
	err = uc.userRepo.UpdateUserByID(ctx, userID, param)
	if _, isEnf := err.(entity.ErrNotFound); isEnf {
		return user, nil
	}

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return uc.userRepo.GetUserByID(ctx, userID)
}

func (uc UserUsecase) UpdateUserPassword(ctx context.Context, userID int64, param entity.UpdateUserPasswordParam) error {
	hashPassword := hashString(param.Password)
	if err := uc.userRepo.UpdateUserPasswordByID(ctx, userID, hashPassword); err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

func (uc UserUsecase) UpdateUserPhoto(ctx context.Context, userID int64, photo *multipart.FileHeader) (*entity.User, error) {
	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	if photo == nil {
		return nil, entity.ErrInvalidData{
			Message: "photo cannot be empty",
			Err:     errors.New("photo cannot be empty"),
		}
	}

	rule := map[string]int64{
		".jpg":  1024 * 1000 * 4,
		".jpeg": 1024 * 1000 * 4,
		".png":  1024 * 1000 * 4,
	}
	ext := strings.ToLower(filepath.Ext(photo.Filename))
	maxSize, ok := rule[ext]
	if !ok {
		return nil, entity.ErrInvalidData{
			Message: "photo extension must be .jpg, .jpeg, or .png",
			Err:     errors.New("photo extension must be .jpg, .jpeg, or .png"),
		}
	}

	if photo.Size > maxSize {
		return nil, entity.ErrInvalidData{
			Message: "Max photo size is 4MB",
			Err:     errors.New("max photo size is 4MB"),
		}
	}

	photoName := fmt.Sprintf("user-%d", user.ID)
	photoExt := filepath.Ext(photo.Filename)
	filename := photoName + photoExt
	photoDirectory := filepath.Join("image", "user")
	url, err := uc.storage.Save(photo, photoDirectory, filename)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	err = uc.userRepo.UpdateUserPhotoByID(ctx, userID, url)
	if _, isEnf := err.(entity.ErrNotFound); isEnf {
		return user, nil
	}

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return uc.userRepo.GetUserByID(ctx, userID)
}

func (uc UserUsecase) DeleteUser(ctx context.Context, userID int64) error {
	if err := uc.userRepo.DeleteUserByID(ctx, userID); err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}
