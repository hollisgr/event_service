package event

type Event struct {
	DeviceCarrier   string         `json:"device_carrier"`
	DeviceFamily    string         `json:"device_family"`
	DeviceId        string         `json:"device_id" validate:"required"`
	DeviceType      string         `json:"device_type"`
	DisplayName     string         `json:"display_name" validate:"required"`
	DMA             string         `json:"dma"`
	EventId         int            `json:"event_id" validate:"gte=0"`
	EventProperties map[string]any `json:"event_properties"`
	// EventProperties struct {
	// 	AuthType string `json:"auth_type" validate:"required"`
	// } `json:"event_properties"`
	EventTime string `json:"event_time" validate:"required"`
	EventType string `json:"event_type" validate:"required"`
	UserId    string `json:"user_id" validate:"required"`
	// UserProperties struct {
	// 	Balance      int    `json:"balance" validate:"gte=0"`
	// 	DeviceType   string `json:"deviceType" validate:"required"`
	// 	IsOnline     bool   `json:"is_online" validate:"required"`
	// 	LanguageCode string `json:"language_code" validate:"required"`
	// 	Rating       string `json:"rating" validate:"required"`
	// 	Registration string `json:"registration" validate:"required"`
	// 	Source       string `json:"source" validate:"required"`
	// 	Status       string `json:"status" validate:"required"`
	// 	Successful   string `json:"successful" validate:"required"`
	// 	Username     string `json:"username" validate:"required"`
	// } `json:"user_properties"`
	UserProperties map[string]any `json:"user_properties"`
	UUID           string         `json:"uuid" validate:"required"`
	VersionName    string         `json:"version_name" validate:"required"`
	Status         string         `json:"status"`
}

type EventWithPipeline struct {
	EventId    int
	PipeLineId int
}

func NewEvent() Event {
	return Event{
		DeviceCarrier: "null",
		DeviceFamily:  "null",
		DeviceId:      "null",
		DeviceType:    "null",
		DisplayName:   "null",
		DMA:           "null",
		EventId:       0,
		Status:        "new",
	}
}

func EmptyEvent() Event {
	return Event{
		DeviceCarrier:   "null",
		DeviceFamily:    "null",
		DeviceId:        "null",
		DeviceType:      "null",
		DisplayName:     "null",
		DMA:             "null",
		EventId:         0,
		EventProperties: map[string]any{},
		EventTime:       "null",
		EventType:       "null",
		UserId:          "null",
		UserProperties:  map[string]any{},
		UUID:            "null",
		VersionName:     "null",
		Status:          "null",
	}
}
