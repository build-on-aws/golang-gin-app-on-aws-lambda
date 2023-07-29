package main

import (
	"errors"
	"function/db"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateURLRequest struct {
	URL string `json:"url"`
}

func CreateShortURL(c *gin.Context) {
	var req CreateURLRequest

	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	url := req.URL

	log.Println("shortcode creation request for", url)

	shortCode, err := db.SaveURL(url)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	response := Response{ShortCode: shortCode}
	c.JSON(201, response)

}

type Response struct {
	ShortCode string `json:"short_code"`
}

const pathParameterName = "shortcode"
const locationHeader = "Location"

func GetShortURL(c *gin.Context) {
	shortCode := c.Param(pathParameterName)

	log.Println("redirect request for shortcode", shortCode)

	longurl, err := db.GetLongURL(shortCode)

	if err != nil {
		if errors.Is(err, db.ErrUrlNotFound) {
			c.AbortWithError(http.StatusNotFound, err)
			return
		} else if errors.Is(err, db.ErrUrlNotActive) {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.Println("redirecting to ", longurl)

	c.Header(locationHeader, longurl)
	c.Status(http.StatusFound)

}

type Payload struct {
	Active bool `json:"active"`
}

func UpdateStatus(c *gin.Context) {
	var payload Payload
	c.BindJSON(&payload)

	shortCode := c.Param(pathParameterName)
	log.Println("status update request for short-code", shortCode, payload.Active)

	err := db.Update(shortCode, payload.Active)
	if err != nil {
		if errors.Is(err, db.ErrUrlNotFound) {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.Println("successfully updated status for", shortCode)
	c.Status(http.StatusNoContent)
}

func DeleteShortURL(c *gin.Context) {
	shortCode := c.Param(pathParameterName)

	log.Println("delete request for short-code", shortCode)

	err := db.Delete(shortCode)
	if err != nil {
		if errors.Is(err, db.ErrUrlNotFound) {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.Println("successfully deleted", shortCode)
	c.Status(http.StatusNoContent)
}
