// Copyright (c) 2018 Iori Mizutani
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package llrp

import (
	"bytes"
	"encoding/binary"
)

// LLRP header values, the header is composed of 2 bytes where the first
// is set to 0x40 or 1024, so 1025 would be 0x0401
const (
	GetReaderCapabilityHeader         = 1025      // type 1
	GetReaderConfigHeader             = 1026      // type 2
	SetReaderConfigHeader             = 1027      // type 3
	GetReaderCapabilityResponseHeader = 1035      // type 11
	GetReaderConfigResponseHeader     = 1036      // type 12
	SetReaderConfigResponseHeader     = 1037      // type 13
	AddRospecHeader                   = 1044      // type 20
	DeleteRospecHeader                = 1045      // type 21
	EnableRospecHeader                = 1048      // type 24
	AddRospecResponseHeader           = 1054      // type 30
	DeleteRospecResponseHeader        = 1055      // type 31
	EnableRospecResponseHeader        = 1058      // type 34
	DeleteAccessSpecHeader            = 1065      // type 41
	MsgGetSupportedVersion            = 1024 + 46 // MessageType(46)
	MsgGetSupportedVersionResponse    = 1024 + 56 //MessageType(56)
	DeleteAccessSpecResponseHeader    = 1075      // type 51
	ROAccessReportHeader              = 1085      // type 61
	KeepaliveHeader                   = 1086      // type 62
	ReaderEventNotificationHeader     = 1087      // type 63
	EventsAndReportsHeader            = 1088      // type 64
	KeepaliveAckHeader                = 1096      // type 72
	ImpinjEnableCutomMessageHeader    = 2047      // type 1023
)

// Pack the data into (partial) LLRP packet payload.
// TODO: count the data size and return resulting length ?
func Pack(data []interface{}) []byte {
	buf := new(bytes.Buffer)
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		if err != nil {
			panic(err)
		}
	}
	return buf.Bytes()
}

// ReadEvent is the struct to hold data on RFTags
type ReadEvent struct {
	ID []byte
	PC []byte
}

// UnmarshalROAccessReportBody extract ReadEvent from the message value in the ROAccessReport
func UnmarshalROAccessReportBody(roarBody []byte) []*ReadEvent {
	//defer timeTrack(time.Now(), fmt.Sprintf("unpacking %v bytes", len(roarBody)))
	res := []*ReadEvent{}

	// iterate through the parameters in roarBody
	for offset := 0; offset < len(roarBody); {
		parameterType := binary.BigEndian.Uint16(roarBody[offset : offset+2])

		switch parameterType {
		case uint16(240): // TagReportData
			offset += 4
		default:
			offset += int(binary.BigEndian.Uint16(roarBody[offset+2 : offset+4]))
			continue
		}

		// look into TagReportData
		// Now the offset is at the first parameter in the TRD
		var id, pc []byte
		if roarBody[offset] == 141 { // EPC-96
			id = roarBody[offset+1 : offset+13]
			offset += 13
			if roarBody[offset] == 140 { // C1G2-PC parameter
				pc = roarBody[offset+1 : offset+3]
				offset += 3
			}
		} else if binary.BigEndian.Uint16(roarBody[offset:offset+2]) == 241 { // EPCData
			epcDataLength := int(binary.BigEndian.Uint16(roarBody[offset+2 : offset+4])) // length
			//epcLengthBits := binary.BigEndian.Uint16(roarBody[offset+4 : offset+6])      // EPCLengthBits
			id = roarBody[offset+6 : offset+epcDataLength]
			offset += epcDataLength
			if roarBody[offset] == 140 { // C1G2-PC parameter
				pc = roarBody[offset+1 : offset+3]
				offset += 3
			}
		}
		// append the id and pc as an ReadEvent
		res = append(res, &ReadEvent{id, pc})
	}
	return res
}
