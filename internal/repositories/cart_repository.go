package repositories

import (
	"errors"

	"github.com/vijayaragavanmg/learning-go-shop/internal/models"
	"gorm.io/gorm"
)

var _ CartRepositoryInterface = (*CartRepository)(nil)

type CartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) *CartRepository {
	return &CartRepository{
		db: db,
	}
}

func (c *CartRepository) GetByUserID(userID uint) (*models.Cart, error) {
	var cart models.Cart

	if err := c.db.Preload("CartItems.Product.Category").Where("user_id = ?",
		userID).First(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

func (c *CartRepository) Create(cart *models.Cart) error {
	return c.db.Create(&cart).Error
}

func (c *CartRepository) Update(cart *models.Cart) error {
	return c.db.Save(&cart).Error
}

func (c *CartRepository) Delete(id uint) error {
	return c.db.Delete(&models.Cart{}, id).Error
}

func (c *CartRepository) GetCartItemByCartIDAndProductID(cartID, productID uint) (*models.CartItem, error) {
	var cartItem models.CartItem
	if err := c.db.Where("cart_id = ? AND product_id = ?", cartID, productID).First(&cartItem).Error; err != nil {
		return nil, err
	}
	return &cartItem, nil
}

func (c *CartRepository) GetCartItemByCartItemIDAndUserID(cartItemID, userID uint) (*models.CartItem, error) {
	var cartItem models.CartItem
	if err := c.db.Joins("JOIN carts ON cart_items.cart_id = carts.id").
		Where("cart_items.id = ? AND carts.user_id = ?", cartItemID, userID).
		First(&cartItem).Error; err != nil {
		return nil, errors.New("cart item not found")
	}
	return &cartItem, nil
}

func (c *CartRepository) CreateCartItem(cartItem *models.CartItem) error {
	return c.db.Create(&cartItem).Error
}

func (c *CartRepository) UpdateCartItem(cartItem *models.CartItem) error {
	return c.db.Save(&cartItem).Error
}

func (c *CartRepository) RemoveCartItemFromCart(userID, cartItemID uint) error {
	return c.db.Unscoped().Where("id = ? AND cart_id IN (?)", cartItemID,
		c.db.Select("id").Table("carts").
			Where("user_id = ?", userID)).
		Delete(&models.CartItem{}).Error
}
