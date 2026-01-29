package models

type DashboardResponse struct {
	Stats   DashboardStats   `json:"stats"`
	Growth  DashboardGrowth  `json:"growth"`
	Metrics DashboardMetrics `json:"metrics"`
}

type DashboardStats struct {
	TotalProductsListed     int64 `json:"total_product_listed"`
	ActiveProducts          int64 `json:"active_products"`
	ClosedSoldProducts      int64 `json:"closed_sold_products"`
	FlaggedReportedProducts int64 `json:"flagged_reported_products"`
	TotalRegisteredSellers  int64 `json:"total_registered_sellers"`
}
type DashboardGrowth struct {
	TotalProductsListed     float64 `json:"total_product_listed"`
	ActiveProducts          float64 `json:"active_products"`
	ClosedSoldProducts      float64 `json:"closed_sold_products"`
	FlaggedReportedProducts float64 `json:"flagged_reported_products"`
	TotalRegisteredSellers  float64 `json:"total_registered_sellers"`
}

type MonthlyMetric struct {
	Month string `json:"month"`
	Value int64  `json:"value"`
}

type DashboardMetrics struct {
	CustomerSignUpMetrics []MonthlyMetric `json:"customer_signup_metrics"`
	TotalSoldProducts     []MonthlyMetric `json:"total_sold_products"`
}

type UserDashboardResponse struct {
	Stats  UserDashboardStats  `json:"user_stats"`
	Growth UserDashboardGrowth `json:"growth"`
}

type UserDashboardStats struct {
	TotalSellers     int64 `json:"total_sellers"`
	ActiveSellers    int64 `json:"active_sellers"`
	SuspendedUsers   int64 `json:"suspended_users"`
	DeactivatedUsers int64 `json:"deactivated_users"`
}

type UserDashboardGrowth struct {
	TotalSellers     float64 `json:"total_sellers"`
	ActiveSellers    float64 `json:"active_sellers"`
	SuspendedUsers   float64 `json:"suspended_users"`
	DeactivatedUsers float64 `json:"deactivated_users"`
}

type UserDashboardUsers struct {
	User           PublicUser `json:"user"`
	ListedProducts int64      `json:"listed_products"`
	SoldProducts   int64      `json:"sold_products"`
}

type UserProductStats struct {
	TotalProductsListed int64 `json:"total_products_listed"`
	ActiveProducts      int64 `json:"active_products"`
	SoldProducts        int64 `json:"sold_products"`
	FlaggedProducts     int64 `json:"flagged_products"`
}
