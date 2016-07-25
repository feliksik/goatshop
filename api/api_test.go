package api_test

import (
	"encoding/json"
	"fmt"
	"gitlab.com/nerdalize/yak/api"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

var (
	server     *httptest.Server
	reader     io.Reader //Ignore this for now
	baseUrl    string
	shop       *api.Shop
	jsonHeader = http.Header{} // set in init
)

func init() {
	shop = api.NewShop()

	server = httptest.NewServer(shop.GetHandler()) //Creating new server with the user handlers

	baseUrl = fmt.Sprintf("%s", server.URL) //Grab the address for the API endpoint
	fmt.Println("hosting on ", baseUrl)

	jsonHeader.Set("Content-Type", "application/json")

}

func loadBasicShop() (*http.Response, error) {

	herdXML := `<?xml version="1.0" encoding="UTF-8" ?>
<herd>
  <labyak name="Betty-1" age="4" sex="f" />
  <labyak name="Betty-2" age="8" sex="f" />
  <labyak name="Betty-3" age="9.5" sex="f" />
</herd>
`
	request, err := http.NewRequest(
		"POST",
		baseUrl+"/yak-shop/load",
		strings.NewReader(herdXML)) //Create request with JSON body

	res, err := http.DefaultClient.Do(request)
	return res, err
}

func TestLoadShop(t *testing.T) {
	res, err := loadBasicShop()
	if err != nil {
		t.Error(err) //Something is wrong while sending request
	}

	if res.StatusCode != 205 {
		t.Errorf("Expected: 205. Actual: %d", res.StatusCode) //Uh-oh this means our test failed
	}

	if nrYaks := len(shop.Herd.Yaks); nrYaks != 3 {
		t.Fatalf("Yaks expected: 3. Actual: %d", nrYaks) //Uh-oh this means our test failed
	}

	y1 := shop.Herd.Yaks[1]

	if age := y1.Age; age != 800 {
		t.Errorf("Age should be 800 days, but is %d", age)
	}
	if y1.Name != "Betty-2" {
		t.Errorf("Name should be 'Betty-2', but is '%s'", y1.Name)
	}
	if y1.Sex != "f" {
		t.Errorf("Sex should be 'f', but is '%s'", y1.Sex)
	}

	if shop.Stock.Skins != 0 {
		t.Errorf("Stock should have 0 skins, but has %d", shop.Stock.Skins)
	}
	if shop.Stock.Milk != 0 {
		t.Errorf("Stock should have 0 skins, but has %d", shop.Stock.Milk)
	}

	if shop.CurrentDay != 0 {
		t.Errorf("Day should be reset to 0, but is %d", shop.CurrentDay)
	}
}

var EPSILON float64 = 0.00000001

func floatEquals(a, b float64) bool {
	if (a-b) < EPSILON && (b-a) < EPSILON {
		return true
	}
	return false
}

/********************************************************
 * START OF SOMETHING THAT COULD BE A LIBRARY
 ********************************************************/

type ResponseSpec struct {
	Body       string
	Header     http.Header
	StatusCode int
}

/*
	create an uncheck HTTP request, that is, do not check url valicity
*/
func HTTPRequest(method string, url string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	return req
}

/* return true iff s1 and s2 are equal json; return false otherwise.
   If one of the strings is not valid json, return (false, error)
*/
func AreEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 1 ('%s')", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 2 ('%s')", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}

/* get body of http response */
func GetBody(res *http.Response) (string, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	return string(body), nil
}

/* compare two strings as json. If they are not equal json, error with errmsg */
func checkEqualJSON(expected, received string) error {
	eq, err := AreEqualJSON(expected, received)

	if err != nil {
		return err
	}
	if !eq {
		return fmt.Errorf("---- received json: %s; \n---- expected json: %s", received, expected)
	}

	return nil // all good

}

func testJsonRequest(req *http.Request, responseSpec ResponseSpec) error {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if responseSpec.StatusCode != 0 {
		if res.StatusCode != responseSpec.StatusCode {
			return fmt.Errorf("Expected status %d, but got %d", responseSpec.StatusCode, res.Status)
		}
	}

	/* check if all required headers are set correctly */
	for k, v := range responseSpec.Header {
		expectedV := v[0]
		actualV := res.Header.Get(k)
		if actualV != expectedV {
			return fmt.Errorf("Expected header '%s = %s', but got '%s = %s ' ", k, expectedV, k, actualV)
		}
	}

	if responseSpec.Body != "" {
		body, err := GetBody(res)
		if err != nil {
			return fmt.Errorf("Could not get response body: %s", err.Error())
		}
		/* see if the response is correct */
		if err := checkEqualJSON(responseSpec.Body, body); err != nil {
			return fmt.Errorf("Response body does not match: " + err.Error())

		}

	}

	return nil

}

/********************************************************
 * END OF SOMETHING THAT COULD BE A LIBRARY
 ********************************************************/

func TestStock(t *testing.T) {

	err := testJsonRequest(
		HTTPRequest("GET", baseUrl+"/yak-shop/stock/13", nil),
		ResponseSpec{
			StatusCode: 200,
			Body: `{
				"milk" : 1104.48,
				"skins" : 3
			}`,
			Header: jsonHeader,
		},
	)
	if err != nil {
		t.Error(err)
	}

	err = testJsonRequest(
		HTTPRequest("GET", baseUrl+"/yak-shop/stock/14", nil),
		ResponseSpec{
			Body: `{
				"milk" : 1188.81,
				"skins" : 4
				}`,
			StatusCode: 200,
			Header:     jsonHeader,
		},
	)

	if err != nil {
		t.Error(err)
	}

}

func TestOrderAvailable(t *testing.T) {
	TestLoadShop(t)

	/*
	 *	test fully available order
	 */

	/* see if the response is correct */
	err := testJsonRequest(
		HTTPRequest("POST",
			baseUrl+fmt.Sprintf("/yak-shop/order/14"),
			strings.NewReader(`{
				"customer" : "Medvedev",
				"order" : {
					"milk" : 1100,
					"skins" : 3
				}
			}`),
		),
		ResponseSpec{
			StatusCode: http.StatusCreated,
			Body:       `{ "milk" : 1100.0, "skins" : 3 }`,
			Header:     jsonHeader,
		},
	)

	if err != nil {
		t.Error(err)
	}

	/* see if the stock is less now */

	err = testJsonRequest(
		HTTPRequest("GET",
			baseUrl+fmt.Sprintf("/yak-shop/stock/14"),
			nil,
		),
		ResponseSpec{
			StatusCode: http.StatusOK,
			Body:       `{ "milk" : 88.81, "skins" : 1 }`,
			Header:     jsonHeader,
		},
	)

	if err != nil {
		t.Error(err)
	}

}

func TestOrder_PartiallyAvailable(t *testing.T) {
	TestLoadShop(t)

	/*
	 *	test partially  available order: there isn't enough milk so it isn't shipped
	 */

	err := testJsonRequest(
		HTTPRequest(
			"POST",
			baseUrl+fmt.Sprintf("/yak-shop/order/14"),
			strings.NewReader(`{
				"customer" : "Medvedev",
				"order" : {
					"milk" : 1200,
					"skins" : 3
				}
			}`),
		),
		ResponseSpec{
			StatusCode: http.StatusPartialContent,
			Body:       `{ "skins" : 3 }`,
			Header:     jsonHeader,
		},
	)

	if err != nil {
		t.Error(err)
	}

	err = testJsonRequest(
		HTTPRequest("GET", baseUrl+fmt.Sprintf("/yak-shop/stock/14"), nil),
		ResponseSpec{
			StatusCode: http.StatusOK,
			Body:       `{ "milk" : 1188.81, "skins" : 1 }`,
			Header:     jsonHeader,
		},
	)
	if err != nil {
		t.Error(err)
	}
}

func TestOrder_Unavailable(t *testing.T) {
	TestLoadShop(t)

	/*
	 *	test partially  available order: there isn't enough milk so it isn't shipped
	 */

	err := testJsonRequest(
		HTTPRequest(
			"POST",
			baseUrl+fmt.Sprintf("/yak-shop/order/14"),
			strings.NewReader(`{
				"customer" : "Medvedev",
				"order" : {
					"milk" : 1200,
					"skins" : 31
				}
			}`),
		),
		ResponseSpec{
			StatusCode: http.StatusNotFound,
			Body:       `{  }`,
			Header:     jsonHeader,
		},
	)

	if err != nil {
		t.Error(err)
	}

	/* see if the stock is the same */

	err = testJsonRequest(
		HTTPRequest("GET", baseUrl+"/yak-shop/stock/14", nil),
		ResponseSpec{
			StatusCode: 200,
			Body:       `{ "milk" : 1188.81, "skins" : 4 }`,
		},
	)

	if err != nil {
		t.Error(err)
	}
}

func TestHerd(t *testing.T) {
	TestLoadShop(t)

	err := testJsonRequest(
		HTTPRequest("GET", baseUrl+"/yak-shop/herd/13", nil),
		ResponseSpec{
			StatusCode: 200,
			Header:     jsonHeader,
			Body: `{
					"herd" : [
						{
							"name" : "Betty-1",
							"age" : 4.13,
							"age-last-shaved" : 4.0
						},
						{
							"name" : "Betty-2",
							"age" : 8.13,
							"age-last-shaved" : 8.0
						},
						{
							"name" : "Betty-3",
							"age" : 9.63,
							"age-last-shaved" : 9.5
						}
					]
				}`,
		},
	)

	if err != nil {
		t.Error(err)
	}

}
