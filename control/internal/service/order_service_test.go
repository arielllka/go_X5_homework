package service_test

import (
	"context"
	"testing"
	"time"

	"example.com/order-service/internal/order"
	"example.com/order-service/internal/service"
	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	orders []order.Order
}

func (r *fakeRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	value.ID = int64(len(r.orders) + 1)
	r.orders = append(r.orders, value)
	return value, nil
}

func (r *fakeRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	for _, value := range r.orders {
		if value.ID == id {
			return value, nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func (r *fakeRepository) List(_ context.Context) ([]order.Order, error) {
	return r.orders, nil
}

func (r *fakeRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	for i := range r.orders {
		if r.orders[i].ID == value.ID {
			r.orders[i] = value
			return value, nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func TestOrderService(t *testing.T) {
	tests := []struct {
		name  string
		items []order.Item
		valid bool
	}{
		{"valid", []order.Item{{Name: "Book", Quantity: 2, Price: 100}, {Name: "Pen", Quantity: 3, Price: 50}}, true},
		{"no_items", nil, false},
		{"empty_name", []order.Item{{Name: "", Quantity: 1, Price: 100}}, false},
		{"blank_name", []order.Item{{Name: "   ", Quantity: 1, Price: 100}}, false},
		{"zero_quantity", []order.Item{{Name: "Book", Quantity: 0, Price: 100}}, false},
		{"negative_quantity", []order.Item{{Name: "Book", Quantity: -1, Price: 100}}, false},
		{"zero_price", []order.Item{{Name: "Book", Quantity: 1, Price: 0}}, false},
		{"negative_price", []order.Item{{Name: "Book", Quantity: 1, Price: -1}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{}
			s := service.NewOrderService(repo)
			value, err := s.Create(context.Background(), order.CreateOrderInput{Items: tt.items})
			if !tt.valid {
				require.ErrorIs(t, err, order.ErrInvalidOrder)
				require.Empty(t, repo.orders)
				return
			}
			require.NoError(t, err)
			require.Equal(t, int64(1), value.ID)
			require.Equal(t, int64(350), value.TotalPrice)
			require.Equal(t, order.StatusNew, value.Status)
			require.NotZero(t, value.CreatedAt)
			require.Len(t, value.History, 1)
			require.Equal(t, order.StatusNew, value.History[0].Status)
			require.Equal(t, value.CreatedAt, value.History[0].ChangedAt)
		})
	}
}

func TestGetByIDAndList(t *testing.T) {
	repo := &fakeRepository{orders: []order.Order{{ID: 1}, {ID: 2}}}
	s := service.NewOrderService(repo)
	value, err := s.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), value.ID)
	_, err = s.GetByID(context.Background(), 99)
	require.ErrorIs(t, err, order.ErrOrderNotFound)
	values, err := s.List(context.Background())
	require.NoError(t, err)
	require.Equal(t, repo.orders, values)
}

func TestUpdateStatus(t *testing.T) {
	tests := []struct {
		from order.Status
		to   order.Status
		err  error
	}{
		{order.StatusNew, order.StatusProcessing, nil},
		{order.StatusNew, order.StatusCancelled, nil},
		{order.StatusProcessing, order.StatusCompleted, nil},
		{order.StatusProcessing, order.StatusCancelled, nil},
		{order.StatusNew, order.StatusCompleted, order.ErrInvalidStatusTransition},
		{order.StatusProcessing, order.StatusNew, order.ErrInvalidStatusTransition},
		{order.StatusCompleted, order.StatusProcessing, order.ErrInvalidStatusTransition},
		{order.StatusCancelled, order.StatusProcessing, order.ErrInvalidStatusTransition},
		{order.StatusNew, "unknown", order.ErrInvalidStatus},
	}
	for _, tt := range tests {
		t.Run(string(tt.from)+"_to_"+string(tt.to), func(t *testing.T) {
			initial := order.StatusChange{Status: tt.from, ChangedAt: time.Now().Add(-time.Minute)}
			repo := &fakeRepository{orders: []order.Order{{ID: 1, Status: tt.from, History: []order.StatusChange{initial}}}}
			s := service.NewOrderService(repo)
			value, err := s.UpdateStatus(context.Background(), 1, tt.to)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Equal(t, tt.from, repo.orders[0].Status)
				require.Len(t, repo.orders[0].History, 1)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.to, value.Status)
			require.Len(t, value.History, 2)
			require.Equal(t, initial, value.History[0])
			require.Equal(t, tt.to, value.History[1].Status)
			require.False(t, value.History[1].ChangedAt.Before(initial.ChangedAt))
			require.Equal(t, value, repo.orders[0])
		})
	}
	_, err := service.NewOrderService(&fakeRepository{}).UpdateStatus(context.Background(), 99, order.StatusProcessing)
	require.ErrorIs(t, err, order.ErrOrderNotFound)
}

func TestHistory(t *testing.T) {
	first := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	history := []order.StatusChange{
		{Status: order.StatusNew, ChangedAt: first},
		{Status: order.StatusProcessing, ChangedAt: first.Add(time.Minute)},
		{Status: order.StatusCompleted, ChangedAt: first.Add(2 * time.Minute)},
	}
	repo := &fakeRepository{orders: []order.Order{{ID: 1, History: []order.StatusChange{history[2], history[0], history[1]}}}}
	s := service.NewOrderService(repo)
	got, err := s.History(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, history, got)
	_, err = s.History(context.Background(), 99)
	require.ErrorIs(t, err, order.ErrOrderNotFound)
}

func TestInvalidOrderID(t *testing.T) {
	for _, tt := range []struct {
		name string
		id   int64
	}{
		{"zero", 0},
		{"negative", -1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// A nil repository makes an accidental storage call fail the test.
			s := service.NewOrderService(nil)
			ctx := context.Background()

			_, err := s.GetByID(ctx, tt.id)
			require.ErrorIs(t, err, order.ErrInvalidOrder)

			_, err = s.UpdateStatus(ctx, tt.id, order.StatusProcessing)
			require.ErrorIs(t, err, order.ErrInvalidOrder)

			_, err = s.UpdateStatus(ctx, tt.id, "unknown")
			require.ErrorIs(t, err, order.ErrInvalidOrder)

			_, err = s.History(ctx, tt.id)
			require.ErrorIs(t, err, order.ErrInvalidOrder)
		})
	}
}
