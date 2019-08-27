package thttp

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"torn/model"
)

const UserSelections = "bars,battlestats,jobpoints,personalstats,refills,basic"

type TornClient struct {
	Client *http.Client
}

func NewTornClient() *TornClient {
	transport := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 10,
	}
	return &TornClient{client}
}

type TornErrorResponse struct {
	Error TornError `json:"error,omitempty"`
}

type TornError struct {
	Code  uint   `json:"code,omitempty"`
	Error string `json:"error,omitempty"`
}

type TornErrorExt struct {
	Text   string
	Remove bool // Remove from pool
	Delay  bool // Delay next request
}

func (e *TornErrorExt) Error() string {
	return e.Text
}

func (ter *TornErrorResponse) GetError() *TornErrorExt {
	switch ter.Error.Code {
	case 0:
		return &TornErrorExt{"Unknown", false, true}
	case 1:
		return &TornErrorExt{"EmptyKey", true, false}
	case 2:
		return &TornErrorExt{"IncorrectKey", true, false}
	case 5:
		return &TornErrorExt{"TooManyRequests", false, true}
	case 8:
		return &TornErrorExt{"IpBlock", false, true}
	case 9:
		return &TornErrorExt{"ApiDisabled", false, true}
	case 10:
		return &TornErrorExt{"UserInFederalJail", false, true}
	case 11:
		return &TornErrorExt{"KeyChangeError", false, true}
	case 12:
		return &TornErrorExt{"KeyReadError", false, true}
	default:
		return &TornErrorExt{"UnmappedError", false, false}
	}
}

func (tc TornClient) GetUser(apiKey string) (*model.User, *TornErrorResponse, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.torn.com/user", nil)
	if err != nil {
		return nil, nil, err
	}

	q := req.URL.Query()
	q.Add("selections", UserSelections)
	q.Add("key", apiKey)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Accept", "application/json")

	resp, err := tc.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	if resp.Body != nil {
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				log.Printf("Unable to close Response body: %s\n", err)
			}
		}()
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	responseType, err := model.GetUserResponseType(body)
	if err != nil {
		return nil, nil, err
	}

	if *responseType == "Torn" || *responseType == "Kafka" {
		user := model.User{}
		err = json.Unmarshal(body, &user)
		if err != nil {
			return nil, nil, err
		}
		return &user, nil, nil
	} else if *responseType == "Error" {
		errorResponse := TornErrorResponse{}
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			return nil, nil, err
		}
		return nil, &errorResponse, nil
	} else {
		return nil, nil, errors.New("Unexpected response type: " + *responseType)
	}
}
