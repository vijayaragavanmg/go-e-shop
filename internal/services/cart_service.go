package services

import (
	"errors"
	"log"

	"github.com/vijayaragavanmg/learning-go-shop/internal/dto"
	"github.com/vijayaragavanmg/learning-go-shop/internal/models"
	"github.com/vijayaragavanmg/learning-go-shop/internal/repositories"
)

var _ CartServiceInterface = (*CartService)(nil)

type CartService struct {
	cartRepo    repositories.CartRepositoryInterface
	productRepo repositories.ProductRepositoryInterface
}

func NewCartService(cartRepo repositories.CartRepositoryInterface,
	productRepo repositories.ProductRepositoryInterface) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *CartService) GetCart(userID uint) (*dto.CartResponse, error) {

	cart, err := s.cartRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	return s.convertToCartResponse(cart), nil
}

func (s *CartService) AddToCart(userID uint, req *dto.AddToCartRequest) (*dto.CartResponse, error) {

	// Check if product exists
	product, err := s.productRepo.GetProductByID(req.ProductID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if product.Stock < req.Quantity {
		return nil, errors.New("insufficient stock")
	}

	// Get or create cart
	// var cart models.Cart
	cart, err := s.cartRepo.GetByUserID(userID)
	if err != nil {
		cart := models.Cart{UserID: userID}

		if err := s.cartRepo.Create(&cart); err != nil {
			return nil, err
		}
	}

	// Check if item already exists in cart

	cartItem, err := s.cartRepo.GetCartItemByCartIDAndProductID(cart.ID, req.ProductID)
	if err != nil {
		// Create new cart item
		cartItem := models.CartItem{
			CartID:    cart.ID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		}
		if err := s.cartRepo.CreateCartItem(&cartItem); err != nil {
			log.Println(err)
			_ = err
		}
	} else {
		// Update existing cart item
		cartItem.Quantity += req.Quantity
		if cartItem.Quantity > product.Stock {
			return nil, errors.New("insufficient stock")
		}

		if err := s.cartRepo.UpdateCartItem(cartItem); err != nil {
			log.Println(err)
			_ = err
		}

	}

	return s.GetCart(userID)
}

func (s *CartService) UpdateCartItem(userID, itemID uint, req *dto.UpdateCartItemRequest) (*dto.CartResponse, error) {

	cartItem, err := s.cartRepo.GetCartItemByCartItemIDAndUserID(itemID, userID)
	if err != nil {
		return nil, errors.New("cart item not found")
	}

	product, err := s.productRepo.GetProductByID(cartItem.ProductID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if product.Stock < req.Quantity {
		return nil, errors.New("insufficient stock")
	}

	cartItem.Quantity = req.Quantity
	if err := s.cartRepo.UpdateCartItem(cartItem); err != nil {
		return nil, err
	}

	return s.GetCart(userID)
}

func (s *CartService) RemoveFromCart(userID, itemID uint) error {
	// return s.db.Joins("JOIN carts ON cart_items.cart_id = carts.id").
	// 	Where("cart_items.id = ? AND carts.user_id = ?", itemID, userID).
	// 	Delete(&models.CartItem{}).Error

	return s.cartRepo.RemoveCartItemFromCart(userID, itemID)
}

func (s *CartService) convertToCartResponse(cart *models.Cart) *dto.CartResponse {

	cartItems := make([]dto.CartItemResponse, len(cart.CartItems)) // memory allocation
	var total float64

	for i := range cart.CartItems {
		subtotal := float64(cart.CartItems[i].Quantity) * cart.CartItems[i].Product.Price
		total += subtotal

		cartItems[i] = dto.CartItemResponse{
			ID: cart.CartItems[i].ID,
			Product: dto.ProductResponse{
				ID:          cart.CartItems[i].Product.ID,
				CategoryID:  cart.CartItems[i].Product.CategoryID,
				Name:        cart.CartItems[i].Product.Name,
				Description: cart.CartItems[i].Product.Description,
				Price:       cart.CartItems[i].Product.Price,
				Stock:       cart.CartItems[i].Product.Stock,
				SKU:         cart.CartItems[i].Product.SKU,
				IsActive:    cart.CartItems[i].Product.IsActive,
				Category: dto.CategoryResponse{
					ID:          cart.CartItems[i].Product.Category.ID,
					Name:        cart.CartItems[i].Product.Category.Name,
					Description: cart.CartItems[i].Product.Category.Description,
					IsActive:    cart.CartItems[i].Product.Category.IsActive,
				},
			},
			Quantity:  cart.CartItems[i].Quantity,
			Subtotal:  subtotal,
			CreatedAt: cart.CartItems[i].CreatedAt,
			UpdatedAt: cart.CartItems[i].UpdatedAt,
		}
	}

	return &dto.CartResponse{
		ID:        cart.ID,
		UserID:    cart.UserID,
		CartItems: cartItems,
		Total:     total,
		CreatedAt: cart.CreatedAt,
		UpdatedAt: cart.UpdatedAt,
	}
}
