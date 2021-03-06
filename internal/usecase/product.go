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

type ProductUsecase struct {
	ProductRepo internal.ProductRepository
	storage     internal.Storage
}

func NewProductUsecase(ProductRepo internal.ProductRepository, storage internal.Storage) *ProductUsecase {
	usecase := new(ProductUsecase)
	usecase.ProductRepo = ProductRepo
	usecase.storage = storage
	return usecase
}

func (uc ProductUsecase) GetMerchantProducts(ctx context.Context, merchantID int64) ([]*entity.Product, error) {
	products, err := uc.ProductRepo.GetProductsByMerchantID(ctx, merchantID)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return products, err
}

func (uc ProductUsecase) GetProduct(ctx context.Context, productID int64) (*entity.Product, error) {
	product, err := uc.ProductRepo.GetProductByID(ctx, productID)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return product, err
}

func (uc ProductUsecase) CreateProduct(ctx context.Context, param entity.CreateProductParam) (*entity.Product, error) {
	existProduct, err := uc.ProductRepo.GetProductBySKUIndex(ctx, param.MerchantID, param.SKU)
	_, errNotFound := err.(entity.ErrNotFound)
	if err != nil && !errNotFound {
		log.Println(err.Error())
		return nil, err
	}

	isExistProductOfMerchant := existProduct != nil && existProduct.SKU == param.SKU
	if isExistProductOfMerchant {
		err := entity.ErrInvalidData{
			Message: "SKU is already registered",
			Err:     errors.New("SKU is already registered"),
		}
		log.Println(err.Error())
		return nil, err
	}

	param.CreatedAt = time.Now()
	product, err := uc.ProductRepo.CreateProduct(ctx, param)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return product, nil
}

func (uc ProductUsecase) UpdateProduct(ctx context.Context, productID int64, param entity.UpdatedProductparam) (*entity.Product, error) {
	product, err := uc.ProductRepo.GetProductByID(ctx, productID)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	existProduct, err := uc.ProductRepo.GetProductBySKUIndex(ctx, product.MerchantID, param.SKU)
	_, errNotFound := err.(entity.ErrNotFound)
	if err != nil && !errNotFound {
		log.Println(err.Error())
		return nil, err
	}

	isExistProductOfMerchant := existProduct != nil && existProduct.ID != product.ID && existProduct.SKU == param.SKU
	if isExistProductOfMerchant {
		err := entity.ErrInvalidData{
			Message: "SKU is already registered",
			Err:     errors.New("SKU is already registered"),
		}
		log.Println(err.Error())
		return nil, err
	}

	param.UpdatedAt = time.Now()
	err = uc.ProductRepo.UpdateProductByID(ctx, productID, param)
	if _, isEnf := err.(entity.ErrNotFound); isEnf {
		return product, nil
	}

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return uc.ProductRepo.GetProductByID(ctx, productID)
}

func (uc ProductUsecase) UpdateProductPhoto(ctx context.Context, productID int64, photo *multipart.FileHeader) (*entity.Product, error) {
	product, err := uc.ProductRepo.GetProductByID(ctx, productID)
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

	photoName := fmt.Sprintf("product-%d", product.ID)
	photoExt := filepath.Ext(photo.Filename)
	filename := photoName + photoExt
	photoDirectory := filepath.Join("image", "product")
	url, err := uc.storage.Save(photo, photoDirectory, filename)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	err = uc.ProductRepo.UpdateProductPhotoByID(ctx, productID, url)
	if _, isEnf := err.(entity.ErrNotFound); isEnf {
		return product, nil
	}

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return uc.ProductRepo.GetProductByID(ctx, productID)
}

func (uc ProductUsecase) DeleteProduct(ctx context.Context, productID int64) error {
	if err := uc.ProductRepo.DeleteProductByID(ctx, productID); err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}
