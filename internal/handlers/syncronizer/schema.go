package syncronizer

import (
	"encoding/json"
	"net/http"

	"github.com/tommyhedley/fiberytsheets/internal/utils"
)

type Relation struct {
	Cardinality   string `json:"cardinality"`
	Name          string `json:"name"`
	TargetName    string `json:"targetName"`
	TargetType    string `json:"targetType"`
	TargetFieldID string `json:"targetFieldId"`
}

type Field struct {
	Ignore      bool      `json:"ignore,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ReadOnly    bool      `json:"readonly,omitempty"`
	Type        string    `json:"type,omitempty"`
	SubType     string    `json:"subType,omitempty"`
	Relation    *Relation `json:"relation,omitempty"`
}

func Schema(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Types   []string   `json:"types"`
		Filter  SyncFilter `json:"filter"`
		Account struct {
			Token string `json:"token"`
		} `json:"account"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	users := map[string]Field{
		"id": {
			Name: "Id",
			Type: "id",
		},
		"timeId": {
			Name:   "Time ID",
			Type:   "text",
			Ignore: true,
		},
		"name": {
			Name: "Name",
			Type: "text",
		},
		"firstName": {
			Name: "First Name",
			Type: "text",
		},
		"lastName": {
			Name: "Last Name",
			Type: "text",
		},
		"active": {
			Name:     "Active",
			SubType:  "boolean",
			ReadOnly: false,
		},
		"email": {
			Name:    "Email",
			SubType: "email",
		},
		"lastActive": {
			Name: "Last Active",
			Type: "date",
		},
		"groupId": {
			Name: "Group ID",
			Type: "text",
			Relation: &Relation{
				Cardinality:   "many-to-one",
				Name:          "Group",
				TargetName:    "Users",
				TargetType:    "group",
				TargetFieldID: "timeId",
			},
		},
	}

	group := map[string]Field{
		"id": {
			Name: "Id",
			Type: "id",
		},
		"timeId": {
			Name:   "Time ID",
			Type:   "text",
			Ignore: true,
		},
		"name": {
			Name: "Name",
			Type: "text",
		},
	}

	allType := map[string]map[string]Field{
		"users": users,
		"group": group,
	}

	returnType := map[string]map[string]Field{}

	for name, fields := range allType {
		for _, t := range params.Types {
			if name == t {
				returnType[name] = fields
			}
		}
	}

	utils.RespondWithJSON(w, http.StatusOK, returnType)
}
