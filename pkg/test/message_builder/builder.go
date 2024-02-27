// Copyright 2023 Google LLC
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

// Package message_builder contains utilities to build messages easily.
// This is particularly useful for tests.
// To use: populate the Builder struct and call BuildMessage. Every invocation of BuildMessage
// creates a message with a different Message Control Id and a date newer than the previous
// invocation.
package message_builder

import (
	"flag"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/bitcrshr/simhospital/pkg/ir"
	"github.com/bitcrshr/simhospital/pkg/message"
	"github.com/pkg/errors"
)

var (
	// AdtA01 represents an Adt^A01 HL7 message.
	AdtA01 = &message.Type{MessageType: "ADT", TriggerEvent: "A01"}
	// AdtA03 represents an Adt^A03 HL7 message.
	AdtA03 = &message.Type{MessageType: "ADT", TriggerEvent: "A03"}
	// AdtA04 represents an Adt^A04 HL7 message.
	AdtA04 = &message.Type{MessageType: "ADT", TriggerEvent: "A04"}
	// AdtA05 represents an Adt^A05 HL7 message.
	AdtA05 = &message.Type{MessageType: "ADT", TriggerEvent: "A05"}
	// AdtA09 represents an Adt^A09 HL7 message.
	AdtA09 = &message.Type{MessageType: "ADT", TriggerEvent: "A09"}
	// AdtA10 represents an Adt^A10 HL7 message.
	AdtA10 = &message.Type{MessageType: "ADT", TriggerEvent: "A10"}
	// AdtA13 represents an Adt^A13 HL7 message.
	AdtA13 = &message.Type{MessageType: "ADT", TriggerEvent: "A13"}
	// AdtA17 represents an Adt^A17 HL7 message.
	AdtA17 = &message.Type{MessageType: "ADT", TriggerEvent: "A17"}
	// AdtA23 represents an Adt^A23 HL7 message.
	AdtA23 = &message.Type{MessageType: "ADT", TriggerEvent: "A23"}
	// AdtA31 represents an Adt^A31 HL7 message.
	AdtA31 = &message.Type{MessageType: "ADT", TriggerEvent: "A31"}
	// AdtA34 represents an Adt^A34 HL7 message.
	AdtA34 = &message.Type{MessageType: "ADT", TriggerEvent: "A34"}
	// AdtA40 represents an Adt^A40 HL7 message.
	AdtA40 = &message.Type{MessageType: "ADT", TriggerEvent: "A40"}
	// OrmO01 represents an ORM^O01 HL7 message.
	OrmO01 = &message.Type{MessageType: "ORM", TriggerEvent: "O01"}
	// OruR01 represents an ORU^R01 HL7 message.
	OruR01 = &message.Type{MessageType: "ORU", TriggerEvent: "R01"}
	// OruR03 represents an ORU^R03 HL7 message.
	OruR03 = &message.Type{MessageType: "ORU", TriggerEvent: "R03"}
	// OruR32 represents an ORU^R32 HL7 message.
	OruR32 = &message.Type{MessageType: "ORU", TriggerEvent: "R32"}
	// OrrO02 represents an ORR^O02 HL7 message.
	OrrO02 = &message.Type{MessageType: "ORR", TriggerEvent: "O02"}

	supportedMessages = []*message.Type{
		AdtA01,
		AdtA03,
		AdtA04,
		AdtA05,
		AdtA09,
		AdtA10,
		AdtA13,
		AdtA17,
		AdtA23,
		AdtA31,
		AdtA34,
		AdtA40,
		OrmO01,
		OruR01,
		OruR03,
		OruR32,
		OrrO02,
	}

	timeIncrements = flag.Duration("message_builder_time_increments", time.Minute,
		"The time difference between messages generated by the Builder in every invocation of BuildMessage")
)

type Builder struct {
	currentControlID    uint64
	currentDate         time.Time
	MessageType         *message.Type
	HeaderInfo          *message.HeaderInfo
	PatientInfo         *ir.PatientInfo
	OtherPatientInfo    *ir.PatientInfo
	order               *ir.Order
	UseOtherPatientInfo bool // Set to true if you want the data from the second, different, patient to be used.
	// Replace is a map of string to string replacements in the resulting message.
	Replace map[string]string
}

func NewBuilderForTests() Builder {
	return NewBuilderWithTime(time.Date(2017, 07, 04, 10, 0, 0, 0, time.UTC))
}

func NewBuilderWithTime(t time.Time) Builder {
	return Builder{
		currentDate: t,
		MessageType: &message.Type{
			// These need to be set by the caller, otherwise the message will be invalid. This is intentional.
			MessageType:  "",
			TriggerEvent: "",
		},
		HeaderInfo: &message.HeaderInfo{
			SendingApplication:   "CERNER",
			SendingFacility:      "FACILITY",
			ReceivingApplication: "SIMHOSP",
			ReceivingFacility:    "OS",
		},
		PatientInfo: &ir.PatientInfo{
			Person: &ir.Person{
				MRN:       "12345",
				NHS:       "4444232133",
				Surname:   "SMITH",
				FirstName: "DAVE",
				Birth:     ir.NewValidTime(time.Date(1984, 7, 4, 12, 35, 18, 0, time.UTC)),
				Address: &ir.Address{
					FirstLine:  "6 Pancras Square",
					SecondLine: "",
					City:       "London",
					PostalCode: "N1C 4AG",
					Country:    "GBR",
					Type:       "HOME",
				},
				Gender: "M",
			},
			VisitID: 12345,
			AttendingDoctor: &ir.Doctor{
				ID:        "111222333444",
				Surname:   "Jensen",
				FirstName: "Alan",
				Prefix:    "Dr",
			},
			Location: &ir.PatientLocation{
				Poc:          "2 West",
				Room:         "Bay01",
				Bed:          "10",
				Facility:     "FACILITY",
				LocationType: "BED",
				Building:     "BUILDING",
				Floor:        "Floor 1",
			},
			PrimaryFacility: &ir.PrimaryFacility{
				Organization: "FAMILY PRACTICE",
				ID:           "10001000",
			},
		},
		OtherPatientInfo: &ir.PatientInfo{
			Person: &ir.Person{
				MRN:       "12346",
				NHS:       "3315554242",
				Surname:   "KELLY",
				FirstName: "SMITH",
				Address: &ir.Address{
					FirstLine:  "6 Pancras Square",
					SecondLine: "20 Bull Lane",
					City:       "London",
					PostalCode: "N1C 4AG",
					Country:    "GBR",
					Type:       "HOME",
				},
				Gender: "F",
			},
			VisitID: 12346,
			AttendingDoctor: &ir.Doctor{
				ID:        "212155551010",
				Surname:   "Leach",
				FirstName: "Lorene",
				Prefix:    "Dr",
			},
			Location: &ir.PatientLocation{
				Poc:          "2 West",
				Room:         "Bay01",
				Bed:          "2",
				Facility:     "FACILITY",
				LocationType: "BED",
				Building:     "BUILDING",
				Floor:        "Floor 1",
			},
		},
		UseOtherPatientInfo: false,
	}
}

// populateOrder populates the builder with an order.
func (h *Builder) populateOrder() *ir.Order {
	if h.order == nil {
		h.order = &ir.Order{
			OrderProfile: &ir.CodedElement{
				ID:           "lpdc-3969",
				Text:         "UREA AND ELECTROLYTES",
				CodingSystem: "WinPath",
			},
			Placer:                fmt.Sprintf("%d", rand.Int()),
			Filler:                fmt.Sprintf("%d", rand.Int()),
			OrderDateTime:         ir.NewValidTime(h.currentDate.Add(-1 * time.Hour)),
			CollectedDateTime:     ir.NewValidTime(h.currentDate.Add(-30 * time.Minute)),
			ReceivedInLabDateTime: ir.NewValidTime(h.currentDate.Add(-20 * time.Minute)),
			ReportedDateTime:      ir.NewValidTime(h.currentDate.Add(-10 * time.Minute)),
			OrderingProvider: &ir.Doctor{
				ID:        "212155551010",
				Surname:   "Leach",
				FirstName: "Lorene",
				Prefix:    "Dr",
			},
			OrderControl:  "RE",
			OrderStatus:   "IP",
			ResultsStatus: "C",
			Results:       h.result(),
		}
	}
	return h.order
}

// result populates the builder with a result.
func (h *Builder) result() []*ir.Result {
	return []*ir.Result{
		{
			TestName: &ir.CodedElement{
				ID:           "lpdc-2011",
				Text:         "Creatinine",
				CodingSystem: "WinPath",
			},
			Value:               "700",
			Unit:                "UML",
			ValueType:           "NM",
			Range:               "39.00 - 308.00",
			Status:              "F",
			AbnormalFlag:        "HIGH",
			Notes:               []string{"Note1", "Note2"},
			ObservationDateTime: ir.NewValidTime(h.currentDate.Add(-30 * time.Minute)),
		},
	}
}

// BuildMessage builds a HL7 message using the data contained in the Builder. It returns the HL7
// message as a string, with segments separated by \r.
// Every invocation of BuildMessage generates a message with a different Message Control ID and a
// date after the previous generated message, so that every resulting message is treated as unique.
// BuildMessage fails if the message type cannot be identified or is not correct.
func (h *Builder) BuildMessage(t *testing.T) string {
	t.Helper()
	msg, err := h.buildMessage()
	if err != nil {
		t.Fatalf("BuildMessage() failed with %v", err)
	}

	return h.applyReplaces(msg)
}

func (h *Builder) applyReplaces(str string) string {
	for k, v := range h.Replace {
		r := regexp.MustCompile(k)
		str = r.ReplaceAllString(str, v)
	}
	return str
}

func (h *Builder) buildMessage() (string, error) {
	h.incrementDate()
	t := h.currentDate
	h.HeaderInfo.MessageControlID = h.nextMessageControlID()
	var patientInfo *ir.PatientInfo
	var otherPatientInfo *ir.PatientInfo
	if h.UseOtherPatientInfo {
		patientInfo = h.OtherPatientInfo
		otherPatientInfo = h.PatientInfo
	} else {
		patientInfo = h.PatientInfo
		otherPatientInfo = h.OtherPatientInfo
	}
	switch *h.MessageType {
	case *AdtA01:
		msg, err := message.BuildAdmissionADTA01(h.HeaderInfo, patientInfo, t, t)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ADT^A01 message")
		}
		return msg.Message, nil
	case *AdtA03:
		msg, err := message.BuildDischargeADTA03(h.HeaderInfo, patientInfo, t, t)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ADT^A03 message")
		}
		return msg.Message, nil
	case *AdtA04:
		msg, err := message.BuildRegistrationADTA04(h.HeaderInfo, patientInfo, t, t)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ADT^A04 message")
		}
		return msg.Message, nil
	case *AdtA05:
		msg, err := message.BuildPreAdmitADTA05(h.HeaderInfo, patientInfo, t, t)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ADT^A05 message")
		}
		return msg.Message, nil
	case *AdtA09:
		msg, err := message.BuildTrackDepartureADTA09(h.HeaderInfo, patientInfo, t, t)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ADT^A09 message")
		}
		return msg.Message, nil
	case *AdtA10:
		msg, err := message.BuildTrackArrivalADTA10(h.HeaderInfo, patientInfo, t, t)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ADT^A10 message")
		}
		return msg.Message, nil
	case *AdtA13:
		msg, err := message.BuildCancelDischargeADTA13(h.HeaderInfo, patientInfo, t, t)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ADT^A13 message")
		}
		return msg.Message, nil
	case *AdtA17:
		msg, err := message.BuildBedSwapADTA17(h.HeaderInfo, patientInfo, t, t, otherPatientInfo)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ADT^A17 message")
		}
		return msg.Message, nil
	case *AdtA23:
		msg, err := message.BuildDeleteVisitADTA23(h.HeaderInfo, patientInfo, t, t)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ADT^A23 message")
		}
		return msg.Message, nil
	case *AdtA31:
		msg, err := message.BuildUpdatePersonADTA31(h.HeaderInfo, patientInfo, t, t)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ADT^A31 message")
		}
		return msg.Message, nil
	case *AdtA34:
		msg, err := message.BuildMergeADTA34(h.HeaderInfo, patientInfo, t, t, otherPatientInfo.Person.MRN)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ADT^A34 message")
		}
		return msg.Message, nil
	case *AdtA40:
		msg, err := message.BuildMergeADTA40(h.HeaderInfo, patientInfo, t, t, []string{otherPatientInfo.Person.MRN})
		if err != nil {
			return "", errors.Wrap(err, "cannot build ADT^A40 message")
		}
		return msg.Message, nil
	case *OrmO01:
		msg, err := message.BuildOrderORMO01(h.HeaderInfo, patientInfo, h.populateOrder(), t)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ORM^O01 message")
		}
		return msg.Message, nil
	case *OruR01:
		msg, err := message.BuildResultORUR01(h.HeaderInfo, patientInfo, h.populateOrder(), t)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ORU^R01 message")
		}
		return msg.Message, nil
	case *OrrO02:
		msg, err := message.BuildPathologyORRO02(h.HeaderInfo, patientInfo, h.populateOrder(), t)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ORR^O02 message")
		}
		return msg.Message, nil
	case *OruR03:
		msg, err := message.BuildResultORUR03(h.HeaderInfo, patientInfo, h.populateOrder(), t)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ORU^R03 message")
		}
		return msg.Message, nil
	case *OruR32:
		msg, err := message.BuildResultORUR32(h.HeaderInfo, patientInfo, h.populateOrder(), t)
		if err != nil {
			return "", errors.Wrap(err, "cannot build ORU^R32 message")
		}
		return msg.Message, nil
	default:
		return "", fmt.Errorf("unimplemented mapping: %v^%v", h.MessageType.MessageType, h.MessageType.TriggerEvent)
	}
}

func (h *Builder) nextMessageControlID() string {
	current := h.currentControlID
	h.currentControlID++
	return strconv.FormatUint(current, 10)
}

func (h *Builder) incrementDate() {
	h.currentDate = h.currentDate.Add(*timeIncrements)
}

// CurrentTime populates the builder with the current time.
func (h *Builder) CurrentTime() ir.NullTime {
	return ir.NewValidTime(h.currentDate)
}
