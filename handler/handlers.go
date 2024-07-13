package handler

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	m "shopping-basket-service/model"
	"time"
)

var jwtSecret = []byte("something")

// AuthMiddleware is a middleware to check JWT token for authentication
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}

		c.Set("user", token.Claims.(jwt.MapClaims)["user"])
		return next(c)
	}
}

func GetBasketsHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	userID := c.Get("user").(float64)
	var baskets []m.Basket
	db.Where("owner_id = ?", uint(userID)).Find(&baskets)
	return c.JSON(http.StatusOK, baskets)
}

func CreateBasketHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	ownerID := c.Get("user").(float64)

	var basket m.Basket
	if err := c.Bind(&basket); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "request body mismatch")
	}

	basket.State = "PENDING"
	basket.OwnerID = uint(ownerID)

	db.Create(&basket)
	return c.JSON(http.StatusCreated, basket)
}

func UpdateBasketHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	id := c.Param("id")
	userID := c.Get("user").(float64)

	var basket m.Basket
	if err := db.First(&basket, id).Error; err != nil || basket.OwnerID != uint(userID) {
		fmt.Printf("user: %v with %T, basket owner: %v with %T", userID, userID, basket.OwnerID, basket.OwnerID)
		return echo.NewHTTPError(http.StatusNotFound, "Basket not found")
	}

	if basket.State == "COMPLETED" {
		return echo.NewHTTPError(http.StatusForbidden, "Cannot update a completed basket")
	}

	err := c.Bind(&basket)
	if err != nil {
		return err
	}

	if basket.State != "PENDING" && basket.State != "COMPLETED" {
		return echo.NewHTTPError(http.StatusBadRequest, "State must be PENDING or COMPLETED")
	}

	basket.UpdatedAt = time.Now()
	basket.OwnerID = uint(userID)
	db.Save(&basket)

	return c.JSON(http.StatusOK, basket)
}

func GetBasketHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	id := c.Param("id")
	userID := c.Get("user").(float64)

	var basket m.Basket
	if err := db.First(&basket, id).Error; err != nil || basket.OwnerID != uint(userID) {
		return echo.NewHTTPError(http.StatusNotFound, "Basket not found")
	}

	return c.JSON(http.StatusOK, basket)
}

func DeleteBasketHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	id := c.Param("id")
	userID := c.Get("user").(float64)

	var basket m.Basket
	if err := db.First(&basket, id).Error; err != nil || basket.OwnerID != uint(userID) {
		return echo.NewHTTPError(http.StatusNotFound, "Basket not found")
	}

	db.Delete(&basket)
	return c.NoContent(http.StatusNoContent)
}

func SignupHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)

	var user m.User
	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	var temp m.User
	counter := db.Where("username = ?", user.Username).First(&temp).RowsAffected
	if counter != 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "There is already a user with this username")
	}

	db.Create(&user)
	return c.JSON(http.StatusCreated, user)
}

func LoginHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)

	var user m.User
	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	fmt.Println(user)

	var existingUser m.User
	if err := db.Where("username = ?", user.Username).First(&existingUser).Error; err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid username or password")
	}
	
	if user.Password != existingUser.Password {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid username or password")
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user"] = float64(existingUser.ID)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": tokenString,
	})
}
