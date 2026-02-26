package repositories

import (
	"errors"
	"fmt"

	"github.com/vijayaragavanmg/learning-go-shop/internal/models"
	"gorm.io/gorm"
)

var _ OrderRepositoryInterface = (*OrderRepository)(nil)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// CreateOrder implements OrderRepositoryInterface.
func (o *OrderRepository) CreateOrder(userID uint) (*models.Order, error) {
	var orderResponse *models.Order
	err := o.db.Transaction(func(tx *gorm.DB) error {

		var cart models.Cart
		if err := tx.Preload("CartItems.Product").Where("user_id = ?", userID).First(&cart).Error; err != nil {
			return errors.New("cart not found")
		}

		if len(cart.CartItems) == 0 {
			return errors.New("cart is empty")
		}

		// Calculate total and validate stock
		var totalAmount float64
		var orderItems []models.OrderItem

		for i := range cart.CartItems {
			cartItem := &cart.CartItems[i]

			if cartItem.Product.Stock < cartItem.Quantity {
				return fmt.Errorf("insufficient stock for product: %s", cartItem.Product.Name)
			}

			itemTotal := float64(cartItem.Quantity) * cartItem.Product.Price
			totalAmount += itemTotal

			orderItems = append(orderItems, models.OrderItem{
				ProductID: cartItem.ProductID,
				Quantity:  cartItem.Quantity,
				Price:     itemTotal,
			})

			// Update product stock
			cartItem.Product.Stock -= cartItem.Quantity
			if err := tx.Save(&cartItem.Product).Error; err != nil {
				return err
			}
		}
		// Create order
		order := models.Order{
			UserID:      userID,
			Status:      models.OrderStatusPending,
			TotalAmount: totalAmount,
			OrderItems:  orderItems,
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// Clear cart
		if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}

		if err := tx.Preload("OrderItems.Product.Category").First(&order, order.ID).Error; err != nil {
			return err
		}
		orderResponse = &order
		return nil // Transaction successful
	})
	if err != nil {
		return nil, err
	}
	return orderResponse, nil
}

// GetOrderByUserIDAndOrderID implements OrderRepositoryInterface.
func (o *OrderRepository) GetOrderByUserIDAndOrderID(userID uint, orderID uint) (*models.Order, error) {
	var order models.Order
	if err := o.db.Preload("OrderItems.Product.Category").
		Where("id = ? AND user_id = ?", orderID, userID).
		First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil

}

// GetOrders implements OrderRepositoryInterface.
func (o *OrderRepository) GetOrders(userID uint, offset int, limit int) ([]models.Order, error) {
	var orders []models.Order
	if err := o.db.Preload("OrderItems.Product.Category").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// GetOrdersCount implements OrderRepositoryInterface.
func (o *OrderRepository) GetOrdersCount(userID uint) (int64, error) {
	var total int64

	if err := o.db.Model(&models.Order{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil

}
