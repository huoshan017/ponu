package time

import "time"

const (
	defaultInterval            = 33 * time.Millisecond
	minInterval                = 5 * time.Millisecond
	defaultTimerRecvListLength = 2048
	defaultRemoveListLength    = 512
	defaultSendNum             = 1
	defaultSenderListLength    = 16
)

type Options struct {
	interval            time.Duration
	timerRecvListLength int32
	removeListLength    int32
	maxSenderNum        int32
	senderListLength    int32
}

type Option func(*Options)

func (options *Options) GetInterval() time.Duration {
	return options.interval
}

func (options *Options) SetInterval(interval time.Duration) {
	options.interval = interval
}

func (options *Options) GetTimerRecvListLength() int32 {
	return options.timerRecvListLength
}

func (options *Options) SetTimerRecvListLength(length int32) {
	options.timerRecvListLength = length
}

func (options *Options) GetRemoveListLength() int32 {
	return options.removeListLength
}

func (options *Options) SetRemoveListLength(length int32) {
	options.removeListLength = length
}

func (options *Options) GetMaxSenderNum() int32 {
	return options.maxSenderNum
}

func (options *Options) SetMaxSenderNum(num int32) {
	options.maxSenderNum = num
}

func (options *Options) GetSenderListLength() int32 {
	return options.senderListLength
}

func (options *Options) SetSenderListLength(length int32) {
	options.senderListLength = length
}

func WithInterval(interval time.Duration) Option {
	return func(options *Options) {
		options.interval = interval
	}
}

func WithTimerRecvListLength(length int32) Option {
	return func(options *Options) {
		options.timerRecvListLength = length
	}
}

func WithRemoveListLength(length int32) Option {
	return func(options *Options) {
		options.removeListLength = length
	}
}

func WithMaxSenderNum(num int32) Option {
	return func(options *Options) {
		options.maxSenderNum = num
	}
}

func WithSenderListLength(length int32) Option {
	return func(options *Options) {
		options.senderListLength = length
	}
}
