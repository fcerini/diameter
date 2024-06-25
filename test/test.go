package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"github.com/fiorix/go-diameter/v4/diam"
	"github.com/fiorix/go-diameter/v4/diam/avp"
	"github.com/fiorix/go-diameter/v4/diam/datatype"
	"github.com/fiorix/go-diameter/v4/diam/dict"
)

// Configuration parameters
const (
	pcrfAddress   = "192.168.1.10:3868" // Replace with actual PCRF address
	realm         = "example.com"       // Replace with actual realm
	vendorID      = 10415               // Huawei vendor ID
	applicationID = 101                 // Diameter Gy application ID
)

// QoS profile details
type QoSProfile struct {
	Name                string
	MaxBandwidthDL      uint32
	MaxBandwidthUL      uint32
	GuaranteedBitrateDL uint32
	GuaranteedBitrateUL uint32
}

// Function to create a CCR message for QoS update
func buildCCR(imsi string, qosProfile QoSProfile) (*diam.Message, error) {
	m := diam.NewRequest(diam.CreditControl, diam.TGPP_S6A_APP_ID, dict.Default)

	// Set mandatory AVPs
	sid := "session;" + strconv.Itoa(int(rand.Uint32()))
	m.NewAVP(avp.SessionID, avp.Mbit|avp.Vbit, 0, datatype.UTF8String(sid))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(applicationID))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity(realm))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity(realm))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("")) // Let PCRF fill this

	// Set CCR specific AVPs
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(diam.CreditControl))
	//CreditControlRequestUpdate
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("QoS"))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 1, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(diam.)),
			//EndUserSubscriptionIDTypeIMSI
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String(imsi)),
		},
	})

	// Add QoS profile AVPs (replace with actual AVP codes for Huawei PCRF)
	m.NewAVP(avp.VendorSpecificApplicationName, avp.Mbit, 0, datatype.Grouped{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.VendorID, avp.Mbit, 0, datatype.Unsigned32(vendorID)),
			diam.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(applicationID)),
		},
	})
	m.NewAVP(avp.VendorSpecificData, avp.Mbit, 0, datatype.Grouped{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.VendorID, avp.Mbit, 0, datatype.Unsigned32(vendorID)),
			// Replace AVP codes below with actual values for Huawei PCRF
			diam.NewAVP(26, avp.Mbit, 0, datatype.UTF8String(qosProfile.Name)),
			diam.NewAVP(1025, avp.Mbit, 0, datatype.Unsigned32(qosProfile.MaxBandwidthDL)),
			diam.NewAVP(1026, avp.Mbit, 0, datatype.Unsigned32(qosProfile.MaxBandwidthUL)),
			diam.NewAVP(1027, avp.Mbit, 0, datatype.Unsigned32(qosProfile.GuaranteedBitrateDL)),
			diam.NewAVP(1028, avp.Mbit, 0, datatype.Unsigned32(qosProfile.GuaranteedBitrateUL)),
		},
	})

	return m, nil
}

// Function to set QoS for a single phone
func setQoS(conn *diam.Conn, imsi string, qosProfile QoSProfile) error {
	// Build the CCR message
	m, err := buildCCR(imsi, qosProfile)
	if err != nil {
		return fmt.Errorf("failed to build CCR message: %w", err)
	}

	// Send the message and wait for response
	ans, err := conn.SendRequest(m)
	if err != nil {
		return fmt.Errorf("failed to send CCR message: %w", err)
	}

	// Check for success
	if ans.Code != diam.CreditControlAnswer {
		return fmt.Errorf("unexpected response code: %d", ans.Code)
	}

	// Check Result-Code AVP
	resultCode := ans.Get(avp.ResultCode).Data.(datatype.Unsigned32)
	if resultCode != diam.ResultCodeSuccess {
		return fmt.Errorf("QoS update failed with result code: %d", resultCode)
	}

	fmt.Printf("QoS updated successfully for IMSI: %s\n", imsi)
	return nil
}

func main() {
	// Connect to PCRF
	conn, err := diam.Dial("tcp", pcrfAddress, nil, nil)
	if err != nil {
		log.Fatalf("failed to connect to PCRF: %v", err)
	}
	defer conn.Close()

	// Define QoS profile
	qosProfile := QoSProfile{
		Name:                "HighPriority",
		MaxBandwidthDL:      10240, // 10 Mbps
		MaxBandwidthUL:      5120,  // 5 Mbps
		GuaranteedBitrateDL: 512,   // 512 kbps
		GuaranteedBitrateUL: 256,   // 256 kbps
	}

	// List of IMSIs (replace with actual IMSIs)
	imsis := []string{
		"460010000000001",
		"460010000000002",
		"460010000000003",
	}

	// Set QoS for each IMSI
	for _, imsi := range imsis {
		err := setQoS(conn, imsi, qosProfile)
		if err != nil {
			fmt.Printf("Failed to set QoS for IMSI %s: %v\n", imsi, err)
		}
	}
}
