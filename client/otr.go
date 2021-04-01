package main

import (
	"crypto/rand"
	"fmt"
	"log"

	"golang.org/x/crypto/otr"
)

// Instead of implementing the OTR-protocol myself, I just use the one provided by go :)
var (
	privKey   otr.PrivateKey
	conv      otr.Conversation
	secChange otr.SecurityChange
)

func generatePrivateKey() {
	key := new(otr.PrivateKey)
	key.Generate(rand.Reader)
	bytes := key.Serialize(nil)

	parsedBytes, ok := privKey.Parse(bytes)
	if !ok {
		log.Printf("failed parsing private key")
	}

	if len(parsedBytes) > 0 {
		log.Printf("the key buffer is not empty after key ")
	}

	conv.PrivateKey = &privKey
	conv.FragmentSize = 1000
}

// the first return is the content and the second is the data to send back to the client.
// The last status is the value if the data is encrypted.
func parseContent(m *message) (string, [][]byte) {
	out, enc, _, mpeer, err := conv.Receive([]byte(m.Content))
	if err != nil {
		log.Printf("error receiving conversation bytes")
	}

	if len(out) > 0 {
		if !enc {
			return fmt.Sprintf("[NOT_ENCRYPTED] <%s> %s\n", m.Name, string(out)), mpeer
		} else {
			return fmt.Sprintf("[%s] <%s> %s\n", formatTimestamp(m.Timestamp),
				m.Name, string(out)), mpeer
		}
	}

	return "", mpeer
}
