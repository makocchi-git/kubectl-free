// Package constants has various definitions
package constants

const (

	// UnitBytes is bytes for "Bytes" output
	UnitBytes = 1

	//
	// International System of Units (SI prefix)
	//

	// UnitKiloBytes is bytes for "KiBytes" output
	UnitKiloBytes = UnitBytes * 1000

	// UnitMegaBytes is bytes for "MiBytes" output
	UnitMegaBytes = UnitKiloBytes * 1000

	// UnitGigaBytes is bytes for "GiBytes" output
	UnitGigaBytes = UnitMegaBytes * 1000

	// UnitBytesStr is unit string for bytes
	UnitBytesStr = "B"

	// UnitKiloBytesStr is unit string for kilobytes
	UnitKiloBytesStr = "K"

	// UnitMegaBytesStr is unit string for megabytes
	UnitMegaBytesStr = "M"

	// UnitGigaBytesStr is unit string for gigabytes
	UnitGigaBytesStr = "G"

	//
	// Binary prefix
	//

	// UnitKibiBytes is bytes for "KiBytes" output
	UnitKibiBytes = UnitBytes * 1024

	// UnitMibiBytes is bytes for "MiBytes" output
	UnitMibiBytes = UnitKiloBytes * 1024

	// UnitGibiBytes is bytes for "GiBytes" output
	UnitGibiBytes = UnitMegaBytes * 1024

	// UnitKibiBytesStr is unit string for kilobytes
	UnitKibiBytesStr = "Ki"

	// UnitMibiBytesStr is unit string for megabytes
	UnitMibiBytesStr = "Mi"

	// UnitGibiBytesStr is unit string for gigabytes
	UnitGibiBytesStr = "Gi"

	//
	// Emoji
	//

	// EmojiReady is smile emoji for node status
	EmojiReady = "😃"

	// EmojiNotReady is crying emoji for node status
	EmojiNotReady = "😭"

	// EmojiPodRunning is running emoji for pod status
	EmojiPodRunning = "✅"

	// EmojiPodSucceeded is succeeded emoji for pod status
	EmojiPodSucceeded = "⭕"

	// EmojiPodPending is pending emoji for pod status
	EmojiPodPending = "🚫"

	// EmojiPodFailed is faled emoji for pod status
	EmojiPodFailed = "❌"

	// EmojiPodUnknown is unknown emoji for pod status
	EmojiPodUnknown = "❓"
)
