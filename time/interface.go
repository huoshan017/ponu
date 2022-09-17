package time

import "github.com/huoshan017/ponu/list"

type IResultSender interface {
	Send(int32, *list.List)
}
