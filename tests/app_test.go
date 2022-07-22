package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/asalkeld/test-app/common"
)

const (
	pollingInterval     = 5 * time.Second
	pollingTimeout      = 90 * time.Second
	pollingTimeoutAPIUp = 5 * time.Minute
)

var (
	localRun     = true
	baseUrl      = "http://localhost:9001/apis/nitric-testr"
	topicBaseURL = "http://localhost:9001/topic"
	storeUrl     = baseUrl + "/store"
	historyUrl   = baseUrl + "/history"
	sendUrl      = baseUrl + "/send"
	safeUrl      = baseUrl + "/safe"
)

func init() {
	if os.Getenv("BASE_URL") != "" {
		localRun = false
		baseUrl = os.Getenv("BASE_URL")
		storeUrl = baseUrl + "/store"
		historyUrl = baseUrl + "/history"
		sendUrl = baseUrl + "/send"
		safeUrl = baseUrl + "/safe"
		fmt.Println(baseUrl)
	}
}

func send(method, url string, data any, headers map[string]string) ([]byte, int, error) {
	fmt.Printf("%s %s\n", method, url)

	if headers == nil {
		headers = map[string]string{}
	}

	var r io.Reader

	if method == http.MethodPost || method == http.MethodPut {
		if s, ok := data.(string); ok {
			r = strings.NewReader(s)

			headers["Content-Type"] = "text/plain; charset=utf-8"
		} else {
			b, err := json.Marshal(data)
			if err != nil {
				return nil, http.StatusBadRequest, err
			}

			headers["Content-Type"] = http.DetectContentType(b)
			r = strings.NewReader(string(b))
		}
	}

	req, err := http.NewRequest(method, url, r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	for k, v := range headers {
		req.Header[k] = []string{v}
	}

	cli := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := cli.Do(req)
	if err != nil {
		if resp != nil {
			return nil, resp.StatusCode, errors.WithMessagef(err, "send %s:%s", method, url)
		}

		return nil, 500, errors.WithMessagef(err, "send %s:%s", method, url)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, errors.WithMessagef(err, "send %s:%s", method, url)
	}

	return body, resp.StatusCode, errors.WithMessagef(err, "send %s:%s", method, url)
}

func listStore() ([]common.Store, error) {
	r, _, err := send(http.MethodGet, storeUrl, nil, nil)
	if err != nil {
		return nil, err
	}

	s := []common.Store{}
	err = json.Unmarshal(r, &s)

	return s, err
}

func history() ([]common.Fact, error) {
	r, _, err := send(http.MethodGet, historyUrl, nil, nil)
	if err != nil {
		return nil, err
	}

	s := []common.Fact{}
	err = json.Unmarshal(r, &s)

	return s, err
}

func deleteOne(base, id string) error {
	r, code, err := send(http.MethodDelete, base+"/"+id, nil, nil)

	if code != http.StatusNoContent {
		return fmt.Errorf("DELETE %s/%s failed %d %s", base, id, code, r)
	}

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
		fmt.Println("history ", err)
		return err
	}

	for _, s := range ss {
		err = deleteOne(historyUrl, s.ID)
		if err != nil {
			fmt.Println("deleteOne ", err)

			return err
		}
	}

	return nil
}

func createStore(s *common.Store) error {
	b, code, err := send(http.MethodPost, storeUrl, s, nil)
	if err != nil {
		return err
	}

	if code != http.StatusOK {
		return fmt.Errorf("Post Store %v", b)
	}

	return nil
}

func sendMsg(m *common.Message) error {
	b, code, err := send(http.MethodPost, sendUrl, m, nil)
	if err != nil {
		return err
	}

	if code != http.StatusOK {
		return fmt.Errorf("Post Msg %v", b)
	}

	return nil
}

func runSchedule(name string) {
	if localRun {
		_, _, _ = send(http.MethodPost, topicBaseURL+"/"+name, "", map[string]string{})
	}
}

func apiIsUp() error {
	_, err := listStore()
	if err != nil {
		fmt.Println("store API not up: " + err.Error())
		return err
	}

	_, err = history()
	if err != nil {
		fmt.Println("history API not up: " + err.Error())
		return err
	}

	time.Sleep(2 * time.Second)

	return nil
}

func waitForFactID(testID, action string) func() error {
	return func() error {
		runSchedule("job")

		hist, err := history()
		if err != nil {
			return err
		}

		fmt.Println("searching for ID=", testID)

		for _, f := range hist {
			if f.Action == action {
				fact := common.Fact{}

				err = json.Unmarshal([]byte(f.Data), &fact)
				if err != nil {
					return err
				}

				fmt.Println(fact)

				if fact.ID == testID {
					return nil
				}
			}
		}

		return errors.New("not found")
	}
}

func TestAppSafe(t *testing.T) {
	g := NewGomegaWithT(t)

	// esp. in CI, wait for the API to come up.
	g.Eventually(apiIsUp).
		WithPolling(pollingInterval).
		WithTimeout(pollingTimeoutAPIUp).
		ShouldNot(HaveOccurred())

	test := "hello this is a test"

	_, code, err := send(http.MethodPost, safeUrl, test, map[string]string{})
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(code).Should(Equal(200))

	r, code, err := send(http.MethodGet, safeUrl, nil, map[string]string{})
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(code).Should(Equal(200))
	g.Expect(r).Should(Equal([]byte(test)))
}

func TestAppStore(t *testing.T) {
	g := NewGomegaWithT(t)

	// esp. in CI, wait for the API to come up.
	g.Eventually(apiIsUp).
		WithPolling(pollingInterval).
		WithTimeout(pollingTimeoutAPIUp).
		ShouldNot(HaveOccurred())

	err := deleteStore()
	g.Expect(err).ShouldNot(HaveOccurred())

	s, err := listStore()
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(len(s)).To(Equal(0))

	err = deleteHistory()
	g.Expect(err).ShouldNot(HaveOccurred())

	err = createStore(&common.Store{ID: "angus", Data: "test34"})
	g.Expect(err).ShouldNot(HaveOccurred())
	err = createStore(&common.Store{ID: "tim", Data: "test98"})
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

func TestAppTopic(t *testing.T) {
	g := NewGomegaWithT(t)

	// esp. in CI, wait for the API to come up.
	g.Eventually(apiIsUp).
		WithPolling(pollingInterval).
		WithTimeout(pollingTimeoutAPIUp).
		ShouldNot(HaveOccurred())

	err := deleteHistory()
	g.Expect(err).ShouldNot(HaveOccurred())

	testID := uuid.New().String()

	err = sendMsg(&common.Message{
		MessageType: "topic",
		ID:          testID,
		PayloadType: "None",
		Payload:     testID,
	})
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(waitForFactID(testID, "received event")).
		WithPolling(pollingInterval).
		WithTimeout(pollingTimeout).
		ShouldNot(HaveOccurred())
}

func TestAppQueue(t *testing.T) {
	g := NewGomegaWithT(t)

	// esp. in CI, wait for the API to come up.
	g.Eventually(apiIsUp).
		WithPolling(pollingInterval).
		WithTimeout(pollingTimeoutAPIUp).
		ShouldNot(HaveOccurred())

	err := deleteHistory()
	g.Expect(err).ShouldNot(HaveOccurred())

	testID := uuid.New().String()
	err = sendMsg(&common.Message{
		MessageType: "queue",
		ID:          testID,
		PayloadType: "None",
		Payload:     testID,
	})
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(waitForFactID(testID, "task complete")).
		WithPolling(pollingInterval).
		WithTimeout(pollingTimeout).
		ShouldNot(HaveOccurred())
}
