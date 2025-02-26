// Copyright 2020 Google LLC
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

// Package address contains functionality to generate addresses.
package address

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/bitcrshr/simhospital/pkg/config"
	"github.com/bitcrshr/simhospital/pkg/ir"
)

// Generator is a generator of addresses.
type Generator struct {
	Nouns   []string
	Address config.Address
}

// Random generates a random address. The address will be in one of the following formats with equal probabilities:
//  1. 1 line address:
//     222 XXX StreetSuffix
//     PostCode
//     City
//     Country
//  2. 2 lines address:
//     111 XXX House
//     XXX StreetSuffix
//     PostCode
//     City
//     Country
//
// Where:
// 222 is a random number between [1, 200]
// 111 is a random number between [1, 100]
// XXX is a random noun
// StreetSuffix is a street suffix, eg.: Road, Street, Place etc.
// PostCode is a random post code. If the data configuration file contains a list of postcodes, it
// is chosen randomly among them. Otherwise, it is generated based on the country.
// City is a random city.
func (g *Generator) Random() *ir.Address {
	a := &ir.Address{
		City:       g.city(),
		PostalCode: g.postcode(),
		Country:    g.Address.Country,
		Type:       "HOME",
	}

	if isUSA(g.Address.Country) || rand.Intn(2) == 0 {
		// 1 line address
		a.FirstLine = fmt.Sprintf("%d %s %s", rand.Intn(200)+1, strings.Title(g.noun()), g.street())
	} else {
		// 2 lines address
		a.FirstLine = fmt.Sprintf("%d %s House", rand.Intn(100)+1, strings.Title(g.noun()))
		a.SecondLine = fmt.Sprintf("%s %s", strings.Title(g.noun()), g.street())
	}
	return a
}

func (g *Generator) postcode() string {
	if len(g.Address.Postalcodes) > 0 {
		return random(g.Address.Postalcodes)
	}
	if isUSA(g.Address.Country) {
		return postcodeUS()
	}
	return postcodeUK()
}

func (g *Generator) city() string {
	return random(g.Address.Cities)
}

func (g *Generator) street() string {
	return random(g.Address.Streets)
}

func (g *Generator) noun() string {
	return random(g.Nouns)
}

// random returns a random item from the given slice.
func random(s []string) string {
	return s[rand.Intn(len(s))]
}

func isUSA(country string) bool {
	return country == "USA" || country == "US"
}
