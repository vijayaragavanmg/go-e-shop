package services

import (
	"log"

	"github.com/vijayaragavanmg/learning-go-shop/internal/dto"
	"github.com/vijayaragavanmg/learning-go-shop/internal/models"
	"github.com/vijayaragavanmg/learning-go-shop/internal/repositories"
	"github.com/vijayaragavanmg/learning-go-shop/internal/utils"
)

var _ ProductServiceInterface = (*ProductService)(nil)

type ProductService struct {
	productRepo repositories.ProductRepositoryInterface
}

func NewProductService(productRepo repositories.ProductRepositoryInterface) *ProductService {
	return &ProductService{productRepo: productRepo}
}

func (s *ProductService) CreateCategory(req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {

	category, err := s.productRepo.CreateCategory(req.Name, req.Description)

	if err != nil {
		return nil, err
	}

	return &dto.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		IsActive:    category.IsActive,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}, nil

}

func (s *ProductService) GetCategories() ([]dto.CategoryResponse, error) {

	categories, err := s.productRepo.GetCategoriesByStatus(true)
	if err != nil {
		return nil, err
	}

	response := make([]dto.CategoryResponse, len(categories))
	for i := range categories {
		response[i] = dto.CategoryResponse{
			ID:          categories[i].ID,
			Name:        categories[i].Name,
			Description: categories[i].Description,
			IsActive:    categories[i].IsActive,
			CreatedAt:   categories[i].CreatedAt,
			UpdatedAt:   categories[i].UpdatedAt,
		}
	}

	return response, nil
}

func (s *ProductService) UpdateCategory(id uint, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {

	category, err := s.productRepo.GetCategoriesByID(id)
	if err != nil {
		return nil, err
	}

	category.Name = req.Name
	category.Description = req.Description
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := s.productRepo.UpdateCategory(category); err != nil {
		return nil, err
	}

	return &dto.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		IsActive:    category.IsActive,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}, nil
}

func (s *ProductService) DeleteCategory(id uint) error {
	return s.productRepo.DeleteCategory(id)
}

func (s *ProductService) CreateProduct(req *dto.CreateProductRequest) (*dto.ProductResponse, error) {

	product, err := s.productRepo.CreateProduct(req.CategoryID, req.Name, req.Description, req.Price, req.Stock, req.SKU)
	if err != nil {
		return nil, err
	}

	return s.GetProduct(product.ID)
}

func (s *ProductService) GetProducts(page, limit int) ([]dto.ProductResponse, *utils.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	total, err := s.productRepo.GetProductsCountByStatus(true)
	if err != nil {
		return nil, nil, err
	}

	products, err := s.productRepo.GetProductsByStatus(true, offset, limit)
	if err != nil {
		return nil, nil, err
	}

	response := make([]dto.ProductResponse, len(products))
	for i := range products {
		response[i] = s.convertToProductResponse(&products[i])
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	meta := &utils.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return response, meta, nil
}

func (s *ProductService) GetProduct(id uint) (*dto.ProductResponse, error) {

	product, err := s.productRepo.GetProductByID(id)
	if err != nil {
		return nil, err
	}

	response := s.convertToProductResponse(product)
	return &response, nil
}

func (s *ProductService) UpdateProduct(id uint, req *dto.UpdateProductRequest) (*dto.ProductResponse, error) {

	product, err := s.productRepo.GetProductByID(id)
	if err != nil {
		return nil, err
	}

	product.CategoryID = req.CategoryID
	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := s.productRepo.UpdateProduct(product); err != nil {
		return nil, err
	}

	return s.GetProduct(id)
}

func (s *ProductService) DeleteProduct(id uint) error {
	return s.productRepo.DeleteProduct(id)
}

func (s *ProductService) AddProductImage(productID uint, url, altText string) error {

	count, err := s.productRepo.GetProductImageCount(productID)
	if err != nil {
		log.Println(err)
		_ = err
	}

	return s.productRepo.AddProductImages(productID, url, altText, count == 0)
}

func (s *ProductService) SearchProducts(req *dto.SearchProductsRequest) ([]dto.ProductSearchResult, *utils.PaginationMeta, error) {
	if req.Page < 1 {
		req.Page = 1
	}

	if req.Limit < 1 {
		req.Limit = 10
	}

	offset := (req.Page - 1) * req.Limit

	// build query
	rows, total, err := s.productRepo.SearchProducts(req.Query, req.CategoryID, req.MaxPrice, req.MaxPrice, offset, req.Limit)
	if err != nil {
		return nil, nil, err
	}

	// Build output response
	results := make([]dto.ProductSearchResult, len(rows))
	for i := range rows {
		results[i] = dto.ProductSearchResult{
			ProductResponse: s.convertToProductResponse(&rows[i].Product),
			Rank:            rows[i].Rank,
		}
	}

	// build pagination meta
	totalPages := int((*total + int64(req.Limit) - 1) / int64(req.Limit))
	meta := &utils.PaginationMeta{
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      *total,
		TotalPages: totalPages,
	}

	return results, meta, nil
}

func (s *ProductService) convertToProductResponse(product *models.Product) dto.ProductResponse {
	images := make([]dto.ProductImageResponse, len(product.Images))
	for i := range product.Images {
		images[i] = dto.ProductImageResponse{
			ID:        product.Images[i].ID,
			URL:       product.Images[i].URL,
			AltText:   product.Images[i].AltText,
			IsPrimary: product.Images[i].IsPrimary,
			CreatedAt: product.Images[i].CreatedAt,
		}
	}

	return dto.ProductResponse{
		ID:          product.ID,
		CategoryID:  product.CategoryID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		SKU:         product.SKU,
		IsActive:    product.IsActive,
		Category: dto.CategoryResponse{
			ID:          product.Category.ID,
			Name:        product.Category.Name,
			Description: product.Category.Description,
			IsActive:    product.Category.IsActive,
			CreatedAt:   product.Category.CreatedAt,
			UpdatedAt:   product.Category.UpdatedAt,
		},
		Images:    images,
		CreatedAt: product.Category.CreatedAt,
		UpdatedAt: product.Category.UpdatedAt,
	}
}
