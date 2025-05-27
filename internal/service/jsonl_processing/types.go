package jsonl_processing

import (
	"time"
)

type ReviewerInfo struct {
	CountryName           string `json:"countryName"`
	DisplayMemberName     string `json:"displayMemberName"`
	FlagName              string `json:"flagName"`
	ReviewGroupName       string `json:"reviewGroupName"`
	RoomTypeName          string `json:"roomTypeName"`
	CountryID             int    `json:"countryId"`
	LengthOfStay          int    `json:"lengthOfStay"`
	ReviewGroupID         int    `json:"reviewGroupId"`
	RoomTypeID            int    `json:"roomTypeId"`
	ReviewerReviewedCount int    `json:"reviewerReviewedCount"`
	IsExpertReviewer      bool   `json:"isExpertReviewer"`
	IsShowGlobalIcon      bool   `json:"isShowGlobalIcon"`
	IsShowReviewedCount   bool   `json:"isShowReviewedCount"`
}

type Comment struct {
	IsShowReviewResponse    bool         `json:"isShowReviewResponse"`
	HotelReviewID           int64        `json:"hotelReviewId"`
	ProviderID              int          `json:"providerId"`
	Rating                  float64      `json:"rating"`
	CheckInDateMonthAndYear string       `json:"checkInDateMonthAndYear"`
	EncryptedReviewData     string       `json:"encryptedReviewData"`
	FormattedRating         string       `json:"formattedRating"`
	FormattedReviewDate     string       `json:"formattedReviewDate"`
	RatingText              string       `json:"ratingText"`
	ResponderName           string       `json:"responderName"`
	ResponseDateText        string       `json:"responseDateText"`
	ResponseTranslateSource string       `json:"responseTranslateSource"`
	ReviewComments          string       `json:"reviewComments"`
	ReviewNegatives         string       `json:"reviewNegatives"`
	ReviewPositives         string       `json:"reviewPositives"`
	ReviewProviderLogo      string       `json:"reviewProviderLogo"`
	ReviewProviderText      string       `json:"reviewProviderText"`
	ReviewTitle             string       `json:"reviewTitle"`
	TranslateSource         string       `json:"translateSource"`
	TranslateTarget         string       `json:"translateTarget"`
	ReviewDate              time.Time    `json:"reviewDate"`
	ReviewerInfo            ReviewerInfo `json:"reviewerInfo"`
	OriginalTitle           string       `json:"originalTitle"`
	OriginalComment         string       `json:"originalComment"`
	FormattedResponseDate   string       `json:"formattedResponseDate"`
}

type Grades struct {
	Cleanliness           float64 `json:"Cleanliness"`
	Facilities            float64 `json:"Facilities"`
	Location              float64 `json:"Location"`
	RoomComfortAndQuality float64 `json:"Room comfort and quality"`
	Service               float64 `json:"Service"`
	ValueForMoney         float64 `json:"Value for money"`
}

type OverallByProvider struct {
	ProviderID   int     `json:"providerId"`
	Provider     string  `json:"provider"`
	OverallScore float64 `json:"overallScore"`
	ReviewCount  int     `json:"reviewCount"`
	Grades       Grades  `json:"grades"`
}

// HotelReviewData is representing the JSON data
type HotelReviewData struct {
	HotelID            int64               `json:"hotelId"`
	Platform           string              `json:"platform"`
	HotelName          string              `json:"hotelName"`
	Comment            Comment             `json:"comment"`
	OverallByProviders []OverallByProvider `json:"overallByProviders"`
}

// ProcessingResult holds the results of processing a JSON file
type ProcessingResult struct {
	TotalRecords   int
	SuccessRecords int
	ErrorRecords   int
	Errors         []ProcessingError
	Duration       time.Duration
}

type ProcessingError struct {
	LineNumber int
	Error      string
	RawData    string
}

type FileToProcess struct {
	Filename string
	S3Path   string
}

type FileResult struct {
	Filename string
	Result   *ProcessingResult
}
