// [START snippet]

package common

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/nitrictech/go-sdk/api/documents"
)

type Fact struct {
	ID      string `json:"id"`
	Occured string `json:"occured"`
	Source  string `json:"source"`
	Action  string `json:"action"`
	Data    string `json:"data"`
}

func RecordFact(col documents.CollectionRef, source, action, data string) {
	fact := &Fact{
		ID:      uuid.New().String(),
		Occured: time.Now().Format(time.RFC3339),
		Source:  source,
		Action:  action,
		Data:    data,
	}
	factMap := make(map[string]interface{})
	err := mapstructure.Decode(fact, &factMap)
	if err != nil {
		fmt.Println("error decoding fact document")
	}

	if err := col.Doc(fact.ID).Set(factMap); err != nil {
		fmt.Println("error writing fact to histroy document")
	}
}

// [END snippet]
