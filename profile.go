package kr

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
	"golang.org/x/crypto/ssh"
)

type Profile struct {
	SSHWirePublicKey []byte          `json:"public_key_wire"`
	Email            string          `json:"email"`
	PGPPublicKey     *[]byte         `json:"pgp_pk,omitempty"`
}

func (p Profile) AuthorizedKeyString() (authString string, err error) {
	authString, err = p.AuthorizedKeyStringWithoutEmail()
	if err != nil {
		return
	}
	authString += " " + strings.Replace(p.Email, " ", "", -1)
	return
}

func (p Profile) AuthorizedKeyStringWithoutEmail() (authString string, err error) {
	pk, err := p.SSHPublicKey()
	if err != nil {
		return
	}
	authString = pk.Type() + " " + base64.StdEncoding.EncodeToString(p.SSHWirePublicKey)
	return
}

func (p Profile) SSHPublicKey() (pk ssh.PublicKey, err error) {
	return ssh.ParsePublicKey(p.SSHWirePublicKey)
}

func (p Profile) RSAPublicKey() (pk *rsa.PublicKey, err error) {
	return SSHWireRSAPublicKeyToRSAPublicKey(p.SSHWirePublicKey)
}

func (p Profile) PublicKeyFingerprint() []byte {
	digest := sha256.Sum256(p.SSHWirePublicKey)
	return digest[:]
}

func (p Profile) Equal(other Profile) bool {
	return bytes.Equal(p.SSHWirePublicKey, other.SSHWirePublicKey) && p.Email == other.Email
}

var KRYPTONITE_ASCII_ARMOR_HEADERS = map[string]string{"Comment": "Created with Kryptonite"}
var KRYPTON_ASCII_ARMOR_HEADERS = map[string]string{"Comment": "Created with Krypton"}

func (p Profile) AsciiArmorPGPPublicKey() (s string, err error) {
	if p.PGPPublicKey == nil {
		err = fmt.Errorf("no pgp public key")
		return
	}
	output := &bytes.Buffer{}
	input, err := armor.Encode(output, "PGP PUBLIC KEY BLOCK", KRYPTON_ASCII_ARMOR_HEADERS)
	if err != nil {
		return
	}
	_, err = input.Write(*p.PGPPublicKey)
	if err != nil {
		return
	}
	err = input.Close()
	if err != nil {
		return
	}
	s = string(output.Bytes())
	return
}

func (p Profile) PGPPublicKeySHA1Fingerprint() (s string, err error) {
	if p.PGPPublicKey == nil {
		err = fmt.Errorf("no pgp public key")
		return
	}
	reader := bytes.NewReader(*p.PGPPublicKey)
	for {
		var pkt packet.Packet
		pkt, err = packet.Read(reader)
		if err != nil {
			break
		}
		switch pkt := pkt.(type) {
		case *packet.PublicKey:
			digest := pkt.Fingerprint[:]
			s = hex.EncodeToString(digest)
			return
		default:
			continue
		}
	}
	err = fmt.Errorf("no pgp public key packet found")
	return
}
