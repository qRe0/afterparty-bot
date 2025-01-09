package models

type TicketResponse struct {
	Id                string `json:"id"`
	Name              string `json:"full_name"`
	TicketType        string `json:"ticket_type"`
	PassedControlZone bool   `json:"passed_control_zone]"`
}

type ClientData struct {
	FIO          string `json:"fio"`
	TicketType   string `json:"ticket_type"`
	Price        int    `json:"price"`
	RepostExists bool   `json:"repost_exists"`
}
