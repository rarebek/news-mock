package v1

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/k0kubun/pp"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nfnt/resize"

	"tarkib.uz/internal/controller/http/models"
	"tarkib.uz/internal/entity"
	"tarkib.uz/internal/usecase"
	"tarkib.uz/pkg/logger"
	tokens "tarkib.uz/pkg/token"
)

var (
	failedAttempts = 0
	isBlocked      = false
)

type authRoutes struct {
	t usecase.Auth
	l logger.Interface
}

func newAuthRoutes(handler *gin.RouterGroup, t usecase.Auth, l logger.Interface) {
	r := &authRoutes{t, l}
	handler.POST("/file/upload", r.upload)

	h := handler.Group("/auth")
	{
		h.POST("/admin/login", r.login)
		h.POST("/superadmin/login", r.superAdminLogin)
		h.POST("/admin/create", r.createAdmin)
		h.DELETE("/admin/delete/:id", r.deleteAdmin)
		h.GET("/admin/getall", r.getAllAdmins)
		h.PUT("/admin/edit", r.editAdmin)
		h.GET("/admin/:id", r.getAdminData)
		h.PUT("/superadmin/edit", r.editSuperAdmin)
	}
}

// @Summary     Login
// @Description Authenticates an admin and returns an access token on successful login.
// @ID          admin-login
// @Tags  	    admin
// @Accept      json
// @Produce     json
// @Param       request body models.AdminLoginRequest true "Phone Number and Password"
// @Success     200 {object} entity.Admin
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Router      /auth/admin/login [post]
func (r *authRoutes) login(c *gin.Context) {
	var request models.AdminLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusBadRequest, models.InvalidRequestBody, false)
		return
	}

	admin, err := r.t.Login(c.Request.Context(), &entity.Admin{
		Username: request.Username,
		Password: request.Password,
	})

	if err != nil {
		switch err.Error() {
		case "no rows in result set":
			r.l.Warn(err.Error())
			errorResponse(c, http.StatusBadRequest, "Bunday admin topilmadi.", false)
		case "xato parol kiritdingiz":
			r.l.Warn(err.Error())
			errorResponse(c, http.StatusUnauthorized, "Username yoki parol xato kiritildi.", false)
		default:
			r.l.Error(err)
			errorResponse(c, http.StatusInternalServerError, models.ErrServerProblems, false)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"admin": admin, "status": true})
}

// @Summary     Super Admin Login
// @Description Authenticates a super admin and returns an access token on successful login.
// @ID          superadmin-login
// @Tags  	    superadmin
// @Accept      json
// @Produce     json
// @Param       request body models.SuperAdminLoginRequest true "Phone Number and Password"
// @Success     200 {object} entity.SuperAdmin
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Router      /auth/superadmin/login [post]
func (r *authRoutes) superAdminLogin(c *gin.Context) {
	var request models.SuperAdminLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusBadRequest, models.InvalidRequestBody, false)
		return
	}

	superAdmin, err := r.t.SuperAdminLogin(c.Request.Context(), &entity.SuperAdmin{
		PhoneNumber: request.PhoneNumber,
		Password:    request.Password,
	})
	pp.Println(superAdmin)

	if superAdmin.IsBlocked {
		r.l.Warn("Super admin is blocked.")
		errorResponse(c, http.StatusForbidden, "Super admin is blocked.", false)
		return
	}

	if err != nil {
		switch err.Error() {
		case "no rows in result set":
			r.l.Warn(err.Error())
			errorResponse(c, http.StatusBadRequest, "Bunday admin topilmadi.", false)
		case "xato parol kiritdingiz":
			r.l.Warn(err.Error())
			handleErr := r.HandleFailedAttempt(c.Request.Context(), request.PhoneNumber)
			if handleErr != nil {
				r.l.Error(handleErr)
				errorResponse(c, http.StatusInternalServerError, models.ErrServerProblems, false)
				return
			}
			errorResponse(c, http.StatusUnauthorized, "Telefon raqam yoki parol xato kiritildi.", false)
		default:
			r.l.Error(err)
			errorResponse(c, http.StatusInternalServerError, models.ErrServerProblems, false)
		}
		return
	}

	if isBlocked {
		r.l.Warn("Super admin is blocked.")
		errorResponse(c, http.StatusForbidden, "Super admin is blocked.", false)
		return
	}

	// Reset failed attempts on successful login
	resetErr := r.ResetFailedAttempts(c.Request.Context(), request.PhoneNumber)
	if resetErr != nil {
		r.l.Error(resetErr)
		errorResponse(c, http.StatusInternalServerError, models.ErrServerProblems, false)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"superadmin": superAdmin,
		"status":     true,
	})
}

// @Summary     Create Admin
// @Description Creates an admin
// @ID          superadmin-create-admin
// @Tags  	    superadmin
// @Accept      json
// @Produce     json
// @Param       request body models.AdminLoginRequest true "Phone Number and Password to create Admin"
// @Success     200 {object} models.AdminLoginResponse
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Security    BearerAuth
// @Router      /auth/admin/create [post]
func (r *authRoutes) createAdmin(c *gin.Context) {
	var request models.AdminLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusBadRequest, models.InvalidRequestBody, false)
		return
	}

	if err := r.t.CreateAdmin(c.Request.Context(), &entity.Admin{
		Username: request.Username,
		Password: request.Password,
	}); err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusBadRequest, models.ErrServerProblems, false)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Admin muvaffaqiyatli yaratildi",
		"status":  true,
	})
}

// @Summary     Delete Admin
// @Description This method deletes admin.
// @ID          superadmin-delete-admin
// @Tags  	    superadmin
// @Accept      json
// @Produce     json
// @Param       id path string true "ID of the admin to delete"
// @Success     200 {object} response
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Security    BearerAuth
// @Router      /auth/admin/delete/{id} [delete]
func (r *authRoutes) deleteAdmin(c *gin.Context) {
	id := c.Param("id")

	if err := r.t.DeleteAdmin(c.Request.Context(), id); err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusBadRequest, models.ErrServerProblems, false)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Admin muvaffaqiyatli o'chirildi",
		"status":  true,
	})
}

// @Summary		Get Admin Data
// @Description This method is getting admin by its id
// @ID          get-admin
// @Tags  	    admin
// @Accept      json
// @Produce     json
// @Param       id path string true "ID of the admin to get"
// @Success     200 {object} models.Admin
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Security    BearerAuth
// @Router      /auth/admin/{id} [get]
func (r *authRoutes) getAdminData(c *gin.Context) {
	id := c.Param("id")

	admin, err := r.t.GetAdminById(c.Request.Context(), id)
	if err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusBadRequest, models.ErrServerProblems, false)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"admin":  admin,
		"status": true,
	})
}

// @Summary     Get All Admins
// @Description Gets All Admins
// @ID          get-all-admins
// @Tags  	    superadmin
// @Accept      json
// @Produce     json
// @Success     200 {object} []entity.Admin
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Security    BearerAuth
// @Router      /auth/admin/getall [get]
func (r *authRoutes) getAllAdmins(c *gin.Context) {
	admins, err := r.t.GetAllAdmins(c.Request.Context())
	if err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusBadRequest, models.ErrServerProblems, false)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"admins": admins,
		"status": true,
	})
}

// @Summary     Edit Admin
// @Description ID of the admin to update and other fields will be updated.
// @ID          edit-admins
// @Tags  	    superadmin
// @Accept      json
// @Produce     json
// @Param       request body models.Admin true "ID of the admin to edit"
// @Success     200 {object} models.Message
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Security    BearerAuth
// @Router      /auth/admin/edit [put]
func (r *authRoutes) editAdmin(c *gin.Context) {
	var admin models.Admin

	if err := c.ShouldBindJSON(&admin); err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusBadRequest, models.ErrServerProblems, false)
		return
	}

	if err := r.t.EditAdmin(c.Request.Context(), &entity.Admin{
		Id:       admin.ID,
		Username: admin.Username,
		Password: admin.Password,
		Avatar:   admin.Avatar,
	}); err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusBadRequest, models.ErrServerProblems, false)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Admin muvaffaqiyatli yangilandi.",
		"status":  true,
	})
}

// @Summary     Edit Super Admin
// @Description Superadmin updates by its id.
// @ID          edit-super-admin
// @Tags  	    superadmin
// @Accept      json
// @Produce     json
// @Param       request body models.SuperAdmin true "Superadmin data for update"
// @Success     200 {object} models.Message
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Security    BearerAuth
// @Router      /auth/superadmin/edit [put]
func (r *authRoutes) editSuperAdmin(c *gin.Context) {
	var admin models.SuperAdmin

	if err := c.ShouldBindJSON(&admin); err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusBadRequest, models.ErrServerProblems, false)
		return
	}

	jwt := tokens.JWTHandler{
		Token:     c.Request.Header.Get("Authorization"),
		SigninKey: "dfhdghkglioe",
	}

	claims, err := jwt.ExtractClaims()
	if err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusBadRequest, models.ErrServerProblems, false)
		return
	}

	id := claims["sub"]

	if err := r.t.ChangeSuperAdminData(c.Request.Context(), &entity.SuperAdmin{
		Id:          id.(string),
		PhoneNumber: admin.PhoneNumber,
		Password:    admin.Password,
		Avatar:      admin.Avatar,
	}); err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusBadRequest, models.ErrServerProblems, false)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Admin muvaffaqiyatli yangilandi.",
		"status":  true,
	})
}

type File struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// @Summary       File upload
// @Description   API for file upload including images and videos
// @Tags          file-upload
// @Accept        multipart/form-data
// @Produce       json
// @Param         file formData file true "File"
// @Param         type formData string true "Bucket type to put file"
// @Success       200 {object} string
// @Failure       400 {object} string
// @Failure       500 {object} string
// @Router        /file/upload [post]
func (f *authRoutes) upload(c *gin.Context) {
	c.Request.ParseMultipartForm(500 << 20) // 500MB

	bucketType := c.Request.FormValue("type")
	if bucketType == "" {
		errorResponse(c, http.StatusBadRequest, "Bucket type is required", false)
		return
	}

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Failed to retrieve file", false)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedImageExts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".pdf": true}
	allowedVoiceExts := map[string]bool{".mp3": true, ".wav": true, ".ogg": true}
	allowedVideoExts := map[string]bool{".mp4": true, ".avi": true, ".mov": true, ".wmv": true, ".flv": true, ".mkv": true, ".webm": true}

	if fileHeader.Size > 20<<20 {
		errorResponse(c, http.StatusBadRequest, "File size exceeds 20MB limit", false)
		return
	}

	minioClient := f.getMinioClient()
	id := uuid.New().String()
	objectName := id + ext

	// Check if the file is an image
	if allowedImageExts[ext] {
		img, _, err := image.Decode(file)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Failed to decode image", false)
			return
		}

		// Create a channel for the original and final image processing
		done := make(chan error, 2)

		// Handle original image upload in a separate goroutine
		go func() {
			var buf bytes.Buffer
			err := jpeg.Encode(&buf, img, nil)
			if err != nil {
				done <- err
				return
			}

			_, err = minioClient.PutObject(context.Background(), bucketType, objectName, bytes.NewReader(buf.Bytes()), int64(buf.Len()), minio.PutObjectOptions{
				ContentType: "image/jpeg",
			})
			done <- err
		}()

		// Handle resizing image to fit within 1920x1080
		go func() {
			originalWidth := img.Bounds().Dx()
			originalHeight := img.Bounds().Dy()

			// Calculate new dimensions while preserving aspect ratio
			ratio := math.Min(float64(1920)/float64(originalWidth), float64(1080)/float64(originalHeight))
			newWidth := int(float64(originalWidth) * ratio)
			newHeight := int(float64(originalHeight) * ratio)
			resizedImg := resize.Resize(uint(newWidth), uint(newHeight), img, resize.Lanczos3)

			// Create a 1920x1080 gray background
			whiteBackground := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
			gray := color.RGBA{R: 229, G: 231, B: 235, A: 255} // Define gray-200 color
			draw.Draw(whiteBackground, whiteBackground.Bounds(), &image.Uniform{gray}, image.Point{}, draw.Src)

			// Center the resized image on the gray background
			offset := image.Pt((1920-newWidth)/2, (1080-newHeight)/2)
			draw.Draw(whiteBackground, resizedImg.Bounds().Add(offset), resizedImg, image.Point{}, draw.Over)

			var buf bytes.Buffer
			err := jpeg.Encode(&buf, whiteBackground, nil)
			if err != nil {
				done <- err
				return
			}

			finalObjectName := id + "_fitted" + ext
			_, err = minioClient.PutObject(context.Background(), bucketType, finalObjectName, bytes.NewReader(buf.Bytes()), int64(buf.Len()), minio.PutObjectOptions{
				ContentType: "image/jpeg",
			})
			done <- err
		}()

		// Wait for both uploads to finish
		for i := 0; i < 2; i++ {
			if err := <-done; err != nil {
				errorResponse(c, http.StatusInternalServerError, "Failed to upload image", false)
				return
			}
		}

		endpoint := os.Getenv("SERVER_IP")
		finalURL := fmt.Sprintf("https://%s/%s/%s", endpoint, bucketType, id+"_fitted"+ext)

		c.JSON(http.StatusOK, gin.H{
			"image_url": finalURL,
			"status":    true,
		})

		return
	}

	// Check if the file is a voice file
	if allowedVoiceExts[ext] {
		_, err = minioClient.PutObject(context.Background(), bucketType, objectName, file, fileHeader.Size, minio.PutObjectOptions{
			ContentType: "audio/mpeg",
		})
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to upload voice file", false)
			return
		}

		endpoint := os.Getenv("SERVER_IP")
		voiceURL := fmt.Sprintf("https://%s/%s/%s", endpoint, bucketType, objectName)

		c.JSON(http.StatusOK, gin.H{
			"voice_url": voiceURL,
			"status":    true,
		})

		return
	}

	// Check if the file is a video file
	if allowedVideoExts[ext] {
		_, err = minioClient.PutObject(context.Background(), bucketType, objectName, file, fileHeader.Size, minio.PutObjectOptions{
			ContentType: "video/" + ext[1:],
		})
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to upload video file", false)
			return
		}

		endpoint := os.Getenv("SERVER_IP")
		videoURL := fmt.Sprintf("https://%s/%s/%s", endpoint, bucketType, objectName)

		c.JSON(http.StatusOK, gin.H{
			"video_url": videoURL,
			"status":    true,
		})

		return
	}

	errorResponse(c, http.StatusBadRequest, "Unsupported file type", false)
}

// getMinioClient initializes and reuses MinIO client
func (f *authRoutes) getMinioClient() *minio.Client {
	minioClient, err := minio.New("minio:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("nodirbek", "nodirbek", ""),
		Secure: false,
	})
	if err != nil {
		f.l.Error(err, "Failed to initialize MinIO client")
	}
	return minioClient
}
func (r *authRoutes) HandleFailedAttempt(ctx context.Context, phoneNumber string) error {
	failedAttempts++

	if failedAttempts >= 3 {
		err := r.t.BlockSuperAdmin(ctx)
		if err != nil {
			return err
		}
		isBlocked = true
	}

	return nil
}

func (r *authRoutes) ResetFailedAttempts(ctx context.Context, phoneNumber string) error {
	failedAttempts = 0
	isBlocked = false
	return nil
}
