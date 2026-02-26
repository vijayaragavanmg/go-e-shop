package repositories

import (
	"github.com/vijayaragavanmg/learning-go-shop/internal/models"
	"gorm.io/gorm"
)

var _ ProductRepositoryInterface = (*ProductRepository)(nil)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

func (p *ProductRepository) CreateCategory(name, description string) (*models.Category, error) {
	category := models.Category{
		Name:        name,
		Description: description,
	}

	if err := p.db.Create(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (p *ProductRepository) GetCategoriesByID(id uint) (*models.Category, error) {
	var category models.Category
	if err := p.db.First(&category, id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (p *ProductRepository) GetCategoriesByStatus(is_active bool) ([]models.Category, error) {

	var categories []models.Category
	if err := p.db.Where("is_active = ?", is_active).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (p *ProductRepository) UpdateCategory(category *models.Category) error {

	return p.db.Save(category).Error
}

func (p *ProductRepository) DeleteCategory(id uint) error {
	return p.db.Delete(&models.Category{}, id).Error

}

func (p *ProductRepository) CreateProduct(categoryID uint, name string, description string, price float64, stock int, sku string) (*models.Product, error) {
	product := models.Product{
		CategoryID:  categoryID,
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		SKU:         sku,
	}

	if err := p.db.Create(&product).Error; err != nil {
		return nil, err
	}

	return &product, nil

}

func (p *ProductRepository) GetProductByID(id uint) (*models.Product, error) {
	var product models.Product

	if err := p.db.Preload("Category").Preload("Images").First(&product, id).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

func (p *ProductRepository) GetProductsByStatus(is_active bool, offset, limit int) ([]models.Product, error) {
	var products []models.Product
	if err := p.db.Preload("Category").Preload("Images").
		Where("is_active = ?", true).
		Offset(offset).Limit(limit).
		Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (p *ProductRepository) GetProductsCountByStatus(is_active bool) (int64, error) {
	var total int64
	if err := p.db.Model(&models.Product{}).Where("is_active = ?", true).Count(&total).Error; err != nil {
		return 0, err
	}
	return int64(total), nil

}
func (p *ProductRepository) UpdateProduct(product *models.Product) error {
	return p.db.Save(&product).Error

}
func (p *ProductRepository) DeleteProduct(id uint) error {
	return p.db.Delete(&models.Product{}, id).Error

}
func (p *ProductRepository) AddProductImages(productID uint, url string, altText string, isPrimary bool) error {
	image := models.ProductImage{
		ProductID: productID,
		URL:       url,
		AltText:   altText,
		IsPrimary: isPrimary,
	}

	return p.db.Create(&image).Error

}
func (p *ProductRepository) GetProductImageCount(productID uint) (int64, error) {
	var count int64
	if err := p.db.Model(&models.ProductImage{}).Where("product_id = ?", productID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int64(count), nil

}

func (p *ProductRepository) SearchProducts(queryString string, categoryID *uint, minPrice *float64, maxPrice *float64, offset int, limit int) ([]models.ProductsWithRank, *int64, error) {
	query := p.db.Model(&models.Product{}).
		Select("products.*, ts_rank(search_vector, plainto_tsquery('english', ?)) as rank", queryString).
		Where("search_vector @@ plainto_tsquery('english', ?)", queryString).
		Where("is_active = ?", true)

	if categoryID != nil {
		query = query.Where("category_id = ?", categoryID)
	}

	if minPrice != nil {
		query = query.Where("price >= ?", minPrice)
	}

	if maxPrice != nil {
		query = query.Where("price <= ?", maxPrice)
	}

	// Count total results
	var total int64
	query.Count(&total)

	var rows []models.ProductsWithRank
	if err := query.
		Order("rank DESC, created_at DESC"). // order by relevance
		Preload("Category").
		Preload("Images").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, nil, err
	}
	return rows, &total, nil
}
