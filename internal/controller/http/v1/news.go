package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"

	"tarkib.uz/internal/controller/http/models"
	"tarkib.uz/internal/entity"
	"tarkib.uz/internal/usecase"
	"tarkib.uz/pkg/logger"
)

type newsRoutes struct {
	t usecase.NewsUseCase
	l logger.Interface
}

func newNewsRoutes(handler *gin.RouterGroup, t usecase.NewsUseCase, l logger.Interface) {
	r := &newsRoutes{t, l}

	h := handler.Group("/news")
	{
		h.POST("/create", r.create)
		h.GET("/getall", r.getAllNews)
		h.DELETE("/delete/:id", r.deleteNews)
		h.GET("/filtered", r.getFilteredNews)
		h.PUT("/update/:id", r.updateNews)
		h.GET("/get/:id", r.getNewsByID)
		h.GET("/search", r.searchGlobalWithLocal)
		h.GET("/convert", r.CurrencyConverter)
		h.GET("/financialData", r.GetFinancialData)
		h.GET("/weatherData", r.GetWeatherData)
		h.GET("/currencies", r.GetCurrencyCodes)
	}
}

// @Summary     Create News
// @Description This method for creating a news
// @ID          create-news
// @Tags  	    news
// @Accept      json
// @Produce     json
// @Param       request body models.News true "News details"
// @Success     200 {object} models.Message
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Security    BearerAuth
// @Router      /news/create [post]
func (n *newsRoutes) create(c *gin.Context) {
	var body models.News
	if err := c.ShouldBindJSON(&body); err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusBadRequest, err.Error(), false)
		return
	}

	if err := n.t.CreateNews(c.Request.Context(),
		&entity.News{
			Name:           body.UzName,
			Description:    body.UzDescription,
			ImageURL:       body.ImageURL,
			VoiceURL:       body.VoiceURL,
			SubCategoryIDs: body.SubCategoryIDs,
			SiteImageLink:  body.SiteImageLink,
			Links:          body.Links,
			Language:       "uz",
			VideoURL:       body.VideoURL,
			Text:           body.UzText,
		},
		&entity.News{
			Name:           body.RuName,
			Description:    body.RuDescription,
			ImageURL:       body.ImageURL,
			VoiceURL:       body.VoiceURL,
			SubCategoryIDs: body.SubCategoryIDs,
			SiteImageLink:  body.SiteImageLink,
			Links:          body.Links,
			Language:       "ru",
			VideoURL:       body.VideoURL,
			Text:           body.RuText,
		}); err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Kechirasiz, serverda muammolar bo'lyapti", false)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Yangilik muvaffaqiyatli yaratildi.",
		"status":  true,
	})
}

// @Summary     Delete News
// @Description This method deleting news
// @ID          delete-news
// @Tags  	    news
// @Accept      json
// @Produce     json
// @Param       id path string true "ID of the news to delete"
// @Success     200 {object} models.Message
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Security    BearerAuth
// @Router      /news/delete/{id} [delete]
func (n *newsRoutes) deleteNews(c *gin.Context) {
	id := c.Param("id")
	if err := n.t.DeleteNews(c.Request.Context(), id); err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Kechirasiz, serverda muammolar bo'lyapti", false)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Yangilik muvaffaqiyatli o'chirildi.",
		"status":  true,
	})
}

// @Summary		Get All News
// @Description This method retrieves all news with pagination.
// @ID          get-all-news
// @Tags  	    news
// @Accept      json
// @Produce     json
// @Param       page  query int true  "Page number"
// @Param       limit query int true  "Number of items per page"
// @Param       language query string true  "Language of news"
// @Success     200 {object} []models.News
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /news/getall [get]
func (n *newsRoutes) getAllNews(c *gin.Context) {
	var (
		page  string
		limit string
	)

	page = c.Query("page")
	limit = c.Query("limit")
	language := c.Query("language")

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusBadRequest, models.ErrServerProblems, false)
		return
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusBadRequest, models.ErrServerProblems, false)
		return
	}

	news, err := n.t.GetAllNews(c.Request.Context(), &entity.GetAllNewsRequest{
		Page:  pageInt,
		Limit: limitInt,
	}, language)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, models.ErrServerProblems, false)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"news":   news,
		"status": true,
	})
}

// @Summary		Get Filtered News
// @Description This method retrieves news based on optional filters (subcategory IDs and category ID) with pagination.
// @ID          get-filtered-news
// @Tags  	    news
// @Accept      json
// @Produce     json
// @Param       sub_category_ids query []string false "List of subcategory IDs"
// @Param       category_id     query string   false "Category ID"
// @Param       language query string true  "Language of news"
// @Param       page            query int      true  "Page number"
// @Param       limit           query int      true  "Number of items per page"
// @Param       search          query string   false  "Search term"
// @Success     200 {object} []models.News
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /news/filtered [get]
func (n *newsRoutes) getFilteredNews(c *gin.Context) {
	subCategoryIDs := c.QueryArray("sub_category_ids")
	categoryID := c.Query("category_id")
	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	searchTerm := c.Query("search")
	language := c.Query("language")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusBadRequest, "Invalid page number", false)
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusBadRequest, "Invalid limit number", false)
		return
	}

	news, err := n.t.GetFilteredNews(c.Request.Context(), &entity.GetFilteredNewsRequest{
		SubCategoryIDs: subCategoryIDs,
		CategoryID:     categoryID,
		Page:           page,
		Limit:          limit,
		SearchTerm:     searchTerm,
	}, language)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Kechirasiz, serverda muammolar bo'lyapti", false)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"news":   news,
		"status": true,
	})
}

// @Summary     Update News
// @Description This method updates an existing news item
// @ID          update-news
// @Tags  	    news
// @Accept      json
// @Produce     json
// @Param       id      path   string      true  "ID of the news to update"
// @Param       request body   models.NewsOneLang true  "Updated news details"
// @Success     200 {object} models.Message
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Security    BearerAuth
// @Router      /news/update/{id} [put]
func (n *newsRoutes) updateNews(c *gin.Context) {
	id := c.Param("id")
	var body models.NewsOneLang
	if err := c.ShouldBindJSON(&body); err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusBadRequest, err.Error(), false)
		return
	}
	

	if err := n.t.UpdateNews(c.Request.Context(), id, &entity.News{
		Name:           body.Name,
		Description:    body.Description,
		ImageURL:       body.ImageURL,
		VoiceURL:       body.VoiceURL,
		SubCategoryIDs: body.SubCategoryIDs,
		Links:          body.Links,
		SiteImageLink:  body.SiteImgeLink,
		VideoURL:       body.VideoURL,
		Text:           body.Text,
	}); err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Kechirasiz, serverda muammolar bo'lyapti", false)
		return
	}

	pp.Println(body)

	c.JSON(http.StatusOK, gin.H{
		"message": "Yangilik muvaffaqiyatli yangilandi.",
		"status":  true,
	})
}

// @Summary     Get News By ID
// @Description This method retrieves a news item by its ID
// @ID          get-news-by-id
// @Tags  	    news
// @Accept      json
// @Produce     json
// @Param       id path string true "ID of the news to retrieve"
// @Success     200 {object} models.News
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /news/get/{id} [get]
func (n *newsRoutes) getNewsByID(c *gin.Context) {
	id := c.Param("id")
	news, err := n.t.GetNewsByID(c.Request.Context(), id)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Kechirasiz, serverda muammolar bo'lyapti", false)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"news":   news,
		"status": true,
	})
}

// @Summary		Global search
// @Description This method for searching globally and from our data.
// @ID          get-filtered-news-global
// @Tags  	    news
// @Accept      json
// @Produce     json
// @Param       sub_category_ids query []string false "List of subcategory IDs"
// @Param       category_id     query string   false "Category ID"
// @Param       page            query int      true  "Page number"
// @Param       language query string true  "Language of news"
// @Param       limit           query int      true  "Number of items per page"
// @Param       search          query string   false  "Search term"
// @Success     200 {object} map[string]interface{} "Returns news, global_link, and status"
// @Failure     400 {object} response "Bad request"
// @Failure     500 {object} response "Internal server error"
// @Router      /news/search [get]
func (n *newsRoutes) searchGlobalWithLocal(c *gin.Context) {
	subCategoryIDs := c.QueryArray("sub_category_ids")
	categoryID := c.Query("category_id")
	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	searchTerm := c.Query("search")
	language := c.Query("language")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusBadRequest, "Invalid page number", false)
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusBadRequest, "Invalid limit number", false)
		return
	}

	news, err := n.t.GetFilteredNews(c.Request.Context(), &entity.GetFilteredNewsRequest{
		SubCategoryIDs: subCategoryIDs,
		CategoryID:     categoryID,
		Page:           page,
		Limit:          limit,
		SearchTerm:     searchTerm,
	}, language)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Kechirasiz, serverda muammolar bo'lyapti", false)
		return
	}

	if news == nil {
		var globalLink string
		if searchTerm != "" {
			globalLink = "https://www.google.com/search?q=" + url.QueryEscape(searchTerm)
		}

		c.JSON(http.StatusOK, gin.H{
			"global_link": globalLink,
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"news":   news,
		"status": true,
	})
}

// Currency struct to unmarshal JSON data
type Currency struct {
	ID       int    `json:"id"`
	Code     string `json:"Code"`
	Ccy      string `json:"Ccy"`
	CcyNmRU  string `json:"CcyNm_RU"`
	CcyNmUZ  string `json:"CcyNm_UZ"`
	CcyNmUZC string `json:"CcyNm_UZC"`
	CcyNmEN  string `json:"CcyNm_EN"`
	Nominal  string `json:"Nominal"`
	Rate     string `json:"Rate"`
	Diff     string `json:"Diff"`
	Date     string `json:"Date"`
}

// @Summary		Currency Converter
// @Description Converts an amount from one currency to another based on the latest exchange rates.
// @ID          currency-converter
// @Tags  	    currency
// @Accept      json
// @Produce     json
// @Param       from   query string true "Currency code to convert from" example("USD")
// @Param       to     query string true "Currency code to convert to" example("UZS")
// @Param       amount query string true "Amount to be converted" example("100")
// @Success     200 {object} map[string]interface{} "Returns original amount, converted amount, from currency, and to currency"
// @Failure     400 {object} response "Bad request"
// @Failure     500 {object} response "Internal server error"
// @Router      /news/convert [get]
func (n *newsRoutes) CurrencyConverter(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	amountStr := c.Query("amount")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusBadRequest, "Invalid amount value", false)
		return
	}

	resp, err := http.Get("https://cbu.uz/uz/arkhiv-kursov-valyut/json/")
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Kechirasiz, serverda muammolar bo'lyapti", false)
		return
	}
	defer resp.Body.Close()

	var currencies []Currency
	if err := json.NewDecoder(resp.Body).Decode(&currencies); err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Failed to parse currency data", false)
		return
	}

	// If converting from UZS to another currency
	if from == "UZS" {
		for _, currency := range currencies {
			if currency.Ccy == to {
				rateOfTo := currency.Rate
				rateFloat, err := strconv.ParseFloat(rateOfTo, 64)
				if err != nil {
					n.l.Error(err)
					errorResponse(c, http.StatusInternalServerError, "Failed to parse currency data", false)
					return
				}
				result := amount / rateFloat

				c.JSON(http.StatusOK, gin.H{
					"from":            from,
					"to":              to,
					"originalAmount":  amount,
					"convertedAmount": result,
				})
				return
			}
		}
		errorResponse(c, http.StatusBadRequest, "Currency not found", false)
		return
	}

	// If converting to UZS from another currency
	if to == "UZS" {
		for _, currency := range currencies {
			if currency.Ccy == from {
				rateOfFrom := currency.Rate
				rateFloat, err := strconv.ParseFloat(rateOfFrom, 64)
				if err != nil {
					n.l.Error(err)
					errorResponse(c, http.StatusInternalServerError, "Failed to parse currency data", false)
					return
				}
				result := amount * rateFloat

				c.JSON(http.StatusOK, gin.H{
					"from":            from,
					"to":              to,
					"originalAmount":  amount,
					"convertedAmount": result,
				})
				return
			}
		}
		errorResponse(c, http.StatusBadRequest, "Currency not found", false)
		return
	}

	// If converting between two foreign currencies
	fromRate, err := findRate(currencies, from)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusBadRequest, err.Error(), false)
		return
	}

	toRate, err := findRate(currencies, to)
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusBadRequest, err.Error(), false)
		return
	}

	convertedAmount := (amount * fromRate) / toRate

	c.JSON(http.StatusOK, gin.H{
		"from":            from,
		"to":              to,
		"originalAmount":  amount,
		"convertedAmount": convertedAmount,
	})
}

func findRate(currencies []Currency, code string) (float64, error) {
	for _, currency := range currencies {
		if currency.Ccy == code {
			rate, err := strconv.ParseFloat(currency.Rate, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid rate for currency %s", code)
			}
			return rate, nil
		}
	}
	return 0, fmt.Errorf("currency not found: %s", code)
}

type FinancialData struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	UpdatedAt string  `json:"updatedAt"`
}

// GetFinancialData godoc
// @Summary Get financial data for various symbols
// @Description Fetches financial data for symbols such as gold, silver, and bitcoin from external APIs
// @Tags Financial
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Returns financial and currency data"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /news/financialData [get]
func (n *newsRoutes) GetFinancialData(c *gin.Context) {
	var financialDatas []FinancialData
	symbols := []string{"XAU", "XAG", "BTC"}

	for _, v := range symbols {
		var onedata FinancialData
		resp, err := http.Get("https://api.gold-api.com/price/" + v)
		if err != nil {
			n.l.Error(err)
			errorResponse(c, http.StatusInternalServerError, "Kechirasiz, serverda muammolar bo'lyapti", false)
			return
		}

		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&onedata); err != nil {
			n.l.Error(err)
			errorResponse(c, http.StatusInternalServerError, "Failed to parse currency data", false)
			return
		}

		financialDatas = append(financialDatas, onedata)
	}

	// Fetch currency data for specific currencies
	currencies, err := fetchCurrencies([]string{"USD", "EUR", "RUB", "KZT", "KGS", "SAR", "JPY", "BYN", "EGP", "PKR", "SGD", "TJS", "AFN", "IRR", "QAR", "TRY", "XDR", "UAH"})
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Failed to fetch currency data", false)
		return
	}

	var uzCurrency = Currency{
		Ccy:      "UZS",
		CcyNmRU:  "Узбекский сум",
		CcyNmUZ:  "O'zbekiston so'mi",
		CcyNmUZC: "",
		CcyNmEN:  "Uzbek sum",
		Nominal:  "",
		Rate:     "",
		Diff:     "",
		Date:     "",
	}

	currencies = append([]Currency{uzCurrency}, currencies...)

	c.JSON(http.StatusOK, gin.H{
		"financial": financialDatas,
		"currency":  currencies,
	})
}

// fetchCurrencies fetches currency data for a list of currency codes
func fetchCurrencies(codes []string) ([]Currency, error) {
	resp, err := http.Get("https://cbu.uz/uz/arkhiv-kursov-valyut/json/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var allCurrencies []Currency
	if err := json.NewDecoder(resp.Body).Decode(&allCurrencies); err != nil {
		return nil, err
	}

	var selectedCurrencies []Currency
	for _, currency := range allCurrencies {
		for _, code := range codes {
			if currency.Ccy == code {
				selectedCurrencies = append(selectedCurrencies, currency)
				break
			}
		}
	}

	return selectedCurrencies, nil
}

type WeatherData struct {
	Temperature float64 `json:"temperature_2m"`
	WeatherCode int     `json:"weathercode"`
}

// @Summary Get weather data
// @Description Fetches current weather data for a specified location using the Open-Meteo API
// @Tags Weather
// @Accept json
// @Produce json
// @Param latitude query float64 true "Latitude of the location" example(40.7128)
// @Param longitude query float64 true "Longitude of the location" example(-74.0060)
// @Success 200 {object} WeatherResponse "Returns current weather data"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /news/weatherData [get]
func (n *newsRoutes) GetWeatherData(c *gin.Context) {
	latitudeStr := c.Query("latitude")
	longitudeStr := c.Query("longitude")

	if latitudeStr == "" || longitudeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Latitude and longitude are required"})
		return
	}

	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&hourly=temperature_2m,precipitation,wind_speed_10m", latitudeStr, longitudeStr)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Add("accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weather data"})
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weather data"})
		return
	}

	var weatherData struct {
		Hourly struct {
			Time          []string  `json:"time"`
			Temperature2m []float64 `json:"temperature_2m"`
			Precipitation []float64 `json:"precipitation"`
			WindSpeed10m  []float64 `json:"wind_speed_10m"`
		} `json:"hourly"`
	}

	// Decode response
	if err := json.NewDecoder(res.Body).Decode(&weatherData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode weather data"})
		return
	}

	// Group data by date
	dailyData := groupByDate(
		weatherData.Hourly.Time,
		weatherData.Hourly.Temperature2m,
		weatherData.Hourly.Precipitation,
		weatherData.Hourly.WindSpeed10m,
	)

	// Extract and sort the dates
	dates := make([]string, 0, len(dailyData))
	for date := range dailyData {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	// Generate the forecast in order
	forecast := make([]map[string]interface{}, 0, len(dates))
	for _, date := range dates {
		data := dailyData[date]
		daytimeTemps, eveningTemps, nightTemps := splitByTimeOfDay(data.Time, data.Temperature2m)
		daytimePrecip, eveningPrecip, nightPrecip := splitByTimeOfDay(data.Time, data.Precipitation)
		daytimeWind, eveningWind, nightWind := splitByTimeOfDay(data.Time, data.WindSpeed10m)

		forecast = append(forecast, map[string]interface{}{
			"date": date,
			"daytime": WeatherForecast{
				Temperature:   calculateAverage(daytimeTemps),
				Precipitation: calculateAverage(daytimePrecip),
				WindSpeed:     calculateAverage(daytimeWind),
			},
			"evening": WeatherForecast{
				Temperature:   calculateAverage(eveningTemps),
				Precipitation: calculateAverage(eveningPrecip),
				WindSpeed:     calculateAverage(eveningWind),
			},
			"night": WeatherForecast{
				Temperature:   calculateAverage(nightTemps),
				Precipitation: calculateAverage(nightPrecip),
				WindSpeed:     calculateAverage(nightWind),
			},
		})
	}

	// Response
	c.JSON(http.StatusOK, gin.H{
		"today":    forecast[0],  // Assuming today is the first entry
		"forecast": forecast[1:], // The rest of the days
	})
}

func groupByDate(times []string, temperatures, precipitations, windSpeeds []float64) map[string]struct {
	Time          []string
	Temperature2m []float64
	Precipitation []float64
	WindSpeed10m  []float64
} {
	dailyData := make(map[string]struct {
		Time          []string
		Temperature2m []float64
		Precipitation []float64
		WindSpeed10m  []float64
	})

	for i, timeStr := range times {
		date := timeStr[:10] // Extract the date part (YYYY-MM-DD)
		dayData := dailyData[date]
		dayData.Time = append(dayData.Time, timeStr)
		dayData.Temperature2m = append(dayData.Temperature2m, temperatures[i])
		dayData.Precipitation = append(dayData.Precipitation, precipitations[i])
		dayData.WindSpeed10m = append(dayData.WindSpeed10m, windSpeeds[i])
		dailyData[date] = dayData
	}

	return dailyData
}

func splitByTimeOfDay(times []string, data []float64) (daytime []float64, evening []float64, night []float64) {
	for i, timeStr := range times {
		hour := extractHourFromTimeString(timeStr)
		if hour >= 6 && hour < 18 {
			daytime = append(daytime, data[i])
		} else if hour >= 18 && hour < 24 {
			evening = append(evening, data[i])
		} else {
			night = append(night, data[i])
		}
	}
	return
}

func extractHourFromTimeString(timeStr string) int {
	hourStr := timeStr[11:13]
	hour, _ := strconv.Atoi(hourStr)
	return hour
}

func calculateAverage(temps []float64) float64 {
	if len(temps) == 0 {
		return 0
	}
	var sum float64
	for _, temp := range temps {
		sum += temp
	}
	return sum / float64(len(temps))
}

type WeatherForecast struct {
	Temperature   float64 `json:"temperature"`
	Precipitation float64 `json:"precipitation"`
	WindSpeed     float64 `json:"wind_speed"`
}

type WeatherResponse struct {
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	GenerationTimeMs     float64 `json:"generationtime_ms"`
	UTCOffsetSeconds     int     `json:"utc_offset_seconds"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	Elevation            float64 `json:"elevation"`
	HourlyUnits          struct {
		Time          string `json:"time"`
		Temperature2m string `json:"temperature_2m"`
	} `json:"hourly_units"`
	Hourly struct {
		Time          []string  `json:"time"`
		Temperature2m []float64 `json:"temperature_2m"`
	} `json:"hourly"`
}

// @Summary		Get Currency Codes
// @Description Returns a list of currency codes and their names in Uzbek.
// @ID          get-currency-codes
// @Tags  	    currency
// @Accept      json
// @Produce     json
// @Success     200 {array} map[string]string "Returns a list of currency codes and their Uzbek names"
// @Failure     500 {object} response "Internal server error"
// @Router      /news/currencies [get]
func (n *newsRoutes) GetCurrencyCodes(c *gin.Context) {
	resp, err := http.Get("https://cbu.uz/uz/arkhiv-kursov-valyut/json/")
	if err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Kechirasiz, serverda muammolar bo'lyapti", false)
		return
	}
	defer resp.Body.Close()

	var currencies []Currency
	if err := json.NewDecoder(resp.Body).Decode(&currencies); err != nil {
		n.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Failed to parse currency data", false)
		return
	}

	currencyList := make([]map[string]string, 0, len(currencies))
	for _, currency := range currencies {
		currencyList = append(currencyList, map[string]string{
			"code": currency.Ccy,
			"name": currency.CcyNmUZ,
		})
	}
	uz := map[string]string{
		"code": "UZS",
		"name": "O'zbekiston so'mi",
	}
	currencyList = append(currencyList, uz)

	c.JSON(http.StatusOK, currencyList)
}
