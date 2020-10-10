package flasher

type FlashStatus int

const (
	FlashStarted FlashStatus = iota
	FlashSuccess
	FlashFailed
)

type LockStatus int

const (
	Unlocked LockStatus = iota
	Locked
)

type Device struct {
	ID string
	Codename string
	FlashStatus FlashStatus
	LockStatus LockStatus
}