package storer

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // needed for postgres driver "go get github.com/lib/pq@latest"
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

	_, err := ps.db.NamedExecContext(ctx, "INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES (:name, :image, :category, :description, :rating, :num_reviews, :price, :count_in_stock)", product)
	if err != nil {
		return nil, fmt.Errorf("error inserting product: %w", err)
	}

	// Get the ID of the newly created product
	var id int

	err = ps.db.GetContext(ctx, &id, "SELECT id FROM products WHERE name=$1 AND image=$2 AND category=$3 AND description=$4 AND rating=$5 AND num_reviews=$6 AND price=$7 AND count_in_stock=$8", product.Name, product.Image, product.Category, product.Description, product.Rating, product.NumReviews, product.Price, product.CountInStock)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created product: %w", err)
	}

	product.ID = int64(id)

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

func (ps *PostgresStorer) GetAllProducts(ctx context.Context) ([]Product, error) {
	// Implement the logic to retrieve all products from the database.
	var products []Product

	err := ps.db.SelectContext(ctx, &products, "SELECT * FROM products")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all products: %w", err)
	}

	return products, nil
}

func (ps *PostgresStorer) UpdateProduct(ctx context.Context, product *Product) (*Product, error) {
	// Implement the logic to update an existing product in the database
	// and return the updated product.
	_, err := ps.db.NamedExec("UPDATE products SET name=:name, image=:image, category=:category, description=:description, rating=:rating, num_reviews=:num_reviews, price=:price, count_in_stock=:count_in_stock, updated_at=:updated_at WHERE id=:id", product)
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
			err := createOrderItems(ctx, tx, &oi)
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
	_, err := tx.NamedExecContext(ctx, "INSERT INTO orders (payment_method, tax_price, shipping_price, total_price, user_id) VALUES (:payment_method, :tax_price, :shipping_price, :total_price, :user_id)", order)
	if err != nil {
		return nil, fmt.Errorf("error inserting order: %w", err)
	}

	var id int

	err = tx.GetContext(ctx, &id, "SELECT id FROM orders WHERE payment_method=$1 AND tax_price=$2 AND shipping_price=$3 AND total_price=$4", order.PaymentMethod, order.TaxPrice, order.ShippingPrice, order.TotalPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created order: %w", err)
	}

	order.ID = int64(id)
	return order, nil
}

func createOrderItems(ctx context.Context, tx *sqlx.Tx, orderItem *OrderItem) error {
	_, err := tx.NamedExecContext(ctx, "INSERT INTO order_items (name, quantity, image, price, product_id, order_id) VALUES (:name, :quantity, :image, :price, :product_id, :order_id)", orderItem)
	if err != nil {
		return fmt.Errorf("error inserting order item: %w", err)
	}

	var id int
	err = tx.GetContext(ctx, &id, "SELECT id FROM order_items WHERE order_id=$1 AND product_id=$2 AND name=$3 AND quantity=$4 AND image=$5 AND price=$6", orderItem.OrderID, orderItem.ProductID, orderItem.Name, orderItem.Quantity, orderItem.Image, orderItem.Price)
	orderItem.ID = int64(id)

	return nil
}

func (ps *PostgresStorer) GetAllOrdersByID(ctx context.Context, userId int64) (*[]Order, error) {
	// Implement the logic to retrieve an order by its ID from the database.
	var orders []Order
	err := ps.db.SelectContext(ctx, &orders, "SELECT * FROM orders WHERE user_id=$1", userId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve order by ID: %w", err)
	}

	var orderItems []OrderItem
	for _, o := range orders {
		err = ps.db.SelectContext(ctx, &orderItems, "SELECT * FROM order_items WHERE order_id=$1", o.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve order items: %w", err)
		}

		o.Items = orderItems
	}

	return &orders, nil
}

func (ps *PostgresStorer) GetAllOrders(ctx context.Context) ([]*Order, error) {
	// Implement the logic to retrieve all orders from the database.
	var orders []*Order
	err := ps.db.SelectContext(ctx, &orders, "SELECT * FROM orders")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all orders: %w", err)
	}

	for _, order := range orders {
		var orderItems []OrderItem
		err = ps.db.SelectContext(ctx, &orderItems, "SELECT * FROM order_items WHERE order_id=$1", order.ID)
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
		_, err := tx.ExecContext(ctx, "DELETE FROM order_items WHERE order_id=$1", id)
		if err != nil {
			return fmt.Errorf("error deleting order items: %w", err)
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM orders WHERE id=$1", id)
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

func (ps *PostgresStorer) CreateUser(ctx context.Context, u *User) (*User, error) {
	_, err := ps.db.NamedExecContext(ctx, "INSERT INTO users (name, email, password, is_admin) VALUES (:name, :email, :password, :is_admin)", u)
	if err != nil {
		return nil, fmt.Errorf("error inserting user: %w", err)
	}

	var id int
	err = ps.db.GetContext(ctx, &id, "SELECT id FROM users WHERE name=$1 AND email=$2 AND password=$3 AND is_admin=$4", u.Name, u.Email, u.Password, u.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created user: %w", err)
	}

	u.ID = int64(id)

	return u, nil
}

func (ps *PostgresStorer) GetUser(ctx context.Context, email string) (*User, error) {
	var u User
	err := ps.db.GetContext(ctx, &u, "SELECT * FROM users WHERE email=$1", email)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &u, nil
}

func (ps *PostgresStorer) GetAllUsers(ctx context.Context) ([]User, error) {
	var users []User
	err := ps.db.SelectContext(ctx, &users, "SELECT * FROM users")
	if err != nil {
		return nil, fmt.Errorf("error Getting all users: %w", err)
	}

	return users, nil
}

func (ps *PostgresStorer) UpdateUser(ctx context.Context, u *User) (*User, error) {
	_, err := ps.db.NamedExecContext(ctx, "UPDATE users SET name=:name, email=:email, password=:password,is_admin=:is_admin, updated_at=:updated_at WHERE id=:id", u)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return u, nil
}

func (ps *PostgresStorer) DeleteUser(ctx context.Context, id int64) error {
	_, err := ps.db.ExecContext(ctx, "DELETE FROM users WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	return nil
}

func (ps *PostgresStorer) CreateSession(ctx context.Context, s *Session) (*Session, error) {
	_, err := ps.db.NamedExecContext(ctx, "INSERT INTO sessions (id, user_email, refresh_token, is_revoked, expires_at) VALUES (:id, :user_email, :refresh_token, :is_revoked, :expires_at)", s)
	if err != nil {
		return nil, fmt.Errorf("error inserting session: %w", err)
	}

	return s, nil
}

func (ps *PostgresStorer) GetSession(ctx context.Context, id string) (*Session, error) {
	var s Session
	err := ps.db.GetContext(ctx, &s, "SELECT * FROM sessions WHERE id=$1", id)
	if err != nil {
		return nil, fmt.Errorf("error getting session: %w", err)
	}

	return &s, nil
}

func (ps *PostgresStorer) RevokeSession(ctx context.Context, id string) error {
	_, err := ps.db.NamedExecContext(ctx, "UPDATE sessions SET is_revoked=true WHERE id=:id", map[string]interface{}{"id": id})
	if err != nil {
		return fmt.Errorf("error revoking user: %w", err)
	}

	return nil
}

func (ps *PostgresStorer) DeleteSession(ctx context.Context, id string) error {
	_, err := ps.db.ExecContext(ctx, "DELETE FROM sessions WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("error deleting session: %w", err)
	}

	return nil
}
