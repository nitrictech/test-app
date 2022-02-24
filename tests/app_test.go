package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/asalkeld/test-app/common"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	baseUrl    = "http://localhost:9001/apis/nitric-testr"
	storeUrl   = baseUrl + "/store"
	historyUrl = baseUrl + "/history"
	sendUrl    = baseUrl + "/send"
)

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

var _ = Describe("App", func() {
	Context("Store", func() {

		When("The store is empty", func() {
			err := deleteStore()
			Expect(err).ShouldNot(HaveOccurred())
			err = deleteHistory()
			Expect(err).ShouldNot(HaveOccurred())

			It("list should be empty", func() {
				s, err := listStore()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(s)).To(Equal(0))
			})

			It("should be able to create and get,list", func() {
				err := createStore(common.Store{ID: "angus", Data: "test34"})
				Expect(err).ShouldNot(HaveOccurred())
				err = createStore(common.Store{ID: "tim", Data: "test98"})
				Expect(err).ShouldNot(HaveOccurred())

				s, err := listStore()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(s)).To(Equal(2))

				err = deleteOne(storeUrl, "angus")
				Expect(err).ShouldNot(HaveOccurred())
				err = deleteOne(storeUrl, "tim")
				Expect(err).ShouldNot(HaveOccurred())

				s, err = listStore()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(s)).To(Equal(0))
			})
		})
	})

	Context("Topic", func() {
		When("Topic messages are sent and received", func() {
			err := deleteHistory()
			Expect(err).ShouldNot(HaveOccurred())

			testID := uuid.New().String()
			err = sendMsg(common.Message{
				MessageType: "topic",
				ID:          testID,
				PayloadType: "None",
				Payload:     "",
			})
			Expect(err).ShouldNot(HaveOccurred())

			It("the message is received", func() {
				hist, err := history()
				Expect(err).ShouldNot(HaveOccurred())
				found := false
				for _, f := range hist {
					if f.Action == "received event" {
						fact := common.Fact{}
						err = json.Unmarshal([]byte(f.Data), &fact)
						Expect(err).ShouldNot(HaveOccurred())
						if fact.ID == testID {
							found = true
							break
						}
					}
				}
				Expect(found).To(BeTrue())
			})
		})
	})

	Context("Queue", func() {
		When("Queue messages are sent and received", func() {
			err := deleteHistory()
			Expect(err).ShouldNot(HaveOccurred())

			testID := uuid.New().String()
			err = sendMsg(common.Message{
				MessageType: "queue",
				ID:          testID,
				PayloadType: "None",
				Payload:     testID,
			})
			Expect(err).ShouldNot(HaveOccurred())

			It("the message is received", func() {
				found := false
				startTime := time.Now()
				for {
					hist, err := history()
					Expect(err).ShouldNot(HaveOccurred())
					for _, f := range hist {
						fmt.Println(f)
						if f.Action == "task complete" {
							fact := common.Fact{}
							err = json.Unmarshal([]byte(f.Data), &fact)
							Expect(err).ShouldNot(HaveOccurred())
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
					if time.Since(startTime).Seconds() > 70 {
						break
					}
					fmt.Println("waiting some more...")
					time.Sleep(10 * time.Second)
				}
				Expect(found).To(BeTrue())
			})
		})
	})
})
