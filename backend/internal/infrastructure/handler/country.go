package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/globaltask/bank/internal/infrastructure/repository"
)

type CountryHandler struct {
	countryRepo repository.CountryRepository
}

func NewCountryHandler(repo repository.CountryRepository) *CountryHandler {
	return &CountryHandler{countryRepo: repo}
}

func ListCountries(repo repository.CountryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		countries, err := repo.List(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data":  countries,
			"total": len(countries),
		})
	}
}

func GetCountry(repo repository.CountryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		country, err := repo.GetByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if country == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "country not found"})
			return
		}

		c.JSON(http.StatusOK, country)
	}
}
