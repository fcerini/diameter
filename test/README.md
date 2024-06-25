```go
package main

import (
	"fmt"
	"log"
	"net"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/dict"
)

// Configuration parameters
const (
	pcrfAddress  = "192.168.1.10:3868" // Replace with actual PCRF address
	realm         = "example.com"      // Replace with actual realm
	vendorID      = 10415               // Huawei vendor ID
	applicationID = 101                // Diameter Gy application ID
)

// QoS profile details
type QoSProfile struct {
	Name            string
	MaxBandwidthDL  uint32
	MaxBandwidthUL  uint32
	GuaranteedBitrateDL uint32
	GuaranteedBitrateUL uint32
}

// Function to create a CCR message for QoS update
func buildCCR(imsi string, qosProfile QoSProfile) (*diam.Message, error) {
	m := diam.NewRequest(diam.CreditControl, diam.TGppS6A, dict.Default)

	// Set mandatory AVPs
	m.NewAVP(avp.SessionID, avp.Mbit|avp.Vbit, 0, datatype.UTF8String(diam.NewSessionID(realm)))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(applicationID))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity(realm))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity(realm))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("")) // Let PCRF fill this

	// Set CCR specific AVPs
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(diam.CreditControlRequestUpdate))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("QoS"))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 1, datatype.Grouped{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(diam.EndUserSubscriptionIDTypeIMSI)),
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
		Name:            "HighPriority",
		MaxBandwidthDL:  10240, // 10 Mbps
		MaxBandwidthUL:  5120,  // 5 Mbps
		GuaranteedBitrateDL: 512, // 512 kbps
		GuaranteedBitrateUL: 256, // 256 kbps
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
```

**Explanation:**

1. **Import necessary libraries:** This includes `diam`, `avp`, `datatype`, and `dict` from the `github.com/fiorix/go-diameter` package for handling Diameter protocol functionalities.
2. **Configure connection parameters:** 
    - Replace `pcrfAddress` with the actual IP address and port of your Huawei PCRF.
    - Update `realm`, `vendorID`, and `applicationID` based on your specific environment.
3. **Define QoS profile:** Create a `QoSProfile` struct to hold the desired QoS parameters (Name, Bandwidth, Bitrate). You'll need to replace the placeholder values with the actual values for your use case.
4. **Implement `buildCCR` function:** This function takes the IMSI and QoS profile as input and constructs a CCR (Credit-Control-Request) message with the necessary AVPs (Attribute-Value Pairs).
    - Pay attention to the AVP codes for QoS parameters, which are specific to your Huawei PCRF. Replace the placeholder values (`26`, `1025`, etc.) with the correct ones from the Huawei documentation.
5. **Implement `setQoS` function:** This function takes the Diameter connection, IMSI, and QoS profile as input and performs the following:
    - Calls `buildCCR` to create the CCR message.
    - Sends the message to the PCRF using `conn.SendRequest`.
    - Handles the response (CCA - Credit-Control-Answer) from the PCRF:
        - Checks for the expected response code (Credit-Control-Answer).
        - Extracts and checks the Result-Code AVP for success or failure.
        - Prints an appropriate message based on the result.
6. **Implement `main` function:**
    - Establishes a TCP connection to the PCRF.
    - Defines the QoS profile to be applied.
    - Iterates through a list of IMSIs and calls `setQoS` for each IMSI to update its QoS profile.

**To use this code:**

1. Make sure you have the `github.com/fiorix/go-diameter` package installed: `go get github.com/fiorix/go-diameter`.
2. Update the configuration parameters and QoS profile with your specific values.
3. Consult your Huawei PCRF documentation to find the correct AVP codes for the QoS parameters and replace the placeholders in the `buildCCR` function.
4. Compile and run the code.

**Note:**

- This code provides a basic framework for integrating with a Huawei PCRF and setting QoS. You might need to adjust it further based on your specific requirements and the Diameter interface implemented by your PCRF.
- Always refer to your Huawei documentation for accurate information on AVP codes, message structures, and other details.