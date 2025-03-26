package controller

import (
	"go-final/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ProductController(router *gin.Engine) {
	routers := router.Group("/get")
	{
		routers.GET("/pd", getProduct)
		routers.GET("/searcP", getSProduct)

	}
	router.POST("/cart", createOrUpdateCart)

}

func getProduct(c *gin.Context) {

	if DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	var pd []model.Product
	if err := DB.Find(&pd).Error; err != nil { // ดึงข้อมูลทั้งหมด
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pd}) // ส่งข้อมูลในรูปแบบ JSON
}

// เส้นค้นหาสินค้า
func getSProduct(c *gin.Context) {

	if DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	// รับพารามิเตอร์จาก query string
	description := c.DefaultQuery("description", "")      // คำค้นหาสำหรับรายละเอียดสินค้า
	minPriceStr := c.DefaultQuery("min_price", "0")       // ราคาต่ำสุด
	maxPriceStr := c.DefaultQuery("max_price", "1000000") // ราคาสูงสุด

	// แปลงราคาจาก string เป็น float64
	minPrice, err := strconv.ParseFloat(minPriceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min_price"})
		return
	}
	maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid max_price"})
		return
	}

	// ค้นหาสินค้าตามเงื่อนไข
	var products []model.Product
	err = DB.Where("description LIKE ? AND price BETWEEN ? AND ?", "%"+description+"%", minPrice, maxPrice).Find(&products).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": products}) // ส่งข้อมูลในรูปแบบ JSON
}

// สร้างรถเข็นใหม่ หรือค้นหารถเข็นที่มีอยู่แล้ว
func createOrUpdateCart(c *gin.Context) {
	var input struct {
		CustomerID string `json:"customer_id" binding:"required"` // customer_id ของเจ้าของรถเข็น
		CartName   string `json:"cart_name" binding:"required"`   // ชื่อรถเข็น
		ProductID  int    `json:"product_id" binding:"required"`  // ID ของสินค้า
		Quantity   int    `json:"quantity" binding:"required"`    // จำนวนสินค้าที่ต้องการเพิ่ม
	}

	// รับค่า JSON และตรวจสอบ
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// แปลง CustomerID จาก string เป็น int
	customerID, err := strconv.Atoi(input.CustomerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer_id"})
		return
	}
	// ค้นหารถเข็นที่มีชื่อและเป็นของลูกค้าคนนั้น
	var cart model.Cart
	if err := DB.Where("customer_id = ? AND cart_name = ?", input.CustomerID, input.CartName).First(&cart).Error; err != nil {
		// หากไม่พบรถเข็น, สร้างรถเข็นใหม่
		cart = model.Cart{
			CustomerID: customerID,
			CartName:   input.CartName,
		}
		if err := DB.Create(&cart).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cart"})
			return
		}
	}

	// ค้นหาสินค้าภายในรถเข็นนั้น
	var cartItem model.CartItem
	if err := DB.Where("cart_id = ? AND product_id = ?", cart.CartID, input.ProductID).First(&cartItem).Error; err == nil {
		// หากสินค้ามีอยู่แล้วในรถเข็น, เพิ่มจำนวนสินค้า
		cartItem.Quantity += input.Quantity
		if err := DB.Save(&cartItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Product quantity updated successfully"})
		return
	}

	// หากสินค้ายังไม่เคยมีในรถเข็น, เพิ่มสินค้าลงไป
	newCartItem := model.CartItem{
		CartID:    cart.CartID,
		ProductID: input.ProductID,
		Quantity:  input.Quantity,
	}
	if err := DB.Create(&newCartItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product to cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product added to cart successfully"})
}
