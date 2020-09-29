package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddRemovePairing(t *testing.T) {
	testServer, err := deviceServer()
	require.NoError(t, err, "deviceServer")
	defer testServer.Close()

	controller, err := NewRandomControllerConfig()
	require.NoError(t, err, "controller setup")

	ctx := context.Background()
	connectionConfig, err := setupDeviceServer(ctx, testServer, controller)
	require.NoError(t, err, "pair")

	addController, err := NewRandomControllerConfig()
	require.NoError(t, err, "addController setup")

	accClient := NewAccessoryClient(NewIPDialer(), controller, connectionConfig)
	defer accClient.Close()
	require.NoError(t, accClient.AddPairing(
		ctx,
		&AddPairingRequest{
			DeviceID:    addController.DeviceID,
			PublicKey:   addController.PublicKey,
			Permissions: 1,
		},
	))

	func() {
		addAccClient := NewAccessoryClient(NewIPDialer(), addController, connectionConfig)
		defer addAccClient.Close()
		accessories, err := addAccClient.Accessories(ctx)
		require.NoError(t, err, "additional controller connection")
		require.Equal(t, "Test", accessories[0].Info().Name.Value)
	}()

	require.NoError(t, accClient.RemovePairing(
		ctx,
		addController.DeviceID,
	))

	func() {
		addAccClient := NewAccessoryClient(NewIPDialer(), addController, connectionConfig)
		defer addAccClient.Close()
		accessories, err := addAccClient.Accessories(ctx)
		require.Error(t, err, "additional controller should not connect after pairing removal")
		require.Nil(t, accessories)
	}()
}

func TestRemoveSetupPairing(t *testing.T) {
	testServer, err := deviceServer()
	require.NoError(t, err, "deviceServer")
	defer testServer.Close()

	controller, err := NewRandomControllerConfig()
	require.NoError(t, err, "controller setup")

	ctx := context.Background()
	connectionConfig, err := setupDeviceServer(ctx, testServer, controller)
	require.NoError(t, err, "pair")

	func() {
		accClient := NewAccessoryClient(NewIPDialer(), controller, connectionConfig)
		defer accClient.Close()
		require.NoError(t, accClient.RemovePairing(
			ctx,
			controller.DeviceID,
		))
	}()

	func() {
		accClient := NewAccessoryClient(NewIPDialer(), controller, connectionConfig)
		defer accClient.Close()
		accessories, err := accClient.Accessories(ctx)
		require.Error(t, err, "controller should not connect after pairing removal")
		require.Nil(t, accessories)
	}()
}
