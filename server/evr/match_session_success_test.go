package evr

import (
	"testing"
)

func TestEncoderFlagUnpacking(t *testing.T) {
	tests := []struct {
		flags                 uint64
		expectedEncryption    bool
		expectedMac           bool
		expectedDigestSize    int
		expectedIteration     int
		expectedMacSecretSize int
		expectedSecretSize    int
		expectedRandomSize    int
	}{
		{0, false, false, 0, 0, 0, 0, 0},                                // test with all flags disabled
		{1, true, false, 0, 0, 0, 0, 0},                                 // test with encryption flag enabled
		{2, false, true, 0, 0, 0, 0, 0},                                 // test with mac flag enabled
		{3, true, true, 0, 0, 0, 0, 0},                                  // test with both encryption and mac flags enabled
		{36037595259470083, true, true, 0x40, 0x00, 0x20, 0x20, 0x20},   // default client flags
		{36037595259469955, true, true, 0x20, 0x00, 0x20, 0x20, 0x20},   // default server flags
		{0x0080080080000102, false, true, 0x40, 0x00, 0x20, 0x20, 0x20}, // default client flags
		{0x0080080080000083, true, true, 0x20, 0x00, 0x20, 0x20, 0x20},  // default server flags
	}

	for _, tt := range tests {
		settings := PacketEncoderSettingsFromFlags(tt.flags)
		if settings.ToFlags() != tt.flags {
			t.Errorf("ToFlags() = %v, want %v", settings.ToFlags(), tt.flags)
		}
		if settings.EncryptionEnabled != tt.expectedEncryption {
			t.Errorf("EncryptionEnabled = %v, want %v", settings.EncryptionEnabled, tt.expectedEncryption)
		}
		if settings.MACEnabled != tt.expectedMac {
			t.Errorf("MacEnabled = %v, want %v", settings.MACEnabled, tt.expectedMac)
		}
		if settings.MACDigestSize != tt.expectedDigestSize {
			t.Errorf("MacDigestSize = %v, want %v", settings.MACDigestSize, tt.expectedDigestSize)
		}
		if settings.MACPBKDF2IterationCount != tt.expectedIteration {
			t.Errorf("MacPBKDF2IterationCount = %v, want %v", settings.MACPBKDF2IterationCount, tt.expectedIteration)
		}
		if settings.MACSecretSize != tt.expectedMacSecretSize {
			t.Errorf("MacSecretSize = %v, want %v", settings.MACSecretSize, tt.expectedMacSecretSize)
		}
		if settings.EncryptionSecretSize != tt.expectedSecretSize {
			t.Errorf("EncryptionSecretSize = %v, want %v", settings.EncryptionSecretSize, tt.expectedSecretSize)
		}
		if settings.StreamSecretSize != tt.expectedRandomSize {
			t.Errorf("RandomSecretSize = %v, want %v", settings.StreamSecretSize, tt.expectedRandomSize)
		}
	}
}

func TestEncoderFlagPacking(t *testing.T) {
	settings := DefaultServerEncoderSettings()
	flags := settings.ToFlags()
	parsed := PacketEncoderSettingsFromFlags(flags)

	if parsed.EncryptionEnabled != settings.EncryptionEnabled {
		t.Errorf("EncryptionEnabled mismatch: got %v, want %v", parsed.EncryptionEnabled, settings.EncryptionEnabled)
	}
	if parsed.MACEnabled != settings.MACEnabled {
		t.Errorf("MACEnabled mismatch: got %v, want %v", parsed.MACEnabled, settings.MACEnabled)
	}
	if parsed.MACDigestSize != settings.MACDigestSize {
		t.Errorf("MACDigestSize mismatch: got %v, want %v", parsed.MACDigestSize, settings.MACDigestSize)
	}
	if parsed.MACPBKDF2IterationCount != settings.MACPBKDF2IterationCount {
		t.Errorf("MACPBKDF2IterationCount mismatch: got %v, want %v", parsed.MACPBKDF2IterationCount, settings.MACPBKDF2IterationCount)
	}
	if parsed.MACSecretSize != settings.MACSecretSize {
		t.Errorf("MACSecretSize mismatch: got %v, want %v", parsed.MACSecretSize, settings.MACSecretSize)
	}
	if parsed.EncryptionSecretSize != settings.EncryptionSecretSize {
		t.Errorf("EncryptionSecretSize mismatch: got %v, want %v", parsed.EncryptionSecretSize, settings.EncryptionSecretSize)
	}
	if parsed.StreamSecretSize != settings.StreamSecretSize {
		t.Errorf("RandomSecretSize mismatch: got %v, want %v", parsed.StreamSecretSize, settings.StreamSecretSize)
	}
}
