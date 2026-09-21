package apperror

import (
	"errors"
	"net/http"
)

var (
	// Auth Errors
	ErrEmailAlreadyRegistered = errors.New("email sudah terdaftar")
	ErrInvalidCredentials     = errors.New("email atau password salah")
	ErrUserNotFound           = errors.New("user tidak ditemukan")
	ErrInvalidUserID          = errors.New("ID user tidak valid")
	ErrUnauthorized           = errors.New("user tidak terotentikasi")

	// Order Errors
	ErrEmptyCart           = errors.New("keranjang belanja kosong")
	ErrInvalidItemQuantity = errors.New("jumlah barang minimal 1")
	ErrInvalidVoucher      = errors.New("kode voucher tidak valid atau sudah kadaluwarsa")
	ErrOrderNotFound       = errors.New("pesanan tidak ditemukan")
	ErrInvalidOrderID      = errors.New("ID pesanan tidak valid")
	ErrInvalidOrderStatus  = errors.New("status pesanan tidak valid")
	ErrAccessDenied        = errors.New("akses ditolak: bukan pemilik pesanan ini")

	// Product Errors
	ErrProductNotFound  = errors.New("produk tidak ditemukan")
	ErrInvalidProductID = errors.New("ID produk tidak valid")
)

// HTTPStatus memetakan domain error ke status code HTTP yang sesuai
func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	switch {
	case errors.Is(err, ErrEmailAlreadyRegistered):
		return http.StatusConflict // 409
	case errors.Is(err, ErrInvalidCredentials), errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized // 401
	case errors.Is(err, ErrAccessDenied):
		return http.StatusForbidden // 403
	case errors.Is(err, ErrUserNotFound), errors.Is(err, ErrOrderNotFound), errors.Is(err, ErrProductNotFound):
		return http.StatusNotFound // 404
	case errors.Is(err, ErrEmptyCart), errors.Is(err, ErrInvalidItemQuantity),
		errors.Is(err, ErrInvalidVoucher), errors.Is(err, ErrInvalidOrderID),
		errors.Is(err, ErrInvalidOrderStatus), errors.Is(err, ErrInvalidProductID),
		errors.Is(err, ErrInvalidUserID):
		return http.StatusBadRequest // 400
	default:
		return http.StatusBadRequest
	}
}
