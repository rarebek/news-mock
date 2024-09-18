package models

type AdminLoginRequest struct {
	Username string `json:"username" example:"test"`
	Password string `json:"password" example:"test"`
	Avatar   string `json:"avatar"`
}

type Admin struct {
	ID       string `json:"id"`
	Username string `json:"username" example:"test"`
	Password string `json:"password" example:"test"`
	Avatar   string `json:"avatar"`
}

type SuperAdminLoginRequest struct {
	PhoneNumber string `json:"phone_number" example:"test"`
	Password    string `json:"password" example:"test"`
	Avatar      string `json:"avatar"`
}

type AdminLoginResponse struct {
	AccessToken string `json:"access_token"`
}

type Message struct {
	Message string `json:"message"`
}

type CategoryResponse struct {
	ID            string                `json:"id"`
	Name          string                `json:"name"`
	SubCategories []SubCategoryResponse `json:"sub_categories"`
}

type SubCategoryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SuperAdmin struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	Avatar      string `json:"avatar"`
}
