package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"

	blogpb "toxictoast/services/blog-service/api/proto"
	foodfoliopb "toxictoast/services/foodfolio-service/api/proto"
	weatherpb "toxictoast/services/gateway-service/api/proto/weather"
)

// MirrorHandler handles mirror dashboard requests
type MirrorHandler struct {
	weatherClient   weatherpb.WeatherServiceClient
	foodfolioClient foodfoliopb.ShoppinglistServiceClient
	blogClient      blogpb.BlogServiceClient
	serviceClients  *ServiceClients
}

// ServiceClients holds all available service connections for health checking
type ServiceClients struct {
	WeatherConn   *grpc.ClientConn
	FoodfolioConn *grpc.ClientConn
	BlogConn      *grpc.ClientConn
}

// NewMirrorHandler creates a new mirror handler
func NewMirrorHandler(weatherConn, foodfolioConn, blogConn *grpc.ClientConn) *MirrorHandler {
	return &MirrorHandler{
		weatherClient:   weatherpb.NewWeatherServiceClient(weatherConn),
		foodfolioClient: foodfoliopb.NewShoppinglistServiceClient(foodfolioConn),
		blogClient:      blogpb.NewBlogServiceClient(blogConn),
		serviceClients: &ServiceClients{
			WeatherConn:   weatherConn,
			FoodfolioConn: foodfolioConn,
			BlogConn:      blogConn,
		},
	}
}

// MirrorDashboardResponse is the aggregated response for the mirror
type MirrorDashboardResponse struct {
	Timestamp time.Time      `json:"timestamp"`
	Weather   *WeatherData   `json:"weather,omitempty"`
	Shopping  *ShoppingData  `json:"shopping,omitempty"`
	Blog      *BlogData      `json:"blog,omitempty"`
	Services  *ServiceStatus `json:"services,omitempty"`
	Calendar  *CalendarData  `json:"calendar,omitempty"` // Future
}

// WeatherData contains weather information
type WeatherData struct {
	Current  *CurrentWeather  `json:"current"`
	Forecast []DailyForecast  `json:"forecast"`
}

// CurrentWeather contains current weather conditions
type CurrentWeather struct {
	Temperature        float64 `json:"temperature"`
	ApparentTemp       float64 `json:"apparentTemperature"`
	WeatherDescription string  `json:"weatherDescription"`
	WeatherCode        int     `json:"weatherCode"`
	WindSpeed          float64 `json:"windSpeed"`
	WindDirection      int     `json:"windDirection"`
	Humidity           int     `json:"humidity"`
	Precipitation      float64 `json:"precipitation"`
	CloudCover         int     `json:"cloudCover"`
}

// DailyForecast contains forecast for one day
type DailyForecast struct {
	Date                     string  `json:"date"`
	TempMax                  float64 `json:"temperatureMax"`
	TempMin                  float64 `json:"temperatureMin"`
	WeatherDescription       string  `json:"weatherDescription"`
	WeatherCode              int     `json:"weatherCode"`
	PrecipitationSum         float64 `json:"precipitationSum"`
	PrecipitationProbability float64 `json:"precipitationProbability"`
	WindSpeedMax             float64 `json:"windSpeedMax"`
	Sunrise                  string  `json:"sunrise"`
	Sunset                   string  `json:"sunset"`
}

// CalendarData - placeholder for future
type CalendarData struct {
	Events []interface{} `json:"events"`
}

// ShoppingData contains shopping list information
type ShoppingData struct {
	ListID        string         `json:"listId"`
	ListName      string         `json:"listName"`
	TotalItems    int            `json:"totalItems"`
	PurchasedItems int           `json:"purchasedItems"`
	PendingItems  int            `json:"pendingItems"`
	Items         []ShoppingItem `json:"items"`
}

// ShoppingItem represents a single item on the shopping list
type ShoppingItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Quantity    int     `json:"quantity"`
	IsPurchased bool    `json:"isPurchased"`
	Category    string  `json:"category,omitempty"`
}

// BlogData contains recent blog posts
type BlogData struct {
	TotalPosts int        `json:"totalPosts"`
	Latest     []BlogPost `json:"latest"`
}

// BlogPost represents a blog post summary
type BlogPost struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Excerpt     string `json:"excerpt,omitempty"`
	Author      string `json:"author,omitempty"`
	PublishedAt string `json:"publishedAt"`
}

// ServiceStatus contains status of all services
type ServiceStatus struct {
	Total     int           `json:"total"`
	Healthy   int           `json:"healthy"`
	Unhealthy int           `json:"unhealthy"`
	Services  []ServiceInfo `json:"services"`
}

// ServiceInfo represents status of a single service
type ServiceInfo struct {
	Name      string `json:"name"`
	Status    string `json:"status"` // "healthy" or "unhealthy"
	LastCheck string `json:"lastCheck"`
}

// GetDashboard returns aggregated dashboard data
func (h *MirrorHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	// Get query parameters for location
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	timezone := r.URL.Query().Get("timezone")

	// Default to Frankfurt if not provided
	if latStr == "" {
		latStr = "50.1109"
	}
	if lonStr == "" {
		lonStr = "8.6821"
	}
	if timezone == "" {
		timezone = "Europe/Berlin"
	}

	// Parse coordinates
	var lat, lon float64
	if _, err := fmt.Sscanf(latStr, "%f", &lat); err != nil {
		http.Error(w, "Invalid latitude", http.StatusBadRequest)
		return
	}
	if _, err := fmt.Sscanf(lonStr, "%f", &lon); err != nil {
		http.Error(w, "Invalid longitude", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response := &MirrorDashboardResponse{
		Timestamp: time.Now(),
	}

	// Fetch weather data
	weatherData, err := h.fetchWeather(ctx, lat, lon, timezone)
	if err != nil {
		log.Printf("Failed to fetch weather: %v", err)
		// Continue without weather data
	} else {
		response.Weather = weatherData
	}

	// Fetch shopping list from foodfolio-service
	shoppingData, err := h.fetchShopping(ctx)
	if err != nil {
		log.Printf("Failed to fetch shopping list: %v", err)
		// Continue without shopping data
	} else {
		response.Shopping = shoppingData
	}

	// Fetch blog posts
	blogData, err := h.fetchBlog(ctx)
	if err != nil {
		log.Printf("Failed to fetch blog posts: %v", err)
		// Continue without blog data
	} else {
		response.Blog = blogData
	}

	// Fetch service status
	serviceStatus := h.fetchServiceStatus()
	response.Services = serviceStatus

	// TODO: Fetch calendar data when calendar-service is ready

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// fetchWeather gets current weather and forecast
func (h *MirrorHandler) fetchWeather(ctx context.Context, lat, lon float64, timezone string) (*WeatherData, error) {
	weatherData := &WeatherData{}

	// Fetch current weather
	currentReq := &weatherpb.WeatherRequest{
		Latitude:  lat,
		Longitude: lon,
		Timezone:  timezone,
	}

	currentResp, err := h.weatherClient.GetCurrentWeather(ctx, currentReq)
	if err != nil {
		return nil, err
	}

	weatherData.Current = &CurrentWeather{
		Temperature:        currentResp.Temperature,
		ApparentTemp:       currentResp.ApparentTemperature,
		WeatherDescription: currentResp.WeatherDescription,
		WeatherCode:        int(currentResp.WeatherCode),
		WindSpeed:          currentResp.WindSpeed,
		WindDirection:      int(currentResp.WindDirection),
		Humidity:           int(currentResp.Humidity),
		Precipitation:      currentResp.Precipitation,
		CloudCover:         int(currentResp.CloudCover),
	}

	// Fetch forecast (3 days)
	forecastReq := &weatherpb.ForecastRequest{
		Latitude:  lat,
		Longitude: lon,
		Days:      3,
		Timezone:  timezone,
	}

	forecastResp, err := h.weatherClient.GetForecast(ctx, forecastReq)
	if err != nil {
		return nil, err
	}

	weatherData.Forecast = make([]DailyForecast, 0, len(forecastResp.Daily))
	for _, day := range forecastResp.Daily {
		forecast := DailyForecast{
			Date:                     day.Date.AsTime().Format("2006-01-02"),
			TempMax:                  day.TemperatureMax,
			TempMin:                  day.TemperatureMin,
			WeatherDescription:       day.WeatherDescription,
			WeatherCode:              int(day.WeatherCode),
			PrecipitationSum:         day.PrecipitationSum,
			PrecipitationProbability: day.PrecipitationProbability,
			WindSpeedMax:             day.WindSpeedMax,
			Sunrise:                  day.Sunrise.AsTime().Format(time.RFC3339),
			Sunset:                   day.Sunset.AsTime().Format(time.RFC3339),
		}
		weatherData.Forecast = append(weatherData.Forecast, forecast)
	}

	return weatherData, nil
}

// fetchShopping gets the first active shopping list with all items
func (h *MirrorHandler) fetchShopping(ctx context.Context) (*ShoppingData, error) {
	// List all shopping lists (first page, non-deleted only)
	listReq := &foodfoliopb.ListShoppinglistsRequest{
		Page:     1,
		PageSize: 10,
		DeletedFilter: &foodfoliopb.DeletedFilter{
			IncludeDeleted: false,
			OnlyDeleted:    false,
		},
	}

	listResp, err := h.foodfolioClient.ListShoppinglists(ctx, listReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list shopping lists: %w", err)
	}

	// Check if we have any lists
	if len(listResp.Shoppinglists) == 0 {
		// Return empty shopping data if no lists exist
		return &ShoppingData{
			ListName: "No shopping list",
			Items:    []ShoppingItem{},
		}, nil
	}

	// Get the first shopping list
	shoppinglist := listResp.Shoppinglists[0]

	// Convert to ShoppingData
	shoppingData := &ShoppingData{
		ListID:         shoppinglist.Id,
		ListName:       shoppinglist.Name,
		TotalItems:     int(shoppinglist.TotalItems),
		PurchasedItems: int(shoppinglist.PurchasedItems),
		PendingItems:   int(shoppinglist.PendingItems),
		Items:          make([]ShoppingItem, 0, len(shoppinglist.Items)),
	}

	// Convert items
	for _, item := range shoppinglist.Items {
		shoppingItem := ShoppingItem{
			ID:          item.Id,
			Quantity:    int(item.Quantity),
			IsPurchased: item.IsPurchased,
		}

		// Get item name and category from item_variant
		if item.ItemVariant != nil {
			if item.ItemVariant.Item != nil {
				shoppingItem.Name = item.ItemVariant.Item.Name

				// Get category name
				if item.ItemVariant.Item.Category != nil {
					shoppingItem.Category = item.ItemVariant.Item.Category.Name
				}
			}
		} else {
			shoppingItem.Name = "Unknown Item"
		}

		shoppingData.Items = append(shoppingData.Items, shoppingItem)
	}

	return shoppingData, nil
}

// fetchBlog gets the latest blog posts
func (h *MirrorHandler) fetchBlog(ctx context.Context) (*BlogData, error) {
	// List recent blog posts (first page, 5 posts)
	status := blogpb.PostStatus_POST_STATUS_PUBLISHED
	listReq := &blogpb.ListPostsRequest{
		Page:     1,
		PageSize: 5,
		Status:   &status,
	}

	listResp, err := h.blogClient.ListPosts(ctx, listReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list blog posts: %w", err)
	}

	blogData := &BlogData{
		TotalPosts: int(listResp.Total),
		Latest:     make([]BlogPost, 0, len(listResp.Posts)),
	}

	// Convert posts
	for _, post := range listResp.Posts {
		blogPost := BlogPost{
			ID:          post.Id,
			Title:       post.Title,
			Excerpt:     post.Excerpt,
			Author:      post.AuthorId,
			PublishedAt: post.PublishedAt.AsTime().Format(time.RFC3339),
		}
		blogData.Latest = append(blogData.Latest, blogPost)
	}

	return blogData, nil
}

// fetchServiceStatus checks the health of all services
func (h *MirrorHandler) fetchServiceStatus() *ServiceStatus {
	now := time.Now().Format(time.RFC3339)
	services := []ServiceInfo{}
	healthy := 0
	unhealthy := 0

	// Check weather service
	if h.serviceClients.WeatherConn != nil {
		status := "healthy"
		state := h.serviceClients.WeatherConn.GetState()
		if state.String() != "READY" {
			status = "unhealthy"
			unhealthy++
		} else {
			healthy++
		}
		services = append(services, ServiceInfo{
			Name:      "weather",
			Status:    status,
			LastCheck: now,
		})
	}

	// Check foodfolio service
	if h.serviceClients.FoodfolioConn != nil {
		status := "healthy"
		state := h.serviceClients.FoodfolioConn.GetState()
		if state.String() != "READY" {
			status = "unhealthy"
			unhealthy++
		} else {
			healthy++
		}
		services = append(services, ServiceInfo{
			Name:      "foodfolio",
			Status:    status,
			LastCheck: now,
		})
	}

	// Check blog service
	if h.serviceClients.BlogConn != nil {
		status := "healthy"
		state := h.serviceClients.BlogConn.GetState()
		if state.String() != "READY" {
			status = "unhealthy"
			unhealthy++
		} else {
			healthy++
		}
		services = append(services, ServiceInfo{
			Name:      "blog",
			Status:    status,
			LastCheck: now,
		})
	}

	return &ServiceStatus{
		Total:     len(services),
		Healthy:   healthy,
		Unhealthy: unhealthy,
		Services:  services,
	}
}
