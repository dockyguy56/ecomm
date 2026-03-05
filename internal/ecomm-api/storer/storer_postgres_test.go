package storer

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func withTestDB(t *testing.T, fn func(st *PostgresStorer, mock sqlmock.Sqlmock)) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	st := NewPostgresStorer(db)
	fn(st, mock)
}

func TestCreateProduct(t *testing.T) {
	p := &Product{
		Name:         "Test Product",
		Image:        "test.jpg",
		Category:     "Test Category",
		Description:  "This is a test product",
		Rating:       5,
		NumReviews:   10,
		Price:        19.99,
		CountInStock: 100,
	}

	tcs := []struct {
		name string
		test func(*PostgresStorer, sqlmock.Sqlmock)
	}{
		{
			name: "success",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES (?, ?, ?, ?, ?, ?, ?, ?)").
					WithArgs(p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock).
					WillReturnResult(sqlmock.NewResult(1, 1))

				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(1)

				mock.ExpectQuery("SELECT id FROM products WHERE name=$1 AND image=$2 AND category=$3 AND description=$4 AND rating=$5 AND num_reviews=$6 AND price=$7 AND count_in_stock=$8").
					WillReturnRows(rows)

				cp, err := st.CreateProduct(context.Background(), p)
				require.NoError(t, err)
				require.Equal(t, int64(1), cp.ID)
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "insert error",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES (?, ?, ?, ?, ?, ?, ?, ?)").
					WithArgs(p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock).
					WillReturnError(fmt.Errorf("insert error"))

				_, err := st.CreateProduct(context.Background(), p)
				require.Error(t, err)
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "retrieve id error",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES (?, ?, ?, ?, ?, ?, ?, ?)").
					WithArgs(p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectQuery("SELECT id FROM products WHERE name=$1 AND image=$2 AND category=$3 AND description=$4 AND rating=$5 AND num_reviews=$6 AND price=$7 AND count_in_stock=$8").
					WillReturnError(fmt.Errorf("failed to retrieve created product"))

				_, err := st.CreateProduct(context.Background(), p)
				require.Error(t, err)
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withTestDB(t, tc.test)
		})
	}
}

func TestGetProductByID(t *testing.T) {
	p := &Product{
		Name:         "Test Product",
		Image:        "test.jpg",
		Category:     "Test Category",
		Description:  "This is a test product",
		Rating:       5,
		NumReviews:   10,
		Price:        19.99,
		CountInStock: 100,
	}

	tcs := []struct {
		name string
		test func(*PostgresStorer, sqlmock.Sqlmock)
	}{
		{
			name: "success",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "image", "category", "description", "rating", "num_reviews", "price", "count_in_stock", "created_at", "updated_at"}).
					AddRow(1, p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock, p.CreatedAt, p.UpdatedAt)

				mock.ExpectQuery("SELECT * FROM products WHERE id=$1").WithArgs(1).WillReturnRows(rows)

				gp, err := st.GetProductByID(context.Background(), 1)
				require.NoError(t, err)
				require.Equal(t, int64(1), gp.ID)
				require.Equal(t, p.Name, gp.Name)
				require.Equal(t, p.Image, gp.Image)
				require.Equal(t, p.Category, gp.Category)
				require.Equal(t, p.Description, gp.Description)
				require.Equal(t, p.Rating, gp.Rating)
				require.Equal(t, p.NumReviews, gp.NumReviews)
				require.Equal(t, p.Price, gp.Price)
				require.Equal(t, p.CountInStock, gp.CountInStock)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed getting product",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM products WHERE id=$1").
					WithArgs(1).
					WillReturnError(fmt.Errorf("failed to get product"))

				_, err := st.GetProductByID(context.Background(), 1)
				require.Error(t, err)
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withTestDB(t, tc.test)
		})
	}
}

func TestGetAllProducts(t *testing.T) {
	p := &Product{
		Name:         "Test Product",
		Image:        "test.jpg",
		Category:     "Test Category",
		Description:  "This is a test product",
		Rating:       5,
		NumReviews:   10,
		Price:        19.99,
		CountInStock: 100,
	}

	tcs := []struct {
		name string
		test func(*PostgresStorer, sqlmock.Sqlmock)
	}{
		{
			name: "success",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "image", "category", "description", "rating", "num_reviews", "price", "count_in_stock", "created_at", "updated_at"}).
					AddRow(1, p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock, p.CreatedAt, p.UpdatedAt)

				mock.ExpectQuery("SELECT * FROM products").WillReturnRows(rows)

				gps, err := st.GetAllProducts(context.Background())
				require.NoError(t, err)
				require.Len(t, gps, 1)
				require.Equal(t, int64(1), gps[0].ID)
				require.Equal(t, p.Name, gps[0].Name)
				require.Equal(t, p.Image, gps[0].Image)
				require.Equal(t, p.Category, gps[0].Category)
				require.Equal(t, p.Description, gps[0].Description)
				require.Equal(t, p.Rating, gps[0].Rating)
				require.Equal(t, p.NumReviews, gps[0].NumReviews)
				require.Equal(t, p.Price, gps[0].Price)
				require.Equal(t, p.CountInStock, gps[0].CountInStock)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed getting products",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM products").WillReturnError(fmt.Errorf("failed to get products"))
				_, err := st.GetAllProducts(context.Background())
				require.Error(t, err)
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withTestDB(t, tc.test)
		})
	}
}

func TestUpdateProduct(t *testing.T) {
	p := &Product{
		Name:         "Test Product",
		Image:        "test.jpg",
		Category:     "Test Category",
		Description:  "This is a test product",
		Rating:       5,
		NumReviews:   10,
		Price:        19.99,
		CountInStock: 100,
	}

	up := &Product{
		Name:         "Test Updated Product",
		Image:        "test Updated.jpg",
		Category:     "Test Updated Category",
		Description:  "This is a test Updated product",
		Rating:       5,
		NumReviews:   10,
		Price:        19.99,
		CountInStock: 100,
	}

	tcs := []struct {
		name string
		test func(*PostgresStorer, sqlmock.Sqlmock)
	}{
		{
			name: "success",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES (?, ?, ?, ?, ?, ?, ?, ?)").
					WithArgs(p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock).
					WillReturnResult(sqlmock.NewResult(1, 1))

				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(1)

				mock.ExpectQuery("SELECT id FROM products WHERE name=$1 AND image=$2 AND category=$3 AND description=$4 AND rating=$5 AND num_reviews=$6 AND price=$7 AND count_in_stock=$8").
					WillReturnRows(rows)

				cp, err := st.CreateProduct(context.Background(), p)
				require.NoError(t, err)
				require.Equal(t, int64(1), cp.ID)

				mock.ExpectExec("UPDATE products SET name=?, image=?, category=?, description=?, rating=?, num_reviews=?, price=?, count_in_stock=?, updated_at=? WHERE id=?").
					WillReturnResult(sqlmock.NewResult(1, 1))

				np, err := st.UpdateProduct(context.Background(), up)
				require.NoError(t, err)
				require.Equal(t, up.Name, np.Name)
				require.Equal(t, up.Image, np.Image)
				require.Equal(t, up.Category, np.Category)
				require.Equal(t, up.Description, np.Description)
				require.Equal(t, up.Rating, np.Rating)
				require.Equal(t, up.NumReviews, np.NumReviews)
				require.Equal(t, up.Price, np.Price)
				require.Equal(t, up.CountInStock, np.CountInStock)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed updating product",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE products SET name=?, image=?, category=?, description=?, rating=?, num_reviews=?, price=?, count_in_stock=?, updated_at=? WHERE id=?").
					WillReturnError(fmt.Errorf("error updating product"))
				_, err := st.UpdateProduct(context.Background(), p)
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withTestDB(t, tc.test)
		})
	}
}

func TestDeleteProduct(t *testing.T) {
	tcs := []struct {
		name string
		test func(*PostgresStorer, sqlmock.Sqlmock)
	}{
		{
			name: "success",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM products WHERE id=$1").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
				err := st.DeleteProduct(context.Background(), 1)
				require.NoError(t, err)
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed deleting product",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM products WHERE id=$1").WithArgs(1).WillReturnError(fmt.Errorf("failed to delete product"))
				err := st.DeleteProduct(context.Background(), 1)
				require.Error(t, err)
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withTestDB(t, tc.test)
		})
	}
}

func TestCreateOrder(t *testing.T) {
	ois := []OrderItem{
		{
			Name:      "test product",
			Quantity:  1,
			Image:     "test.jpg",
			Price:     99.99,
			ProductID: 1,
		},
		{
			Name:      "test product 2",
			Quantity:  2,
			Image:     "test2.jpg",
			Price:     199.99,
			ProductID: 2,
		},
	}

	o := &Order{
		PaymentMethod: "test payment method",
		TaxPrice:      10.0,
		ShippingPrice: 20.0,
		TotalPrice:    129.99,
		Items:         ois,
	}

	tcs := []struct {
		name string
		test func(*PostgresStorer, sqlmock.Sqlmock)
	}{
		{
			name: "success",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO orders (payment_method, tax_price, shipping_price, total_price) VALUES (?, ?, ?, ?)").
					WillReturnResult(sqlmock.NewResult(1, 1))
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(1)
				mock.ExpectQuery("SELECT id FROM orders WHERE payment_method=$1 AND tax_price=$2 AND shipping_price=$3 AND total_price=$4").
					WillReturnRows(rows)
				mock.ExpectExec("INSERT INTO order_items (name, quantity, image, price, product_id, order_id) VALUES (?, ?, ?, ?, ?, ?)").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectQuery("SELECT id FROM order_items WHERE order_id=$1 AND product_id=$2 AND name=$3 AND quantity=$4 AND image=$5 AND price=$6").
					WillReturnRows(rows)
				mock.ExpectExec("INSERT INTO order_items (name, quantity, image, price, product_id, order_id) VALUES (?, ?, ?, ?, ?, ?)").
					WillReturnResult(sqlmock.NewResult(2, 1))
				mock.ExpectCommit()

				co, err := st.CreateOrder(context.Background(), o)
				require.NoError(t, err)
				require.Equal(t, int64(1), co.ID)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed creating order",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO orders (payment_method, tax_price, shipping_price, total_price) VALUES (?, ?, ?, ?)").
					WillReturnError(fmt.Errorf("failed to create order"))
				mock.ExpectRollback()

				_, err := st.CreateOrder(context.Background(), o)
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to retrieve last created order id",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO orders (payment_method, tax_price, shipping_price, total_price) VALUES (?, ?, ?, ?)").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectQuery("SELECT id FROM orders WHERE payment_method=$1 AND tax_price=$2 AND shipping_price=$3 AND total_price=$4").
					WillReturnError(fmt.Errorf("failed to retrieve created order"))
				mock.ExpectRollback()

				_, err := st.CreateOrder(context.Background(), o)
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed creating order item",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO orders (payment_method, tax_price, shipping_price, total_price) VALUES (?, ?, ?, ?)").
					WillReturnResult(sqlmock.NewResult(1, 1))
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(1)
				mock.ExpectQuery("SELECT id FROM orders WHERE payment_method=$1 AND tax_price=$2 AND shipping_price=$3 AND total_price=$4").
					WillReturnRows(rows)
				mock.ExpectExec("INSERT INTO order_items (name, quantity, image, price, product_id, order_id) VALUES (?, ?, ?, ?, ?, ?)").
					WillReturnError(fmt.Errorf("failed to create order item"))
				mock.ExpectRollback()

				_, err := st.CreateOrder(context.Background(), o)
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "last created order item ID error",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO orders (payment_method, tax_price, shipping_price, total_price) VALUES (?, ?, ?, ?)").
					WillReturnResult(sqlmock.NewResult(1, 1))
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(1)
				mock.ExpectQuery("SELECT id FROM orders WHERE payment_method=$1 AND tax_price=$2 AND shipping_price=$3 AND total_price=$4").
					WillReturnRows(rows)
				mock.ExpectExec("INSERT INTO order_items (name, quantity, image, price, product_id, order_id) VALUES (?, ?, ?, ?, ?, ?)").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectQuery("SELECT id FROM order_items WHERE order_id=$1 AND product_id=$2 AND name=$3 AND quantity=$4 AND image=$5 AND price=$6").
					WillReturnError(fmt.Errorf("failed to retrieve created order items"))
				mock.ExpectRollback()

				_, err := st.CreateOrder(context.Background(), o)
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to rollback transaction",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO orders (payment_method, tax_price, shipping_price, total_price) VALUES (?, ?, ?, ?)").
					WillReturnError(fmt.Errorf("failed to create order"))
				mock.ExpectRollback().WillReturnError(fmt.Errorf("failed to rollback transaction"))

				_, err := st.CreateOrder(context.Background(), o)
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to begin transaction",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(fmt.Errorf("failed to begin transaction"))

				_, err := st.CreateOrder(context.Background(), o)
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to commit transaction",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO orders (payment_method, tax_price, shipping_price, total_price) VALUES (?, ?, ?, ?)").
					WillReturnResult(sqlmock.NewResult(1, 1))
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(1)
				mock.ExpectQuery("SELECT id FROM orders WHERE payment_method=$1 AND tax_price=$2 AND shipping_price=$3 AND total_price=$4").
					WillReturnRows(rows)
				mock.ExpectExec("INSERT INTO order_items (name, quantity, image, price, product_id, order_id) VALUES (?, ?, ?, ?, ?, ?)").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectQuery("SELECT id FROM order_items WHERE order_id=$1 AND product_id=$2 AND name=$3 AND quantity=$4 AND image=$5 AND price=$6").
					WillReturnRows(rows)
				mock.ExpectExec("INSERT INTO order_items (name, quantity, image, price, product_id, order_id) VALUES (?, ?, ?, ?, ?, ?)").
					WillReturnResult(sqlmock.NewResult(2, 1))
				mock.ExpectCommit().WillReturnError(fmt.Errorf("failed to commit transaction"))

				_, err := st.CreateOrder(context.Background(), o)
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withTestDB(t, tc.test)
		})
	}
}

func TestGetOrderByID(t *testing.T) {
	ois := []OrderItem{
		{
			Name:      "test product",
			Quantity:  1,
			Image:     "test.jpg",
			Price:     99.99,
			ProductID: 1,
		},
		{
			Name:      "test product 2",
			Quantity:  2,
			Image:     "test2.jpg",
			Price:     199.99,
			ProductID: 2,
		},
	}

	o := &Order{
		PaymentMethod: "test payment method",
		TaxPrice:      10.0,
		ShippingPrice: 20.0,
		TotalPrice:    129.99,
		Items:         ois,
	}

	tcs := []struct {
		name string
		test func(*PostgresStorer, sqlmock.Sqlmock)
	}{
		{
			name: "success",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				orderRows := sqlmock.NewRows([]string{"id", "payment_method", "tax_price", "shipping_price", "total_price", "created_at", "updated_at"}).
					AddRow(1, o.PaymentMethod, o.TaxPrice, o.ShippingPrice, o.TotalPrice, o.CreatedAt, o.UpdatedAt)
				mock.ExpectQuery("SELECT * FROM orders WHERE id=$1").
					WithArgs(1).
					WillReturnRows(orderRows)

				orderItemRows := sqlmock.NewRows([]string{"id", "name", "quantity", "image", "price", "product_id", "order_id"}).
					AddRow(1, ois[0].Name, ois[0].Quantity, ois[0].Image, ois[0].Price, ois[0].ProductID, 1).
					AddRow(2, ois[1].Name, ois[1].Quantity, ois[1].Image, ois[1].Price, ois[1].ProductID, 1)
				mock.ExpectQuery("SELECT * FROM order_items WHERE order_id=$1").
					WithArgs(1).
					WillReturnRows(orderItemRows)

				mo, err := st.GetOrderByID(context.Background(), 1)
				require.NoError(t, err)
				require.Equal(t, int64(1), mo.ID)

				for i, oi := range ois {
					require.Equal(t, oi.Name, mo.Items[i].Name)
					require.Equal(t, oi.Quantity, mo.Items[i].Quantity)
					require.Equal(t, oi.Image, mo.Items[i].Image)
					require.Equal(t, oi.Price, mo.Items[i].Price)
					require.Equal(t, oi.ProductID, mo.Items[i].ProductID)
				}

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to retrieve order",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM orders WHERE id=$1").
					WillReturnError(fmt.Errorf("failed to retrieve order"))

				_, err := st.GetOrderByID(context.Background(), 1)
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to retrieve order items",
			test: func(st *PostgresStorer, mock sqlmock.Sqlmock) {
				orderRows := sqlmock.NewRows([]string{"id", "payment_method", "tax_price", "shipping_price", "total_price", "created_at", "updated_at"}).
					AddRow(1, o.PaymentMethod, o.TaxPrice, o.ShippingPrice, o.TotalPrice, o.CreatedAt, o.UpdatedAt)
				mock.ExpectQuery("SELECT * FROM orders WHERE id=$1").
					WillReturnRows(orderRows)

				mock.ExpectQuery("SELECT * FROM order_items WHERE order_id=$1").
					WillReturnError(fmt.Errorf("failed to retrieve order items"))

				_, err := st.GetOrderByID(context.Background(), 1)
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withTestDB(t, tc.test)
		})
	}
}

func TestGetAllOrders(t *testing.T) {
	ois := []OrderItem{
		{
			Name:      "test product",
			Quantity:  1,
			Image:     "test.jpg",
			Price:     99.99,
			ProductID: 1,
		},
		{
			Name:      "test product 2",
			Quantity:  2,
			Image:     "test2.jpg",
			Price:     199.99,
			ProductID: 2,
		},
	}

	o := &Order{
		PaymentMethod: "test payment method",
		TaxPrice:      10.0,
		ShippingPrice: 20.0,
		TotalPrice:    129.99,
		Items:         ois,
	}

	tcs := []struct {
		name string
		test func(*PostgresStorer, sqlmock.Sqlmock)
	}{
		{
			name: "success",
			test: func(ps *PostgresStorer, mock sqlmock.Sqlmock) {
				orderRows := sqlmock.NewRows([]string{"id", "payment_method", "tax_price", "shipping_price", "total_price", "created_at", "updated_at"}).
					AddRow(1, o.PaymentMethod, o.TaxPrice, o.ShippingPrice, o.TotalPrice, o.CreatedAt, o.UpdatedAt)
				mock.ExpectQuery("SELECT * FROM orders").
					WillReturnRows(orderRows)

				orderItemRows := sqlmock.NewRows([]string{"id", "name", "quantity", "image", "price", "product_id", "order_id"}).
					AddRow(1, ois[0].Name, ois[0].Quantity, ois[0].Image, ois[0].Price, ois[0].ProductID, 1).
					AddRow(2, ois[1].Name, ois[1].Quantity, ois[1].Image, ois[1].Price, ois[1].ProductID, 1)
				mock.ExpectQuery("SELECT * FROM order_items WHERE order_id=$1").
					WillReturnRows(orderItemRows)

				mo, err := ps.GetAllOrders(context.Background())
				require.NoError(t, err)
				require.Len(t, mo, 1)
				require.Equal(t, int64(1), mo[0].ID)

				for i, oi := range ois {
					require.Equal(t, oi.Name, mo[0].Items[i].Name)
					require.Equal(t, oi.Quantity, mo[0].Items[i].Quantity)
					require.Equal(t, oi.Image, mo[0].Items[i].Image)
					require.Equal(t, oi.Price, mo[0].Items[i].Price)
					require.Equal(t, oi.ProductID, mo[0].Items[i].ProductID)
				}

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to retrieve orders",
			test: func(ps *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM orders").
					WillReturnError(fmt.Errorf("failed to retrieve orders"))

				_, err := ps.GetAllOrders(context.Background())
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to retrieve order items",
			test: func(ps *PostgresStorer, mock sqlmock.Sqlmock) {
				orderRows := sqlmock.NewRows([]string{"id", "payment_method", "tax_price", "shipping_price", "total_price", "created_at", "updated_at"}).
					AddRow(1, o.PaymentMethod, o.TaxPrice, o.ShippingPrice, o.TotalPrice, o.CreatedAt, o.UpdatedAt)
				mock.ExpectQuery("SELECT * FROM orders").
					WillReturnRows(orderRows)
				mock.ExpectQuery("SELECT * FROM order_items WHERE order_id=$1").
					WillReturnError(fmt.Errorf("failed to retrieve order items"))

				_, err := ps.GetAllOrders(context.Background())
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withTestDB(t, tc.test)
		})
	}
}

func TestDeleteOrder(t *testing.T) {
	tcs := []struct {
		name string
		test func(*PostgresStorer, sqlmock.Sqlmock)
	}{
		{
			name: "success",
			test: func(ps *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM order_items WHERE order_id=$1").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("DELETE FROM orders WHERE id=$1").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()

				err := ps.DeleteOrder(context.Background(), 1)
				require.NoError(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to delete order items",
			test: func(ps *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM order_items WHERE order_id=$1").
					WillReturnError(fmt.Errorf("failed to delete order items"))
				mock.ExpectRollback()

				err := ps.DeleteOrder(context.Background(), 1)
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to delete order",
			test: func(ps *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM order_items WHERE order_id=$1").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("DELETE FROM orders WHERE id=$1").
					WillReturnError(fmt.Errorf("failed to delete order"))
				mock.ExpectRollback()

				err := ps.DeleteOrder(context.Background(), 1)
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to begin transaction",
			test: func(ps *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(fmt.Errorf("failed to begin transaction"))

				err := ps.DeleteOrder(context.Background(), 1)
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to rollback transaction",
			test: func(ps *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM order_items WHERE order_id=$1").
					WillReturnError(fmt.Errorf("failed to delete order items"))
				mock.ExpectRollback().WillReturnError(fmt.Errorf("failed to rollback transaction"))

				err := ps.DeleteOrder(context.Background(), 1)
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to commit transaction",
			test: func(ps *PostgresStorer, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM order_items WHERE order_id=$1").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("DELETE FROM orders WHERE id=$1").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(fmt.Errorf("failed to commit transaction"))

				err := ps.DeleteOrder(context.Background(), 1)
				require.Error(t, err)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withTestDB(t, tc.test)
		})
	}
}
