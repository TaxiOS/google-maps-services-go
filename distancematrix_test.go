// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package maps

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"
)

func TestDistanceMatrixSydPyrToPar(t *testing.T) {

	// Distance Matrix from Sydney And Pyrmont to Parramatta with most steps elided.
	response := `{
   "destination_addresses" : [ "Parramatta NSW, Australia" ],
   "origin_addresses" : [ "Sydney NSW, Australia", "Pyrmont NSW, Australia" ],
   "rows" : [
      {
         "elements" : [
            {
               "distance" : {
                  "text" : "23.8 km",
                  "value" : 23846
               },
               "duration" : {
                  "text" : "37 mins",
                  "value" : 2215
               },
               "status" : "OK"
            }
         ]
      },
      {
         "elements" : [
            {
               "distance" : {
                  "text" : "22.2 km",
                  "value" : 22242
               },
               "duration" : {
                  "text" : "34 mins",
                  "value" : 2058
               },
               "status" : "OK"
            }
         ]
      }
   ],
   "status" : "OK"
}`

	server := mockServer(200, response)
	defer server.Close()
	c, _ := NewClient(WithAPIKey(apiKey))
	c.baseURL = server.URL
	r := &DistanceMatrixRequest{
		Origins:      []string{"Sydney", "Pyrmont"},
		Destinations: []string{"Parramatta"},
	}

	resp, err := c.DistanceMatrix(context.Background(), r)

	if err != nil {
		t.Errorf("r.Get returned non nil error, was %+v", err)
	}

	correctResponse := &DistanceMatrixResponse{
		OriginAddresses:      []string{"Sydney NSW, Australia", "Pyrmont NSW, Australia"},
		DestinationAddresses: []string{"Parramatta NSW, Australia"},
		Rows: []DistanceMatrixElementsRow{
			DistanceMatrixElementsRow{
				Elements: []*DistanceMatrixElement{
					&DistanceMatrixElement{
						Status:   "OK",
						Duration: 2215000000000,
						Distance: Distance{HumanReadable: "23.8 km", Meters: 23846},
					},
				},
			},
			DistanceMatrixElementsRow{
				Elements: []*DistanceMatrixElement{
					&DistanceMatrixElement{
						Status:   "OK",
						Duration: 2058000000000,
						Distance: Distance{HumanReadable: "22.2 km", Meters: 22242},
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(resp, correctResponse) {
		t.Errorf("expected %+v, was %+v", correctResponse, resp)
	}
}

func TestDistanceMatrixMissingOrigins(t *testing.T) {
	c, _ := NewClient(WithAPIKey(apiKey))
	r := &DistanceMatrixRequest{
		Origins:      []string{},
		Destinations: []string{"Parramatta"},
	}

	if _, err := c.DistanceMatrix(context.Background(), r); err == nil {
		t.Errorf("Missing Origins should return error")
	}
}

func TestDistanceMatrixMissingDestinations(t *testing.T) {
	c, _ := NewClient(WithAPIKey(apiKey))
	r := &DistanceMatrixRequest{
		Origins:      []string{"Sydney", "Pyrmont"},
		Destinations: []string{},
	}

	if _, err := c.DistanceMatrix(context.Background(), r); err == nil {
		t.Errorf("Missing Destinations should return error")
	}
}

func TestDistanceMatrixDepartureAndArrivalTime(t *testing.T) {
	c, _ := NewClient(WithAPIKey(apiKey))
	r := &DistanceMatrixRequest{
		Origins:       []string{"Sydney", "Pyrmont"},
		Destinations:  []string{"Parramatta", "Perth"},
		DepartureTime: "now",
		ArrivalTime:   "4pm",
	}

	if _, err := c.DistanceMatrix(context.Background(), r); err == nil {
		t.Errorf("Having both Departure time and Arrival time should return error")
	}
}

func TestDistanceMatrixTravelModeTransit(t *testing.T) {
	c, _ := NewClient(WithAPIKey(apiKey))
	var transitModes []TransitMode
	transitModes = append(transitModes, TransitModeBus)
	r := &DistanceMatrixRequest{
		Origins:      []string{"Sydney"},
		Destinations: []string{"Parramatta"},
		TransitMode:  transitModes,
	}

	if _, err := c.DistanceMatrix(context.Background(), r); err == nil {
		t.Errorf("Declaring TransitMode without Mode=Transit should return error")
	}
}

func TestDistanceMatrixTransitRoutingPreference(t *testing.T) {
	c, _ := NewClient(WithAPIKey(apiKey))
	r := &DistanceMatrixRequest{
		Origins:                  []string{"Sydney"},
		Destinations:             []string{"Parramatta"},
		TransitRoutingPreference: TransitRoutingPreferenceFewerTransfers,
	}

	if _, err := c.DistanceMatrix(context.Background(), r); err == nil {
		t.Errorf("Declaring TransitRoutingPreference without Mode=TravelModeTransit should return error")
	}
}

func TestDistanceMatrixWithCancelledContext(t *testing.T) {
	c, _ := NewClient(WithAPIKey(apiKey))
	r := &DistanceMatrixRequest{
		Origins:      []string{"Sydney", "Pyrmont"},
		Destinations: []string{"Parramatta", "Perth"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.DistanceMatrix(ctx, r); err == nil {
		t.Errorf("Cancelled context should return non-nil err")
	}
}

func TestDistanceMatrixFailingServer(t *testing.T) {
	server := mockServer(500, `{"status" : "ERROR"}`)
	defer server.Close()
	c, _ := NewClient(WithAPIKey(apiKey))
	c.baseURL = server.URL
	r := &DistanceMatrixRequest{
		Origins:      []string{"Sydney", "Pyrmont"},
		Destinations: []string{},
	}

	if _, err := c.DistanceMatrix(context.Background(), r); err == nil {
		t.Errorf("Failing server should return error")
	}
}

func TestDistanceMatrixRequestURL(t *testing.T) {
	expectedQuery := "avoid=t%7Co%7Cl%7Cl%7Cs&departure_time=now&destinations=Perth%7CParramatta&key=AIzaNotReallyAnAPIKey&language=en&mode=transit&origins=Sydney%7CPyrmont&transit_mode=rail&transit_routing_preference=less_walking&units=imperial"

	server := mockServerForQuery(expectedQuery, 200, `{"status":"OK"}"`)
	defer server.s.Close()

	c, _ := NewClient(WithAPIKey(apiKey))
	c.baseURL = server.s.URL

	r := &DistanceMatrixRequest{
		Origins:                  []string{"Sydney", "Pyrmont"},
		Destinations:             []string{"Perth", "Parramatta"},
		Mode:                     TravelModeTransit,
		Language:                 "en",
		Avoid:                    AvoidTolls,
		Units:                    UnitsImperial,
		DepartureTime:            "now",
		TransitMode:              []TransitMode{TransitModeRail},
		TransitRoutingPreference: TransitRoutingPreferenceLessWalking,
	}

	_, err := c.DistanceMatrix(context.Background(), r)
	if err != nil {
		t.Errorf("Unexpected error in constructing request URL: %+v", err)
	}
	if server.successful != 1 {
		t.Errorf("Got URL(s) %v, want %s", server.failed, expectedQuery)
	}
}
