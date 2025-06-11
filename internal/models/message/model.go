package message

type Message struct {
	CohortName string `json:"cohort_name"`
	UserId     string `json:"user_id"`
	Delay      int    `json:"delay"`
	Message    string `json:"message"`
	Image      string `json:"image"`
}

func NewMessage() Message {
	return Message{
		CohortName: "test_group",
		UserId:     "",
		Delay:      1,
		Message:    "this is sample text",
		// Image:      "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRK17ZrDQgMv0jVqJlMIgvYg6_6XXDG6WS9Iq6eRjMiGMPdRUaajLOFgIAUOtGIL9Y2uQ0&usqp=CAU",
		Image: "https://previews.123rf.com/images/paylessimages/paylessimages1412/paylessimages141200527/43857336-beanbags-and-palm.jpg",
	}
}
