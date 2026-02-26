package services

import (
	"log"

	"github.com/vijayaragavanmg/learning-go-shop/internal/dto"
	"github.com/vijayaragavanmg/learning-go-shop/internal/models"
	"github.com/vijayaragavanmg/learning-go-shop/internal/repositories"
	"github.com/vijayaragavanmg/learning-go-shop/internal/utils"
)

var _ OrderServiceInterface = (*OrderService)(nil)

type OrderService struct {
	orderRepo repositories.OrderRepositoryInterface
}

// NewOrderService creates the order service type
func NewOrderService(orderRepo repositories.OrderRepositoryInterface) *OrderService {
	return &OrderService{orderRepo: orderRepo}
}

func (s *OrderService) CreateOrder(userID uint) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.CreateOrder(userID)
	if err != nil {
		return nil, err
	}

	response := s.convertToOrderResponse(order)
	return &response, nil

}

func (s *OrderService) GetOrders(userID uint, page, limit int) ([]dto.OrderResponse, *utils.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	total, err := s.orderRepo.GetOrdersCount(userID)

	if err != nil {
		log.Panicln(err)
		_ = err
	}

	orders, err := s.orderRepo.GetOrders(userID, offset, limit)
	if err != nil {
		return nil, nil, err
	}

	response := make([]dto.OrderResponse, len(orders))
	for i := range orders {
		order := &orders[i]
		response[i] = s.convertToOrderResponse(order)
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

func (s *OrderService) GetOrder(userID, orderID uint) (*dto.OrderResponse, error) {

	order, err := s.orderRepo.GetOrderByUserIDAndOrderID(userID, orderID)
	if err != nil {
		return nil, err
	}

	response := s.convertToOrderResponse(order)

	return &response, nil
}

func (s *OrderService) convertToOrderResponse(order *models.Order) dto.OrderResponse {
	orderItems := make([]dto.OrderItemResponse, len(order.OrderItems))
	for i := range order.OrderItems {
		item := order.OrderItems[i]

		orderItems[i] = dto.OrderItemResponse{
			ID: item.ID,
			Product: dto.ProductResponse{
				ID:          item.Product.ID,
				CategoryID:  item.Product.CategoryID,
				Name:        item.Product.Name,
				Description: item.Product.Description,
				Price:       item.Product.Price,
				Stock:       item.Product.Stock,
				SKU:         item.Product.SKU,
				IsActive:    item.Product.IsActive,
				Category: dto.CategoryResponse{
					ID:          item.Product.Category.ID,
					Name:        item.Product.Category.Name,
					Description: item.Product.Category.Description,
					IsActive:    item.Product.Category.IsActive,
				},
			},
			Quantity:  item.Quantity,
			Price:     item.Price,
			CreatedAt: item.CreatedAt,
		}
	}

	return dto.OrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
		OrderItems:  orderItems,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}
}
