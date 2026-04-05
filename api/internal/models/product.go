package models

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

var seedProducts = []Product{
	{ID: "1", Name: "Wireless Headphones", Description: "Noise-cancelling over-ear headphones", Price: 99.99},
	{ID: "2", Name: "Mechanical Keyboard", Description: "TKL layout with Cherry MX switches", Price: 149.99},
	{ID: "3", Name: "USB-C Hub", Description: "7-in-1 multiport adapter", Price: 39.99},
	{ID: "4", Name: "Monitor Stand", Description: "Adjustable ergonomic stand", Price: 59.99},
	{ID: "5", Name: "Webcam HD", Description: "1080p with built-in microphone", Price: 79.99},
	{ID: "6", Name: "Mouse Pad XL", Description: "Extended desk mat 90x40cm", Price: 24.99},
}

func GetAllProducts() []Product {
	return seedProducts
}
