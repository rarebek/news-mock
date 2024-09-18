package v1

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/k0kubun/pp"

	"tarkib.uz/internal/controller/http/models"
	"tarkib.uz/internal/entity"
	"tarkib.uz/internal/usecase"
	"tarkib.uz/pkg/logger"
)

type categoryRoutes struct {
	t usecase.CategoryUseCase
	l logger.Interface
}

func newCategoryRoutes(handler *gin.RouterGroup, t usecase.CategoryUseCase, l logger.Interface) {
	r := &categoryRoutes{t, l}

	// Category routes
	handler.POST("/category", r.AppendCategoryWithSubCategoriesHandler)
	handler.PUT("/category/:id", r.UpdateCategoryHandler)
	handler.DELETE("/category/:id", r.DeleteCategoryHandler)
	handler.GET("/category/categories", r.GetAllCategoriesWithSubCategories)
	handler.GET("/categories", r.GetAllCategoriesHandler)
	handler.GET("/category/:id", r.GetOneCategoryByID)

	// Subcategory routes
	// handler.POST("/subcategory", r.AppendSubCategoryHandler)
	handler.PUT("/subcategory/append", r.AppendSubCategoryHandler)
	handler.PUT("/subcategory/:id", r.UpdateSubCategoryHandler)
	handler.DELETE("/subcategory/:id", r.DeleteSubCategoryHandler)
	handler.GET("/subcategories/:id", r.GetAllSubCategoriesHandler)

	h := handler.Group("/category")
	{
		h.POST("/source", r.CreateSource)
		h.DELETE("/source/delete/:id", r.DeleteSources)
		h.GET("/source", r.GetAllSources)
	}
}

// Get godoc
// @Summary get an existing category
// @Description get category by ID with Uzbek and Russian names
// @Tags Category
// @Accept  json
// @Produce  json
// @Param id path string true "Category ID"
// @Success 200 {object} response
// @Failure 400 {object} response
// @Failure 404 {object} response
// @Router /category/{id} [get]
func (n *categoryRoutes) GetOneCategoryByID(c *gin.Context) {
	id := c.Param("id")
	pp.Println("ID", id)

	res, err := n.t.GetOneCategoryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// AppendCategoryWithSubCategoriesHandler godoc
// @Summary Create a new category with subcategories
// @Description Create a new category with its associated subcategories in Uzbek and Russian
// @Tags Category
// @Accept  json
// @Produce  json
// @Param category body entity.CategoryWithSubCategories true "Category with subcategories data"
// @Success 200 {object} entity.CategoryWithSubCategories
// @Failure 400 {object} response
// @Router /category [post]
func (n *categoryRoutes) AppendCategoryWithSubCategoriesHandler(c *gin.Context) {
	var categoryWithSubCategories entity.CategoryWithSubCategories
	if err := c.ShouldBindJSON(&categoryWithSubCategories); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pp.Println(categoryWithSubCategories)

	id := uuid.NewString()

	// First, create the category
	err := n.t.AppendCategory(c, &entity.Category{
		ID:     id,
		NameUz: categoryWithSubCategories.NameUz,
		NameRu: categoryWithSubCategories.NameRu,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category: " + err.Error()})
		return
	}

	// Now, create the associated subcategories
	for _, subcategory := range categoryWithSubCategories.SubCategories {
		subcategory.CategoryID = id
		err := n.t.AppendSubCategory(c, &subcategory)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subcategory: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, models.Message{
		Message: "Category and subcategories created successfully!",
	})
}

// func (n *categoryRoutes) AppendCategoryHandler(c *gin.Context) {
// 	var category entity.Category
// 	if err := c.BindJSON(&category); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	err := n.t.AppendCategory(c, &category)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, models.Message{
// 		Message: "Category created successfully!",
// 	})
// }

// UpdateCategoryHandler godoc
// @Summary Update an existing category
// @Description Update category by ID with Uzbek and Russian names
// @Tags Category
// @Accept  json
// @Produce  json
// @Param category body entity.CategoryWithSubCategories true "Updated category data"
// @Success 200 {object} response
// @Failure 400 {object} response
// @Failure 404 {object} response
// @Router /category/{id} [put]
func (n *categoryRoutes) UpdateCategoryHandler(c *gin.Context) {
	var category entity.CategoryWithSubCategories
	if err := c.BindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := n.t.UpdateCategory(c, category.ID, category.NameUz, category.NameRu)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, subcategory := range category.SubCategories {
		err := n.t.UpdateSubCategory(c, subcategory.ID, subcategory.NameUz, subcategory.NameRu)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category updated"})
}

// DeleteCategoryHandler godoc
// @Summary Delete a category
// @Description Delete category by ID
// @Tags Category
// @Produce  json
// @Param id path string true "Category ID"
// @Success 200 {object} response
// @Failure 404 {object} response
// @Router /category/{id} [delete]
func (n *categoryRoutes) DeleteCategoryHandler(c *gin.Context) {
	id := c.Param("id")

	err := n.t.DeleteCategory(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}

// GetAllCategoriesHandler godoc
// @Summary Get all categories
// @Description Retrieve all categories
// @Tags Category
// @Produce  json
// @Success 200 {array} entity.Category
// @Failure 500 {object} response
// @Router /categories [get]
func (n *categoryRoutes) GetAllCategoriesHandler(c *gin.Context) {
	categories, err := n.t.GetAllCategories(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// Append Subcategory godoc
// @Summary Append to existing subcategories
// @Description Append multiple subcategories with Uzbek and Russian names
// @Tags Subcategory
// @Accept  json
// @Produce  json
// @Param subcategories body []entity.SubCategory true "Array of subcategories data"
// @Success 200 {object} response
// @Failure 400 {object} response
// @Failure 404 {object} response
// @Router /subcategory/append [put]
func (n *categoryRoutes) AppendSubCategoryHandler(c *gin.Context) {
	var subcategories []entity.SubCategory

	// Read the body as a raw byte slice
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	// Log the raw body for debugging
	pp.Println("Received body:", string(body))

	// Re-parse the body into the correct format
	err = json.Unmarshal(body, &subcategories)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format: " + err.Error()})
		return
	}

	// Process each subcategory
	for _, subcategory := range subcategories {
		err := n.t.AppendSubCategory(c, &subcategory)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to append subcategory: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subcategories created successfully!",
	})
}

// UpdateSubCategoryHandler godoc
// @Summary Update an existing subcategory
// @Description Update subcategory by ID with Uzbek and Russian names
// @Tags Subcategory
// @Accept  json
// @Produce  json
// @Param id path string true "Subcategory ID"
// @Param subcategory body entity.SubCategory true "Updated subcategory data"
// @Success 200 {object} response
// @Failure 400 {object} response
// @Failure 404 {object} response
// @Router /subcategory/{id} [put]
func (n *categoryRoutes) UpdateSubCategoryHandler(c *gin.Context) {
	id := c.Param("id")
	var subcategory entity.SubCategory
	if err := c.BindJSON(&subcategory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := n.t.UpdateSubCategory(c, id, subcategory.NameUz, subcategory.NameRu)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subcategory updated"})
}

// DeleteSubCategoryHandler godoc
// @Summary Delete a subcategory
// @Description Delete subcategory by ID
// @Tags Subcategory
// @Produce  json
// @Param id path string true "Subcategory ID"
// @Success 200 {object} response
// @Failure 404 {object} response
// @Router /subcategory/{id} [delete]
func (n *categoryRoutes) DeleteSubCategoryHandler(c *gin.Context) {
	id := c.Param("id")

	err := n.t.DeleteSubCategory(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subcategory deleted"})
}

// GetAllSubCategoriesHandler godoc
// @Summary Get all subcategories for a category
// @Description Retrieve all subcategories for a given category ID
// @Tags Subcategory
// @Produce  json
// @Param id path string true "Category ID"
// @Success 200 {array} entity.SubCategory
// @Failure 404 {object} response
// @Failure 500 {object} response
// @Router /subcategories/{id} [get]
func (n *categoryRoutes) GetAllSubCategoriesHandler(c *gin.Context) {
	categoryID := c.Param("id")

	subcategories, err := n.t.GetAllSubCategories(c, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subcategories)
}

// GetAllCategoriesWithSubCategories godoc
// @Summary     Get all categories with subcategories
// @Description This method retrieves all categories with their subcategories, based on language query parameter
// @ID          getall-categories-with-subcategories
// @Tags        category
// @Accept      json
// @Produce     json
// @Param       language  query  string  true  "Language of the category (e.g. 'uz' or 'ru')"
// @Success     200 {object} []entity.CategoryWithSubCategories
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Router      /category/categories [get]
func (n *categoryRoutes) GetAllCategoriesWithSubCategories(c *gin.Context) {
	language := c.Query("language")
	categories, err := n.t.GetAllCategoriesWithSubCategories(c.Request.Context(), language)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Failed to get categories with subcategories", false)
		return
	}

	c.JSON(http.StatusOK, categories)
}

// @Summary     Create Source
// @Description This method creates a new source
// @ID          create-source
// @Tags  	    source
// @Accept      json
// @Produce     json
// @Param       request body   entity.Source true  "Source body"
// @Success     200 {object} models.Message
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Security    BearerAuth
// @Router      /category/source [post]
func (n *categoryRoutes) CreateSource(c *gin.Context) {
	var response entity.Source

	err := c.ShouldBindJSON(&response)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Invalid request body", false)
		return
	}

	if err := n.t.CreateSource(c.Request.Context(), &response); err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Failed to create source", false)
	}

	c.JSON(200, models.Message{
		Message: "Source created successfully!",
	})
}

// @Summary     GetAllSource
// @Description This method gets all sources
// @ID          getall-source
// @Tags  	    source
// @Accept      json
// @Produce     json
// @Success     200 {object} []entity.Source
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Security    BearerAuth
// @Router      /category/source [get]
func (n *categoryRoutes) GetAllSources(c *gin.Context) {
	sources, err := n.t.GetAllSources(c.Request.Context())
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Failed to get sources", false)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "data": sources})
}

// @Summary     Delete Source
// @Description This method deletes given source by its id
// @ID          delete-source
// @Tags  	    source
// @Accept      json
// @Produce     json
// @Param       id  path string true  "Id of source"
// @Success     200 {object} []entity.Source
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Router      /category/source/delete/{id} [delete]
func (n *categoryRoutes) DeleteSources(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		n.l.Error("Missing id parameter")
		errorResponse(c, http.StatusBadRequest, "Missing id parameter", false)
		return
	}

	err := n.t.DeleteSource(c.Request.Context(), id)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Failed to delete source", false)
		return
	}
}
