package models

type TicketResponse struct {
	Id         string `json:"id"`
	Name       string `json:"full_name"`
	TicketType string `json:"ticket_type"`
}
