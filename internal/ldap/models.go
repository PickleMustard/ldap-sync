package ldap

import (
	"time"
)

type User struct {
	ID           string      `json:"id"`
	CreationDate time.Time   `json:"creationDate"`
	DisplayName  string      `json:"displayName"`
	FirstName    string      `json:"firstName"`
	LastName     string      `json:"lastName"`
	UUID         string      `json:"entryuuid"`
	Email        string      `json:"email"`
	Groups       []GroupRef  `json:"groups"`
	Attributes   []Attribute `json:"attributes"`
}

type GroupRef struct {
	ID   int    `json:"id"`
	Name string `json:"displayName"`
}

type Attribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type UsersResponse struct {
	Users []User
}
