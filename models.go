package main

import (
	"gorm.io/gorm"
	"time"
)

// User model
// @Description User account information for authentication and authorization
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	Username  string         `json:"username" gorm:"unique;not null" example:"admin" description:"Unique username for login"`
	Password  string         `json:"-" gorm:"not null" description:"Hashed password (not exposed in API)"`
	CreatedAt time.Time      `json:"created_at" example:"2024-01-01T00:00:00Z" description:"When the user was created"`
	UpdatedAt time.Time      `json:"updated_at" example:"2024-01-01T00:00:00Z" description:"When the user was last updated"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index" description:"Soft delete timestamp (not exposed in API)"`
} // @User

// Hedgehog model
// @Description Information about a hedgehog in the rescue center
type Hedgehog struct {
	ID            uint           `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	Name          string         `json:"name" gorm:"not null" example:"Spillo" description:"Name of the hedgehog"`
	Description   string         `json:"description" example:"Riccio trovato nel giardino" description:"Additional information about the hedgehog"`
	ArrivalDate   time.Time      `json:"arrival_date" example:"2024-01-15T10:30:00Z" description:"When the hedgehog arrived at the center" format:"date-time"`
	Status        string         `json:"status" gorm:"default:'in_care'" example:"in_care" enums:"in_care,recovered,deceased,released" description:"Current status of the hedgehog"`
	ReleaseDate   *time.Time     `json:"release_date,omitempty" example:"2024-07-28T10:30:00Z" description:"When the hedgehog was or will be released" format:"date-time"`
	AreaID        *uint          `json:"area_id" example:"1" description:"ID of the area where the hedgehog is located"`
	Area          *Area          `json:"area,omitempty" gorm:"foreignKey:AreaID" description:"Area where the hedgehog is located"`
	Therapies     []Therapy      `json:"therapies,omitempty" description:"Treatments and therapies for the hedgehog"`
	WeightRecords []WeightRecord `json:"weight_records,omitempty" description:"Weight history records"`
	CreatedAt     time.Time      `json:"created_at" example:"2024-01-15T10:30:00Z" description:"When the record was created" format:"date-time"`
	UpdatedAt     time.Time      `json:"updated_at" example:"2024-01-15T10:30:00Z" description:"When the record was last updated" format:"date-time"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index" description:"Soft delete timestamp (not exposed in API)"`
} // @Hedgehog

// Room model
// @Description A physical room in the rescue center that contains areas for hedgehogs
type Room struct {
	ID          uint           `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	Name        string         `json:"name" gorm:"not null" example:"Sala Principale" description:"Name of the room"`
	Description string         `json:"description" example:"Stanza principale con gabbie e aree di recupero" description:"Additional details about the room"`
	Width       float64        `json:"width" gorm:"default:100" example:"500" description:"Width of the room in cm" minimum:"1"`
	Height      float64        `json:"height" gorm:"default:100" example:"400" description:"Height of the room in cm" minimum:"1"`
	Areas       []Area         `json:"areas,omitempty" description:"Areas contained within this room"`
	CreatedAt   time.Time      `json:"created_at" example:"2024-01-15T10:30:00Z" description:"When the record was created" format:"date-time"`
	UpdatedAt   time.Time      `json:"updated_at" example:"2024-01-15T10:30:00Z" description:"When the record was last updated" format:"date-time"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index" description:"Soft delete timestamp (not exposed in API)"`
} // @Room

// Area model
// @Description A specific area within a room where hedgehogs are housed
type Area struct {
	ID          uint           `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	Name        string         `json:"name" gorm:"not null" example:"Gabbia 1" description:"Name of the area"`
	RoomID      uint           `json:"room_id" example:"1" description:"ID of the room containing this area"`
	Room        Room           `json:"room,omitempty" gorm:"foreignKey:RoomID" description:"Room containing this area"`
	X           float64        `json:"x" example:"10" description:"X coordinate position within the room in cm" minimum:"0"`
	Y           float64        `json:"y" example:"20" description:"Y coordinate position within the room in cm" minimum:"0"`
	Width       float64        `json:"width" example:"100" description:"Width of the area in cm" minimum:"1"`
	Height      float64        `json:"height" example:"80" description:"Height of the area in cm" minimum:"1"`
	MaxCapacity int            `json:"max_capacity" gorm:"default:1" example:"2" description:"Maximum number of hedgehogs this area can house" minimum:"1"`
	Hedgehogs   []Hedgehog     `json:"hedgehogs,omitempty" description:"Hedgehogs currently housed in this area"`
	CreatedAt   time.Time      `json:"created_at" example:"2024-01-15T10:30:00Z" description:"When the record was created" format:"date-time"`
	UpdatedAt   time.Time      `json:"updated_at" example:"2024-01-15T10:30:00Z" description:"When the record was last updated" format:"date-time"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index" description:"Soft delete timestamp (not exposed in API)"`
} // @Area

// Therapy model
// @Description A medical treatment or therapy administered to a hedgehog
type Therapy struct {
	ID          uint           `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	HedgehogID  uint           `json:"hedgehog_id" example:"1" description:"ID of the hedgehog receiving this therapy"`
	Name        string         `json:"name" gorm:"not null" example:"Antibiotico" description:"Name of the therapy or treatment"`
	Description string         `json:"description" example:"Somministrazione di antibiotico per infezione" description:"Detailed description of the therapy"`
	StartDate   time.Time      `json:"start_date" example:"2024-01-15T10:30:00Z" description:"When the therapy started" format:"date-time"`
	EndDate     *time.Time     `json:"end_date" example:"2024-01-30T10:30:00Z" description:"When the therapy is scheduled to end" format:"date-time"`
	Status      string         `json:"status" gorm:"default:'active'" example:"active" enums:"active,completed,suspended" description:"Current status of the therapy"`
	CreatedAt   time.Time      `json:"created_at" example:"2024-01-15T10:30:00Z" description:"When the record was created" format:"date-time"`
	UpdatedAt   time.Time      `json:"updated_at" example:"2024-01-15T10:30:00Z" description:"When the record was last updated" format:"date-time"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index" description:"Soft delete timestamp (not exposed in API)"`
} // @Therapy

// WeightRecord model
// @Description A record of a hedgehog's weight measurement
type WeightRecord struct {
	ID         uint           `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	HedgehogID uint           `json:"hedgehog_id" example:"1" description:"ID of the hedgehog this weight record belongs to"`
	Weight     float64        `json:"weight" example:"450.5" description:"Weight in grams" minimum:"1"`
	Date       time.Time      `json:"date" example:"2024-01-15T10:30:00Z" description:"When the weight was measured" format:"date-time"`
	Notes      string         `json:"notes" example:"Peso stabile" description:"Additional notes about the weight measurement"`
	CreatedAt  time.Time      `json:"created_at" example:"2024-01-15T10:30:00Z" description:"When the record was created" format:"date-time"`
	UpdatedAt  time.Time      `json:"updated_at" example:"2024-01-15T10:30:00Z" description:"When the record was last updated" format:"date-time"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index" description:"Soft delete timestamp (not exposed in API)"`
} // @WeightRecord

// Notification types
// @Description Type of notification that can be generated by the system
type NotificationType string // @NotificationType

// @enum therapy_expired therapy_expiring weight_drop weight_stagnation no_weighing hedgehog_recovered system_alert
const (
	NotificationTherapyExpired    NotificationType = "therapy_expired"    // When a therapy has passed its end date
	NotificationTherapyExpiring   NotificationType = "therapy_expiring"   // When a therapy is about to expire
	NotificationWeightDrop        NotificationType = "weight_drop"        // When a hedgehog has lost significant weight
	NotificationWeightStagnation  NotificationType = "weight_stagnation"  // When a hedgehog's weight hasn't changed for a period
	NotificationNoWeighing        NotificationType = "no_weighing"        // When a hedgehog hasn't been weighed recently
	NotificationHedgehogRecovered NotificationType = "hedgehog_recovered" // When a hedgehog has recovered and can be released
	NotificationSystemAlert       NotificationType = "system_alert"       // System-level alerts
)

// @Description Priority level for notifications
type NotificationPriority string // @NotificationPriority

// @enum low medium high critical
const (
	PriorityLow      NotificationPriority = "low"      // Informational notifications
	PriorityMedium   NotificationPriority = "medium"   // Standard notifications
	PriorityHigh     NotificationPriority = "high"     // Important notifications
	PriorityCritical NotificationPriority = "critical" // Urgent notifications requiring immediate attention
)

// Notification model
// @Description A notification generated by the system to alert users about important events
type Notification struct {
	ID          uint                 `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	Type        NotificationType     `json:"type" gorm:"not null" example:"therapy_expired" description:"Type of notification"`
	Priority    NotificationPriority `json:"priority" gorm:"default:'medium'" example:"high" description:"Priority level of the notification"`
	Title       string               `json:"title" gorm:"not null" example:"Terapia Scaduta" description:"Short title of the notification"`
	Message     string               `json:"message" gorm:"not null" example:"La terapia 'Antibiotico' per Spillo è scaduta il 15/01/2024" description:"Detailed message of the notification"`
	HedgehogID  *uint                `json:"hedgehog_id" example:"1" description:"ID of the related hedgehog, if applicable"`
	Hedgehog    *Hedgehog            `json:"hedgehog,omitempty" gorm:"foreignKey:HedgehogID" description:"Related hedgehog information"`
	TherapyID   *uint                `json:"therapy_id" example:"1" description:"ID of the related therapy, if applicable"`
	Therapy     *Therapy             `json:"therapy,omitempty" gorm:"foreignKey:TherapyID" description:"Related therapy information"`
	Data        string               `json:"data" example:"{\"days_overdue\": 5}" description:"Additional JSON data related to the notification"`
	Read        bool                 `json:"read" gorm:"default:false" example:"false" description:"Whether the notification has been read"`
	Dismissed   bool                 `json:"dismissed" gorm:"default:false" example:"false" description:"Whether the notification has been dismissed"`
	CreatedAt   time.Time            `json:"created_at" example:"2024-01-15T10:30:00Z" description:"When the notification was created" format:"date-time"`
	ExpiresAt   *time.Time           `json:"expires_at" example:"2024-02-15T10:30:00Z" description:"When the notification expires" format:"date-time"`
	ActionURL   string               `json:"action_url" example:"/hedgehogs/1" description:"URL for the action button"`
	ActionLabel string               `json:"action_label" example:"Gestisci Terapia" description:"Label for the action button"`
} // @Notification

// NotificationSettings model
// @Description Configuration settings for the notification system
type NotificationSettings struct {
	ID                        uint      `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	TherapyExpiredEnabled     bool      `json:"therapy_expired_enabled" gorm:"default:true" example:"true" description:"Whether to enable notifications for expired therapies"`
	TherapyExpiringDays       int       `json:"therapy_expiring_days" gorm:"default:3" example:"3" description:"Days before therapy expiration to send notification" minimum:"1"`
	WeightDropThreshold       float64   `json:"weight_drop_threshold" gorm:"default:50" example:"50" description:"Threshold in grams for weight drop notifications" minimum:"1"`
	WeightDropDays            int       `json:"weight_drop_days" gorm:"default:7" example:"7" description:"Period in days to check for weight drops" minimum:"1"`
	WeightStagnationDays      int       `json:"weight_stagnation_days" gorm:"default:14" example:"14" description:"Days of weight stagnation before notification" minimum:"1"`
	NoWeighingDays            int       `json:"no_weighing_days" gorm:"default:7" example:"7" description:"Days without weighing before notification" minimum:"1"`
	EmailNotificationsEnabled bool      `json:"email_notifications_enabled" gorm:"default:false" example:"false" description:"Whether to send notifications via email"`
	EmailAddress              string    `json:"email_address" example:"admin@laninna.org" description:"Email address for notifications"`
	WebhookURL                string    `json:"webhook_url" example:"https://hooks.slack.com/services/xxx" description:"Webhook URL for external notifications"`
	CreatedAt                 time.Time `json:"created_at" example:"2024-01-15T10:30:00Z" description:"When the settings were created" format:"date-time"`
	UpdatedAt                 time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z" description:"When the settings were last updated" format:"date-time"`
} // @NotificationSettings
