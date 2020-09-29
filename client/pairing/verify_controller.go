package pairing

import (
	"bytes"
	"fmt"
	"io"

	"github.com/brutella/hc/crypto"
	"github.com/brutella/hc/crypto/chacha20poly1305"
	"github.com/brutella/hc/db"
	"github.com/brutella/hc/hap"
	"github.com/brutella/hc/hap/pair"
	"github.com/brutella/hc/util"
)

// VerifyClientController verifies the stored accessory public key and negotiates a shared secret
// which is used encrypt the upcoming session.
//
// Verification fails when the accessory is not known, the public key for the accessory was not found,
// or the packet's seal could not be verified.
//
// This is copied from
// https://github.com/brutella/hc/blob/d2805b655c7894e335b61e0a0b1e9680792b4ab1/hap/pair/verify_client_controller.go
// but adds the SessionCryptographer() method to allow extracting the negotiated shared secret.
type VerifyClientController struct {
	client   hap.Device
	database db.Database
	session  *pair.VerifySession
}

// NewVerifyClientController returns a new verify client controller.
func NewVerifyClientController(client hap.Device, database db.Database) *VerifyClientController {
	controller := VerifyClientController{
		client:   client,
		database: database,
		session:  pair.NewVerifySession(),
	}

	return &controller
}

// Handle processes a container to verify if an accessory is paired correctly.
func (v *VerifyClientController) Handle(in util.Container) (util.Container, error) {
	var out util.Container
	var err error

	method := pair.PairMethodType(in.GetByte(pair.TagPairingMethod))

	// It is valid that method is not sent
	// If method is sent then it must be 0x00
	if method != pair.PairingMethodDefault {
		return nil, fmt.Errorf("invalid pairing method: %v", method)
	}

	seq := pair.VerifyStepType(in.GetByte(pair.TagSequence))
	switch seq {
	case pair.VerifyStepStartResponse:
		out, err = v.handlePairStepVerifyResponse(in)
	case pair.VerifyStepFinishResponse:
		out, err = v.handlePairVerifyStepFinishResponse(in)
	default:
		return nil, fmt.Errorf("invalid verify step: %v", seq)
	}

	return out, err
}

// InitialKeyVerifyRequest returns the first request the client sends to an accessory to start the paring verifcation process.
// The request contains the client public key and sequence set to VerifyStepStartRequest.
func (v *VerifyClientController) InitialKeyVerifyRequest() io.Reader {
	out := util.NewTLV8Container()
	out.SetByte(pair.TagPairingMethod, 0)
	out.SetByte(pair.TagSequence, pair.VerifyStepStartRequest.Byte())
	out.SetBytes(pair.TagPublicKey, v.session.PublicKey[:])

	return out.BytesBuffer()
}

// Server -> Client
// - B: server public key
// - encrypted message
//      - username
//      - signature: from server session public key, server name, client session public key
//
// Client -> Server
// - encrypted message
//      - username
//      - signature: from client session public key, server name, server session public key,
func (v *VerifyClientController) handlePairStepVerifyResponse(in util.Container) (util.Container, error) {
	serverPublicKey := in.GetBytes(pair.TagPublicKey)
	if len(serverPublicKey) != 32 {
		return nil, fmt.Errorf("Invalid server public key size %d", len(serverPublicKey))
	}

	var otherPublicKey [32]byte
	copy(otherPublicKey[:], serverPublicKey)
	v.session.GenerateSharedKeyWithOtherPublicKey(otherPublicKey)
	if err := v.session.SetupEncryptionKey([]byte("Pair-Verify-Encrypt-Salt"), []byte("Pair-Verify-Encrypt-Info")); err != nil {
		return nil, fmt.Errorf("session SetupEncryptionKey: %v", err)
	}

	// Decrypt
	data := in.GetBytes(pair.TagEncryptedData)
	message := data[:(len(data) - 16)]
	var mac [16]byte
	copy(mac[:], data[len(message):]) // 16 byte (MAC)

	decryptedBytes, err := chacha20poly1305.DecryptAndVerify(v.session.EncryptionKey[:], []byte("PV-Msg02"), message, mac, nil)
	if err != nil {
		return nil, err
	}

	decryptedIn, err := util.NewTLV8ContainerFromReader(bytes.NewBuffer(decryptedBytes))
	if err != nil {
		return nil, err
	}

	username := decryptedIn.GetString(pair.TagUsername)
	signature := decryptedIn.GetBytes(pair.TagSignature)

	// Validate signature
	var material []byte
	material = append(material, v.session.OtherPublicKey[:]...)
	material = append(material, username...)
	material = append(material, v.session.PublicKey[:]...)

	var entity db.Entity
	if entity, err = v.database.EntityWithName(username); err != nil {
		return nil, fmt.Errorf("Server %s is unknown", username)
	}

	if len(entity.PublicKey) == 0 {
		return nil, fmt.Errorf("No LTPK available for client %s", username)
	}

	if !crypto.ValidateED25519Signature(entity.PublicKey, material, signature) {
		return nil, fmt.Errorf("Could not validate signature")
	}

	out := util.NewTLV8Container()
	out.SetByte(pair.TagSequence, pair.VerifyStepFinishRequest.Byte())

	encryptedOut := util.NewTLV8Container()
	encryptedOut.SetString(pair.TagUsername, v.client.Name())

	material = make([]byte, 0)
	material = append(material, v.session.PublicKey[:]...)
	material = append(material, v.client.Name()...)
	material = append(material, v.session.OtherPublicKey[:]...)

	signature, err = crypto.ED25519Signature(v.client.PrivateKey(), material)
	if err != nil {
		return nil, err
	}

	encryptedOut.SetBytes(pair.TagSignature, signature)

	encryptedBytes, mac, err := chacha20poly1305.EncryptAndSeal(v.session.EncryptionKey[:], []byte("PV-Msg03"), encryptedOut.BytesBuffer().Bytes(), nil)
	if err != nil {
		return nil, fmt.Errorf("chacha20poly1305.EncryptAndSeal: %v", err)
	}

	out.SetBytes(pair.TagEncryptedData, append(encryptedBytes, mac[:]...))

	return out, nil
}

// Server -> Client
// - only error ocde (optional)
func (v *VerifyClientController) handlePairVerifyStepFinishResponse(in util.Container) (util.Container, error) {
	code := in.GetByte(pair.TagErrCode)
	if code != byte(pair.ErrCodeNo) {
		return nil, fmt.Errorf("verify finish error: %d", code)
	}

	return nil, nil
}

// SessionCryptographer returns the Cryptographer negotiated during the verification process.
// This cryptographer can be used for further communication with the hap accessory.
func (v *VerifyClientController) SessionCryptographer() (crypto.Cryptographer, error) {
	c, err := crypto.NewSecureClientSessionFromSharedKey(v.session.SharedKey)
	if err != nil {
		return nil, fmt.Errorf("crypto.NewSecureClientSessionFromSharedKey: %v", err)
	}

	return c, nil
}
