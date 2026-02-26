package repositories

import "github.com/vijayaragavanmg/learning-go-shop/internal/models"

type UserRepositoryInterface interface {
	GetByEmail(email string) (*models.User, error)
	GetByID(id uint) (*models.User, error)
	GetByEmailAndActive(email string, isActive bool) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(id uint) error

	CreateRefreshToken(token *models.RefreshToken) error
	GetValidRefreshToken(token string) (*models.RefreshToken, error)
	DeleteRefreshToken(token string) error
	DeleteRefreshTokenByID(id uint) error
}

type CartRepositoryInterface interface {
	GetByUserID(userID uint) (*models.Cart, error)
	Create(cart *models.Cart) error
	Update(cart *models.Cart) error
	Delete(id uint) error

	GetCartItemByCartIDAndProductID(cartID, productID uint) (*models.CartItem, error)
	GetCartItemByCartItemIDAndUserID(cartItemID, userID uint) (*models.CartItem, error)
	CreateCartItem(cartItem *models.CartItem) error
	UpdateCartItem(cartItem *models.CartItem) error
	RemoveCartItemFromCart(userID, cartItemID uint) error
}

type ProductRepositoryInterface interface {
	CreateCategory(name, description string) (*models.Category, error)
	GetCategoriesByID(id uint) (*models.Category, error)
	GetCategoriesByStatus(is_active bool) ([]models.Category, error)
	UpdateCategory(category *models.Category) error
	DeleteCategory(id uint) error

	CreateProduct(categoryID uint, name string, description string, price float64, stock int, sku string) (*models.Product, error)
	GetProductByID(id uint) (*models.Product, error)
	GetProductsByStatus(is_active bool, offset, limit int) ([]models.Product, error)
	GetProductsCountByStatus(is_active bool) (int64, error)
	UpdateProduct(product *models.Product) error
	DeleteProduct(id uint) error
	AddProductImages(productID uint, url string, altText string, isPrimary bool) error
	GetProductImageCount(productID uint) (int64, error)
	SearchProducts(queryString string, categoryID *uint, minPrice *float64, maxPrice *float64, offset int, limit int) ([]models.ProductsWithRank, *int64, error)
}

type OrderRepositoryInterface interface {
	CreateOrder(userID uint) (*models.Order, error)
	GetOrderByUserIDAndOrderID(userID, orderID uint) (*models.Order, error)
	GetOrders(userID uint, offset, limit int) ([]models.Order, error)
	GetOrdersCount(userID uint) (int64, error)
}
