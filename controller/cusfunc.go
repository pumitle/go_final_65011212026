package controller

import (
	"go-final/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ฟังก์ชันนี้จะใช้ในการตั้งค่าตัวแปร DB
func SetDB(db *gorm.DB) {
	DB = db
}

func CustomerController(router *gin.Engine) {
	routers := router.Group("/get")
	{
		routers.GET("/user", getUsers)
		// routers.GET("/user/:id", getUserByID)
	}
	router.POST("/auth/login", loginCus)
	router.POST("/auth/register", registerCus)
	router.PUT("/upAdd", updateAddress)
	router.PUT("/changePass", changePassword)

}

func getUsers(c *gin.Context) {

	if DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	var users []model.Customer
	if err := DB.Find(&users).Error; err != nil { // ดึงข้อมูลทั้งหมด
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users}) // ส่งข้อมูลในรูปแบบ JSON
}

func loginCus(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	// รับค่า JSON และตรวจสอบ
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// ตรวจสอบอีเมลในฐานข้อมูล
	var user model.Customer
	if err := DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// ตรวจสอบ Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// ลบ Password ก่อนส่งกลับ
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user,
	})

}

func registerCus(c *gin.Context) {
	var input struct {
		FirstName   string `json:"first_name" binding:"required"`   // ชื่อจริง
		LastName    string `json:"last_name" binding:"required"`    // นามสกุล
		Email       string `json:"email" binding:"required"`        // อีเมล
		PhoneNumber string `json:"phone_number" binding:"required"` // เบอร์โทร
		Address     string `json:"address" binding:"required"`      // ที่อยู่
		Password    string `json:"password" binding:"required"`     // รหัสผ่าน
	}

	// รับค่า JSON และตรวจสอบ
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// ตรวจสอบว่าอีเมลมีอยู่แล้วหรือไม่
	var existingUser model.Customer
	if err := DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// สร้าง Hash ของ Password
	hashedPassword, err := HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// สร้าง User ใหม่
	user := model.Customer{
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		Email:       input.Email,
		PhoneNumber: input.PhoneNumber,
		Address:     input.Address,
		Password:    hashedPassword,
		CreatedAt:   time.Now(), // เวลาที่สร้าง
		UpdatedAt:   time.Now(), // เวลาที่อัพเดท
	}

	if err := DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ลบ Password ก่อนส่งกลับ
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "Registration successful",
		"user":    user,
	})
}

// แก้ไขที่อยู่
func updateAddress(c *gin.Context) {
	var input struct {
		CustomerID string `json:"customer_id" binding:"required"` // customer_id ที่ต้องการอัปเดต
		NewAddress string `json:"new_address" binding:"required"` // ที่อยู่ใหม่
	}

	// รับค่า JSON และตรวจสอบ
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// ค้นหาลูกค้าตาม customer_id
	var customer model.Customer
	if err := DB.Where("customer_id = ?", input.CustomerID).First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	// อัปเดตที่อยู่ของลูกค้า
	customer.Address = input.NewAddress
	customer.UpdatedAt = time.Now() // อัปเดตเวลา

	// บันทึกการเปลี่ยนแปลง
	if err := DB.Save(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update address"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Address updated successfully",
		"customer": customer,
	})
}

// /เปลี่ยนรหัสผ่าน
func changePassword(c *gin.Context) {
	var input struct {
		CustomerID  string `json:"customer_id" binding:"required"`  // customer_id ของผู้ใช้
		OldPassword string `json:"old_password" binding:"required"` // รหัสผ่านเก่า
		NewPassword string `json:"new_password" binding:"required"` // รหัสผ่านใหม่
	}

	// รับค่า JSON และตรวจสอบ
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// ค้นหาลูกค้าตาม customer_id
	var customer model.Customer
	if err := DB.Where("customer_id = ?", input.CustomerID).First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	// ตรวจสอบรหัสผ่านเก่า
	if err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(input.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid old password"})
		return
	}

	// สร้าง Hash ของรหัสผ่านใหม่
	hashedPassword, err := HashPassword(input.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	// อัปเดตรหัสผ่านใหม่
	customer.Password = hashedPassword
	customer.UpdatedAt = time.Now() // อัปเดตเวลา

	// บันทึกการเปลี่ยนแปลง
	if err := DB.Save(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// ลบ Password ก่อนส่งกลับ
	// customer.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"message":  "Password updated successfully",
		"customer": customer,
	})
}

// ฟังก์ชันสำหรับสร้าง Hash ของ Password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ฟังก์ชันสำหรับตรวจสอบ Password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
