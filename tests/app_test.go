package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/asalkeld/test-app/common"
	"github.com/google/uuid"
)

var (
	localRun   = true
	baseUrl    = "http://localhost:9001/apis/nitric-testr"
	storeUrl   = baseUrl + "/store"
	historyUrl = baseUrl + "/history"
	sendUrl    = baseUrl + "/send"
)

func init() {
	if os.Getenv("BASE_URL") != "" {
		localRun = false
		baseUrl = os.Getenv("BASE_URL")
		storeUrl = baseUrl + "/store"
		historyUrl = baseUrl + "/history"
		sendUrl = baseUrl + "/send"
		fmt.Println(baseUrl)
	}
}

func listStore() ([]common.Store, error) {
	resp, err := http.Get(storeUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	s := []common.Store{}
	err = json.Unmarshal(body, &s)
	return s, err
}

func history() ([]common.Fact, error) {
	resp, err := http.Get(historyUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	s := []common.Fact{}
	err = json.Unmarshal(body, &s)
	return s, err
}

func deleteOne(base, id string) error {
	fmt.Println("deleting ", id)
	req, err := http.NewRequest("DELETE", base+"/"+id, nil)
	if err != nil {
		return err
	}
	_, err = http.DefaultClient.Do(req)
	return err
}

func deleteStore() error {
	ss, err := listStore()
	if err != nil {
		return err
	}
	for _, s := range ss {
		err = deleteOne(storeUrl, s.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteHistory() error {
	ss, err := history()
	if err != nil {
		return err
	}
	for _, s := range ss {
		err = deleteOne(historyUrl, s.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func createStore(s common.Store) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	r, err := http.Post(storeUrl, http.DetectContentType(b), bytes.NewReader(b))
	if err != nil {
		return err
	}

	if r.StatusCode != 200 {
		return fmt.Errorf("Post Store %v", r)
	}
	return nil
}

func sendMsg(m common.Message) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	r, err := http.Post(sendUrl, http.DetectContentType(b), bytes.NewReader(b))
	if err != nil {
		return err
	}

	if r.StatusCode != 200 {
		return fmt.Errorf("Post Message %v", r)
	}
	return nil
}

func TestAppStore(t *testing.T) {
	g := NewGomegaWithT(t)

	err := deleteStore()
	g.Expect(err).ShouldNot(HaveOccurred())
	err = deleteHistory()
	g.Expect(err).ShouldNot(HaveOccurred())

	s, err := listStore()
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(len(s)).To(Equal(0))

	err = createStore(common.Store{ID: "angus", Data: "test34"})
	g.Expect(err).ShouldNot(HaveOccurred())
	err = createStore(common.Store{ID: "tim", Data: "test98"})
	g.Expect(err).ShouldNot(HaveOccurred())

	s, err = listStore()
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(len(s)).To(Equal(2))

	err = deleteOne(storeUrl, "angus")
	g.Expect(err).ShouldNot(HaveOccurred())
	err = deleteOne(storeUrl, "tim")
	g.Expect(err).ShouldNot(HaveOccurred())

	s, err = listStore()
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(len(s)).To(Equal(0))
}

func waitForFactID(testID, action string, waitSecs int) (bool, error) {
	found := false
	startTime := time.Now()
	for {
		hist, err := history()
		if err != nil {
			return false, err
		}
		fmt.Println("searching for ID=", testID)
		for _, f := range hist {
			if f.Action == action {
				fact := common.Fact{}
				err = json.Unmarshal([]byte(f.Data), &fact)
				if err != nil {
					return false, err
				}
				fmt.Println(fact)
				if fact.ID == testID {
					found = true
					break
				}
			}
		}
		if found {
			break
		}
		if time.Since(startTime).Seconds() > float64(waitSecs) {
			break
		}
		fmt.Println("waiting some more...")
		time.Sleep(10 * time.Second)
	}
	return found, nil
}

func TestAppTopic(t *testing.T) {
	g := NewGomegaWithT(t)

	err := deleteHistory()
	g.Expect(err).ShouldNot(HaveOccurred())

	testID := uuid.New().String()
	err = sendMsg(common.Message{
		MessageType: "topic",
		ID:          testID,
		PayloadType: "None",
		Payload:     "",
	})
	g.Expect(err).ShouldNot(HaveOccurred())

	found, err := waitForFactID(testID, "received event", 30)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(found).To(BeTrue())

}

func TestAppQueue(t *testing.T) {
	g := NewGomegaWithT(t)

	err := deleteHistory()
	g.Expect(err).ShouldNot(HaveOccurred())

	testID := uuid.New().String()
	err = sendMsg(common.Message{
		MessageType: "queue",
		ID:          testID,
		PayloadType: "None",
		Payload:     testID,
	})
	g.Expect(err).ShouldNot(HaveOccurred())

	found, err := waitForFactID(testID, "task complete", 90)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(found).To(BeTrue())
}
