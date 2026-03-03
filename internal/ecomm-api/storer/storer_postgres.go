package storer

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type PostgresStorer struct {
	db *sqlx.DB
}

func NewPostgresStorer(db *sqlx.DB) *PostgresStorer {
	return &PostgresStorer{db: db}
}

func (ps *PostgresStorer) CreateProduct(ctx context.Context, product *Product) (*Product, error) {
	// Implement the logic to insert a new product into the database
	// and return the created product with its ID.

	res, err := ps.db.NamedExec("INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES (:name, :image, :category, :description, :rating, :num_reviews, :price, :count_in_stock)", product)
	if err != nil {
		return nil, fmt.Errorf("error inserting product: %w", err)
	}

	// Get the ID of the newly created product
	productID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	product.ID = productID
	return product, nil
}

func (ps *PostgresStorer) GetProductByID(ctx context.Context, id int64) (*Product, error) {
	// Implement the logic to retrieve a product by its ID from the database.
	var product Product

	err := ps.db.GetContext(ctx, &product, "SELECT * FROM products WHERE id=$1", id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve product by ID: %w", err)
	}

	return &product, nil
}

func (ps *PostgresStorer) GetAllProducts(ctx context.Context) ([]*Product, error) {
	// Implement the logic to retrieve all products from the database.
	var products []*Product

	err := ps.db.SelectContext(ctx, &products, "SELECT * FROM products")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all products: %w", err)
	}

	return products, nil
}

func (ps *PostgresStorer) UpdateProduct(ctx context.Context, product *Product) (*Product, error) {
	// Implement the logic to update an existing product in the database
	// and return the updated product.
	_, err := ps.db.NamedExec("UPDATE products SET name=:name, image=:image, category=:category, description=:description, rating=:rating, num_reviews=:num_reviews, price=:price, count_in_stock=:count_in_stock WHERE id=:id", product)
	if err != nil {
		return nil, fmt.Errorf("error updating product: %w", err)
	}

	return product, nil
}

func (ps *PostgresStorer) DeleteProduct(ctx context.Context, id int64) error {
	// Implement the logic to delete a product by its ID from the database.
	_, err := ps.db.ExecContext(ctx, "DELETE FROM products WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}

	return nil
}

func (ps *PostgresStorer) CreateOrder(ctx context.Context, order *Order) (*Order, error) {
	// start a transaction
	err := ps.execTx(ctx, func(tx *sqlx.Tx) error {
		// Insert orders
		o, err := createOrder(ctx, tx, order)
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		for _, oi := range order.Items {
			oi.OrderID = o.ID
			err := createOrderItems(ctx, tx, *oi)
			if err != nil {
				return fmt.Errorf("failed to create order item: %w", err)
			}
		}

		return nil
		// Insert order item
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}

func createOrder(ctx context.Context, tx *sqlx.Tx, order *Order) (*Order, error) {
	res, err := tx.NamedExecContext(ctx, "INSERT INTO orders (payment_method, tax_price, shipping_price, total_price) VALUES (:payment_method, :tax_price, :shipping_price, :total_price)", order)
	if err != nil {
		return nil, fmt.Errorf("error inserting order: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	order.ID = id
	return order, nil
}

func createOrderItems(ctx context.Context, tx *sqlx.Tx, orderItem OrderItem) error {
	res, err := tx.NamedExecContext(ctx, "INSERT INTO order_items (name, quantity, image, price, product_id, order_id) VALUES (:name, :quantity, :image, :price, :product_id, :order_id)", orderItem)
	if err != nil {
		return fmt.Errorf("error inserting order item: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}
	orderItem.ID = id

	return nil
}

func (ps *PostgresStorer) GetOrderByID(ctx context.Context, id int64) (*Order, error) {
	// Implement the logic to retrieve an order by its ID from the database.
	var order Order
	err := ps.db.GetContext(ctx, &order, "SELECT * FROM orders WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve order by ID: %w", err)
	}

	var orderItems []*OrderItem
	err = ps.db.SelectContext(ctx, &orderItems, "SELECT * FROM order_items WHERE order_id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve order items: %w", err)
	}

	order.Items = orderItems
	return &order, nil
}

func (ps *PostgresStorer) GetAllOrders(ctx context.Context) ([]*Order, error) {
	// Implement the logic to retrieve all orders from the database.
	var orders []*Order
	err := ps.db.SelectContext(ctx, &orders, "SELECT * FROM orders")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all orders: %w", err)
	}

	for _, order := range orders {
		var orderItems []*OrderItem
		err = ps.db.SelectContext(ctx, &orderItems, "SELECT * FROM order_items WHERE order_id = ?", order.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve order items for order %d: %w", order.ID, err)
		}
		order.Items = orderItems
	}
	return orders, nil
}

// func (ps *PostgresStorer) UpdateOrder(ctx context.Context, order *Order) (*Order, error) {
// }

func (ps *PostgresStorer) DeleteOrder(ctx context.Context, id int64) error {
	err := ps.execTx(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, "DELETE FROM order_items WHERE order_id = ?", id)
		if err != nil {
			return fmt.Errorf("error deleting order items: %w", err)
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM orders WHERE id = ?", id)
		if err != nil {
			return fmt.Errorf("error deleting order: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	return nil
}

func (ps *PostgresStorer) execTx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := ps.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction rollback failed: %v, original error: %w", rbErr, err)
		}
		return fmt.Errorf("transaction failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}
